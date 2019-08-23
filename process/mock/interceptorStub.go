package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type InterceptorStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
}

func (is *InterceptorStub) ProcessReceivedMessage(message p2p.MessageP2P) error {
	return is.ProcessReceivedMessageCalled(message)
}

// IsInterfaceNil returns true if there is no value under the interface
func (is *InterceptorStub) IsInterfaceNil() bool {
	if is == nil {
		return true
	}
	return false
}
