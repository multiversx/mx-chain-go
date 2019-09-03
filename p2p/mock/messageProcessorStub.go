package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type MessageProcessorStub struct {
	ProcessMessageCalled func(message p2p.MessageP2P) error
}

func (mps *MessageProcessorStub) ProcessReceivedMessage(message p2p.MessageP2P) error {
	return mps.ProcessMessageCalled(message)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mps *MessageProcessorStub) IsInterfaceNil() bool {
	if mps == nil {
		return true
	}
	return false
}
