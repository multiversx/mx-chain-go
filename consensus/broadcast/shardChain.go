package broadcast

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type shardChain struct {
	*common
	marshalizer      marshal.Marshalizer
	messenger        consensus.P2PMessenger
	shardCoordinator sharding.Coordinator
	syncTimer        ntp.SyncTimer
}

// NewShardChain creates a new shardChain object
func NewShardChain(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	shardCoordinator sharding.Coordinator,
	privateKey crypto.PrivateKey,
	singleSigner crypto.SingleSigner,
	syncTimer ntp.SyncTimer,
) (*shardChain, error) {

	err := checkShardChainNilParameters(marshalizer, messenger, shardCoordinator, privateKey, singleSigner, syncTimer)
	if err != nil {
		return nil, err
	}

	cm := &common{
		marshalizer:      marshalizer,
		messenger:        messenger,
		privateKey:       privateKey,
		shardCoordinator: shardCoordinator,
		singleSigner:     singleSigner,
	}

	sc := &shardChain{
		common:           cm,
		marshalizer:      marshalizer,
		messenger:        messenger,
		shardCoordinator: shardCoordinator,
		syncTimer:        syncTimer,
	}

	return sc, nil
}

func checkShardChainNilParameters(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	shardCoordinator sharding.Coordinator,
	privateKey crypto.PrivateKey,
	singleSigner crypto.SingleSigner,
	syncTimer ntp.SyncTimer,
) error {
	if marshalizer == nil {
		return spos.ErrNilMarshalizer
	}
	if messenger == nil {
		return spos.ErrNilMessenger
	}
	if shardCoordinator == nil {
		return spos.ErrNilShardCoordinator
	}
	if privateKey == nil {
		return spos.ErrNilPrivateKey
	}
	if singleSigner == nil {
		return spos.ErrNilSingleSigner
	}
	if syncTimer == nil {
		return spos.ErrNilSyncTimer
	}

	return nil
}

// BroadcastBlock will send on intra shard topics the header and block body
func (sc *shardChain) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	if blockBody == nil {
		return spos.ErrNilBody
	}

	err := blockBody.IntegrityAndValidity()
	if err != nil {
		return err
	}

	if header == nil {
		return spos.ErrNilHeader
	}

	msgHeader, err := sc.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	msgBlockBody, err := sc.marshalizer.Marshal(blockBody)
	if err != nil {
		return err
	}

	selfIdentifier := sc.shardCoordinator.CommunicationIdentifier(sc.shardCoordinator.SelfId())

	go sc.messenger.Broadcast(factory.HeadersTopic+selfIdentifier, msgHeader)

	//TODO: Investigate if the body block needs to be sent on intra shard topic as each miniblock is already sent on cross shard topics
	go sc.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, msgBlockBody)

	return nil
}

// BroadcastHeader will send on metachain topics the header
func (sc *shardChain) BroadcastHeader(header data.HeaderHandler) error {
	if header == nil {
		return spos.ErrNilHeader
	}

	msgHeader, err := sc.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	shardHeaderForMetachainTopic := factory.ShardHeadersForMetachainTopic +
		sc.shardCoordinator.CommunicationIdentifier(sharding.MetachainShardId)

	go sc.messenger.Broadcast(shardHeaderForMetachainTopic, msgHeader)

	return nil
}

// BroadcastMiniBlocks will send on miniblock topic the miniblocks
func (sc *shardChain) BroadcastMiniBlocks(miniBlocks map[uint32][]byte) error {
	mbs := 0
	for k, v := range miniBlocks {
		mbs++
		miniBlocksTopic := factory.MiniBlocksTopic +
			sc.shardCoordinator.CommunicationIdentifier(k)

		go sc.messenger.Broadcast(miniBlocksTopic, v)
	}

	if mbs > 0 {
		log.Info(fmt.Sprintf("%sStep 1: Sent %d miniblocks\n", sc.syncTimer.FormattedCurrentTime(), mbs))
	}

	return nil
}

// BroadcastTransactions will send on transaction topic the transactions
func (sc *shardChain) BroadcastTransactions(transactions map[uint32][][]byte) error {
	dataPacker, err := partitioning.NewSizeDataPacker(sc.marshalizer)
	if err != nil {
		return err
	}

	txs := 0
	for k, v := range transactions {
		txs += len(v)
		// forward txs to the destination shards in packets
		packets, err := dataPacker.PackDataInChunks(v, core.MaxBulkTransactionSize)
		if err != nil {
			return err
		}

		for _, buff := range packets {
			transactionTopic := factory.TransactionTopic +
				sc.shardCoordinator.CommunicationIdentifier(k)

			go sc.messenger.Broadcast(transactionTopic, buff)
		}
	}

	if txs > 0 {
		log.Info(fmt.Sprintf("%sStep 1: Sent %d transactions\n", sc.syncTimer.FormattedCurrentTime(), txs))
	}

	return nil
}
