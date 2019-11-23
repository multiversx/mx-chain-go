package broadcast

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("consensus/broadcast")

type commonMessenger struct {
	marshalizer      marshal.Marshalizer
	messenger        consensus.P2PMessenger
	privateKey       crypto.PrivateKey
	shardCoordinator sharding.Coordinator
	singleSigner     crypto.SingleSigner
}

// BroadcastConsensusMessage will send on consensus topic the consensus message
func (cm *commonMessenger) BroadcastConsensusMessage(message *consensus.Message) error {
	signature, err := cm.signMessage(message)
	if err != nil {
		return err
	}

	message.Signature = signature

	buff, err := cm.marshalizer.Marshal(message)
	if err != nil {
		return err
	}

	consensusTopic := core.ConsensusTopic +
		cm.shardCoordinator.CommunicationIdentifier(cm.shardCoordinator.SelfId())

	go cm.messenger.Broadcast(consensusTopic, buff)

	return nil
}

func (cm *commonMessenger) signMessage(message *consensus.Message) ([]byte, error) {
	buff, err := cm.marshalizer.Marshal(message)
	if err != nil {
		return nil, err
	}

	signature, err := cm.singleSigner.Sign(cm.privateKey, buff)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// BroadcastMiniBlocks will send on miniblocks topic the cross-shard miniblocks
func (cm *commonMessenger) BroadcastMiniBlocks(miniBlocks map[uint32][]byte) error {
	for k, v := range miniBlocks {
		miniBlocksTopic := factory.MiniBlocksTopic +
			cm.shardCoordinator.CommunicationIdentifier(k)

		go cm.messenger.Broadcast(miniBlocksTopic, v)
	}

	if len(miniBlocks) > 0 {
		log.Debug("sent miniblocks",
			"num minblocks", len(miniBlocks),
		)
	}

	return nil
}

// BroadcastTransactions will send on transaction topic the transactions
func (cm *commonMessenger) BroadcastTransactions(transactions map[string][][]byte) error {
	dataPacker, err := partitioning.NewSimpleDataPacker(cm.marshalizer)
	if err != nil {
		return err
	}

	txs := 0
	for topic, v := range transactions {
		txs += len(v)
		// forward txs to the destination shards in packets
		packets, err := dataPacker.PackDataInChunks(v, core.MaxBulkTransactionSize)
		if err != nil {
			return err
		}

		for _, buff := range packets {
			go cm.messenger.Broadcast(topic, buff)
		}
	}

	if txs > 0 {
		log.Debug("sent transactions",
			"num txs", txs,
		)
	}

	return nil
}
