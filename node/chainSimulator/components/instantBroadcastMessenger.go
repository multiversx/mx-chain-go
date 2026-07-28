package components

import (
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/sharding"
)

type instantBroadcastMessenger struct {
	consensus.BroadcastMessenger
	shardCoordinator      sharding.Coordinator
	beforeBroadcastHeader func(header data.HeaderHandler)
	proposalDataHandler   func(header data.HeaderHandler, bodyBytes []byte, pkBytes []byte) error
	deliveryTracker       blockBodyDeliveryTracker
	headerDeliveryTracker blockHeaderDeliveryTracker
	mutPendingBodies      sync.Mutex
	pendingBodies         map[int64]*consensus.Message
}

// NewInstantBroadcastMessenger creates a new instance of type instantBroadcastMessenger
func NewInstantBroadcastMessenger(broadcastMessenger consensus.BroadcastMessenger, shardCoordinator sharding.Coordinator) (*instantBroadcastMessenger, error) {
	if check.IfNil(broadcastMessenger) {
		return nil, errors.ErrNilBroadcastMessenger
	}
	if check.IfNil(shardCoordinator) {
		return nil, errors.ErrNilShardCoordinator
	}

	return &instantBroadcastMessenger{
		BroadcastMessenger: broadcastMessenger,
		shardCoordinator:   shardCoordinator,
		pendingBodies:      make(map[int64]*consensus.Message),
	}, nil
}

func (messenger *instantBroadcastMessenger) setBeforeBroadcastHeader(handler func(header data.HeaderHandler)) {
	messenger.beforeBroadcastHeader = handler
}

func (messenger *instantBroadcastMessenger) setProposalDataHandler(
	handler func(header data.HeaderHandler, bodyBytes []byte, pkBytes []byte) error,
) {
	messenger.proposalDataHandler = handler
}

func (messenger *instantBroadcastMessenger) setBlockBodyDeliveryTracker(tracker blockBodyDeliveryTracker) {
	messenger.deliveryTracker = tracker
}

func (messenger *instantBroadcastMessenger) setBlockHeaderDeliveryTracker(tracker blockHeaderDeliveryTracker) {
	messenger.headerDeliveryTracker = tracker
}

// BroadcastConsensusMessage holds a separately sent v2 proposal body until BroadcastHeader.
// This lets an epoch-start header switch all validators to v2 before the body is delivered.
func (messenger *instantBroadcastMessenger) BroadcastConsensusMessage(message *consensus.Message) error {
	if message != nil && consensus.MessageType(message.MsgType) == bls.MtBlockBody {
		messenger.mutPendingBodies.Lock()
		messenger.pendingBodies[message.RoundIndex] = message
		messenger.mutPendingBodies.Unlock()

		return nil
	}

	return messenger.BroadcastMessenger.BroadcastConsensusMessage(message)
}

// BroadcastHeader synchronizes the consensus-version switch before delivering an epoch-start
// proposal. Production receivers switch asynchronously after observing this header; without this
// simulator barrier the v2 proposal can arrive while one validator still has v1 callbacks.
func (messenger *instantBroadcastMessenger) BroadcastHeader(header data.HeaderHandler, pkBytes []byte) error {
	var bodyMessage *consensus.Message
	if !check.IfNil(header) {
		messenger.mutPendingBodies.Lock()
		bodyMessage = messenger.pendingBodies[int64(header.GetRound())]
		delete(messenger.pendingBodies, int64(header.GetRound()))
		messenger.mutPendingBodies.Unlock()
	}

	if !check.IfNil(header) && header.IsStartOfEpochBlock() {
		if messenger.beforeBroadcastHeader != nil {
			messenger.beforeBroadcastHeader(header)
		}

		if messenger.proposalDataHandler != nil && bodyMessage != nil {
			err := messenger.proposalDataHandler(header, bodyMessage.Body, pkBytes)
			if err != nil {
				log.Warn("instantBroadcastMessenger.BroadcastHeader: prepare proposal data", "error", err)
			}
		}
	}

	if bodyMessage != nil {
		if messenger.deliveryTracker != nil {
			messenger.deliveryTracker.prepareBlockBodyDelivery(bodyMessage)
		}

		err := messenger.BroadcastMessenger.BroadcastConsensusMessage(bodyMessage)
		if err != nil {
			return err
		}

		if messenger.deliveryTracker != nil &&
			!messenger.deliveryTracker.waitBlockBodyDelivery(bodyMessage, simulatedConsensusMaxWait) {
			log.Debug(
				"instantBroadcastMessenger.BroadcastHeader: proposal body delivery timed out",
				"round", bodyMessage.RoundIndex,
			)
		}
	}

	if messenger.headerDeliveryTracker != nil {
		messenger.headerDeliveryTracker.prepareBlockHeaderDelivery(header, pkBytes)
	}

	err := messenger.BroadcastMessenger.BroadcastHeader(header, pkBytes)
	if err != nil {
		return err
	}

	if messenger.headerDeliveryTracker != nil &&
		!messenger.headerDeliveryTracker.waitBlockHeaderDelivery(header, simulatedConsensusMaxWait) {
		log.Debug(
			"instantBroadcastMessenger.BroadcastHeader: proposal header delivery timed out",
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
		)
	}

	return nil
}

