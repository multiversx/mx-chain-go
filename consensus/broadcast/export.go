package broadcast

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// SignMessage will sign and return the given message
func (cm *commonMessenger) SignMessage(message *consensus.Message) ([]byte, error) {
	return cm.signMessage(message)
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
