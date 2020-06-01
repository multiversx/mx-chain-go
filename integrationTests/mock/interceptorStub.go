package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// InterceptorStub -
type InterceptorStub struct {
	ProcessReceivedMessageCalled     func(message p2p.MessageP2P) error
	SetInterceptedDebugHandlerCalled func(handler process.InterceptedDebugHandler) error
}

// ProcessReceivedMessage -
func (is *InterceptorStub) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	return is.ProcessReceivedMessageCalled(message)
}

// SetInterceptedDebugHandler -
func (is *InterceptorStub) SetInterceptedDebugHandler(handler process.InterceptedDebugHandler) error {
	if is.SetInterceptedDebugHandlerCalled != nil {
		return is.SetInterceptedDebugHandlerCalled(handler)
	}

	return nil
}

// RegisterHandler -
func (is *InterceptorStub) RegisterHandler(_ func(toShard uint32, data []byte)) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (is *InterceptorStub) IsInterfaceNil() bool {
	return is == nil
}