// BroadcastBlockDataLeader broadcasts the block data as consensus group leader
func (messenger *instantBroadcastMessenger) BroadcastBlockDataLeader(_ data.HeaderHandler, miniBlocks map[uint32][]byte, transactions map[string][][]byte, pkBytes []byte) error {
	if messenger.shardCoordinator.SelfId() == common.MetachainShardId {
		return messenger.broadcastMiniblockData(miniBlocks, transactions, pkBytes)
	}

	return messenger.broadcastBlockDataLeaderWhenShard(miniBlocks, transactions, pkBytes)
}

// PrepareBroadcastBlockDataWithEquivalentProofs broadcasts v2 block data synchronously. The
// production messenger deliberately waits before broadcasting this data, but simulator rounds are
// advanced much faster than wall-clock time; retaining that delay makes cross-shard data arrive
// many simulated rounds after its header. The direct simulator path also broadcasts these complete
// maps synchronously after every committed block.
func (messenger *instantBroadcastMessenger) PrepareBroadcastBlockDataWithEquivalentProofs(
	_ data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	pkBytes []byte,
) {
	_ = messenger.broadcastMiniblockData(miniBlocks, transactions, pkBytes)
}

func (messenger *instantBroadcastMessenger) broadcastBlockDataLeaderWhenShard(miniBlocks map[uint32][]byte, transactions map[string][][]byte, pkBytes []byte) error {
	if len(miniBlocks) == 0 {
		return nil
	}

	metaMiniBlocks, metaTransactions := messenger.extractMetaMiniBlocksAndTransactions(miniBlocks, transactions)

	return messenger.broadcastMiniblockData(metaMiniBlocks, metaTransactions, pkBytes)
}

func (messenger *instantBroadcastMessenger) broadcastMiniblockData(miniBlocks map[uint32][]byte, transactions map[string][][]byte, pkBytes []byte) error {
	if len(miniBlocks) > 0 {
		err := messenger.BroadcastMiniBlocks(miniBlocks, pkBytes)
		if err != nil {
			log.Warn("instantBroadcastMessenger.BroadcastBlockData: broadcast miniblocks", "error", err.Error())
		}
	}

	if len(transactions) > 0 {
		err := messenger.BroadcastTransactions(transactions, pkBytes)
		if err != nil {
			log.Warn("instantBroadcastMessenger.BroadcastBlockData: broadcast transactions", "error", err.Error())
		}
	}

	return nil
}

func (messenger *instantBroadcastMessenger) extractMetaMiniBlocksAndTransactions(
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
) (map[uint32][]byte, map[string][][]byte) {

	metaMiniBlocks := make(map[uint32][]byte)
	metaTransactions := make(map[string][][]byte)

	for shardID, mbsMarshalized := range miniBlocks {
		if shardID != core.MetachainShardId {
			continue
		}

		metaMiniBlocks[shardID] = mbsMarshalized
		delete(miniBlocks, shardID)
	}

	identifier := messenger.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)

	for broadcastTopic, txsMarshalized := range transactions {
		if !strings.Contains(broadcastTopic, identifier) {
			continue
		}

		metaTransactions[broadcastTopic] = txsMarshalized
		delete(transactions, broadcastTopic)
	}

	return metaMiniBlocks, metaTransactions
}

// IsInterfaceNil returns true if there is no value under the interface
func (messenger *instantBroadcastMessenger) IsInterfaceNil() bool {
	return messenger == nil
}
