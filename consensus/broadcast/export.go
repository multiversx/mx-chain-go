package broadcast

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// SignMessage will sign and return the given message
func (cm *commonMessenger) SignMessage(message *consensus.Message) ([]byte, error) {
	return cm.signMessage(message)
}

// HeaderReceived is the callback registered by the shard chain messenger
// to be called when a header is added to the headers pool
func (scm *shardChainMessenger) HeaderReceived(headerHandler data.HeaderHandler, hash []byte) {
	scm.headerReceived(headerHandler, hash)
}

// GetShardHeaderHashesFromMetachainBlock returns the header hashes for specified shard ID from the given metaHeader
func GetShardHeaderHashesFromMetachainBlock(headerHandler data.HeaderHandler, shardID uint32) ([][]byte, []*headerDataForValidator, error) {
	return getShardHeaderHashesFromMetachainBlock(headerHandler, shardID)
}

// NewCommonMessenger will return a new instance of a commonMessenger
func NewCommonMessenger(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	privateKey crypto.PrivateKey,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
) (*commonMessenger, error) {

	return &commonMessenger{
		marshalizer:      marshalizer,
		messenger:        messenger,
		privateKey:       privateKey,
		shardCoordinator: shardCoordinator,
		singleSigner:     singleSigner,
	}, nil
}
