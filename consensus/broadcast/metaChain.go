package broadcast

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type metaChain struct {
	*common
	marshalizer marshal.Marshalizer
	messenger   consensus.P2PMessenger
}

// NewMetaChain creates a new metaChain object
func NewMetaChain(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	privateKey crypto.PrivateKey,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
) (*metaChain, error) {

	err := checkMetaChainNilParameters(marshalizer, messenger, privateKey, shardCoordinator, singleSigner)
	if err != nil {
		return nil, err
	}

	cm := &common{
		marshalizer:      marshalizer,
		messenger:        messenger,
		privateKey:       privateKey,
		singleSigner:     singleSigner,
		shardCoordinator: shardCoordinator,
	}

	mc := &metaChain{
		common:      cm,
		marshalizer: marshalizer,
		messenger:   messenger,
	}

	return mc, nil
}

func checkMetaChainNilParameters(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	privateKey crypto.PrivateKey,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
) error {
	if marshalizer == nil {
		return spos.ErrNilMarshalizer
	}
	if messenger == nil {
		return spos.ErrNilMessenger
	}
	if privateKey == nil {
		return spos.ErrNilPrivateKey
	}
	if shardCoordinator == nil {
		return spos.ErrNilShardCoordinator
	}
	if singleSigner == nil {
		return spos.ErrNilSingleSigner
	}

	return nil
}

// BroadcastBlock will send on meta shard topics and on meta-to-shard topics the header
func (mc *metaChain) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	if header == nil {
		return spos.ErrNilMetaHeader
	}

	msgHeader, err := mc.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	go mc.messenger.Broadcast(factory.MetachainBlocksTopic, msgHeader)

	return nil
}

// BroadcastHeader will send on mini block topic the mini blocks
func (mc *metaChain) BroadcastHeader(header data.HeaderHandler) error {
	// meta chain does not need to broadcast separately the header, as it have no body and BroadcastBlock does all
	// the job for it, but this method is created to satisfy the BroadcastMessenger interface
	return nil
}

// BroadcastMiniBlocks will send on mini block topic the mini blocks
func (mc *metaChain) BroadcastMiniBlocks(miniBlocks map[uint32][]byte) error {
	// meta chain does not need to broadcast miniblocks but this method is created to satisfy the BroadcastMessenger
	// interface
	return nil
}

// BroadcastTransactions will send on transaction topic the transactions
func (mc *metaChain) BroadcastTransactions(transactions map[uint32][][]byte) error {
	// meta chain does not need to broadcast transactions but this method is created to satisfy the BroadcastMessenger
	// interface
	return nil
}
