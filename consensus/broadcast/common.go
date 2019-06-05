package broadcast

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

var log = logger.DefaultLogger()

type common struct {
	marshalizer      marshal.Marshalizer
	messenger        consensus.P2PMessenger
	privateKey       crypto.PrivateKey
	shardCoordinator sharding.Coordinator
	singleSigner     crypto.SingleSigner
}

// BroadcastConsensusMessage will send on consensus topics the consensus message
func (c *common) BroadcastConsensusMessage(message *consensus.Message) error {
	signature, err := c.signMessage(message)
	if err != nil {
		return err
	}

	message.Signature = signature

	buff, err := c.marshalizer.Marshal(message)
	if err != nil {
		return err
	}

	consensusTopic := core.ConsensusTopic +
		c.shardCoordinator.CommunicationIdentifier(c.shardCoordinator.SelfId())

	go c.messenger.Broadcast(consensusTopic, buff)

	return nil
}

func (c *common) signMessage(message *consensus.Message) ([]byte, error) {
	buff, err := c.marshalizer.Marshal(message)
	if err != nil {
		return nil, err
	}

	signature, err := c.singleSigner.Sign(c.privateKey, buff)
	if err != nil {
		return nil, err
	}

	return signature, nil
}
