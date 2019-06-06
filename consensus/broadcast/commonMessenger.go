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
