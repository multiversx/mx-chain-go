package components

import (
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/sharding"
)

type instantBroadcastMessenger struct {
	consensus.BroadcastMessenger
	shardCoordinator sharding.Coordinator
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
	}, nil
}

// BroadcastBlockDataLeader broadcasts the block data as consensus group leader
func (messenger *instantBroadcastMessenger) BroadcastBlockDataLeader(_ data.HeaderHandler, miniBlocks map[uint32][]byte, transactions map[string][][]byte, pkBytes []byte) error {
	if messenger.shardCoordinator.SelfId() == common.MetachainShardId {
		return messenger.broadcastMiniblockData(miniBlocks, transactions, pkBytes)
	}

	return messenger.broadcastBlockDataLeaderWhenShard(miniBlocks, transactions, pkBytes)
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
