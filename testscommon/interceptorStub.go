package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptorStub -
type InterceptorStub struct {
	ProcessReceivedMessageCalled     func(message p2p.MessageP2P) error
	SetInterceptedDebugHandlerCalled func(debugger process.InterceptedDebugger) error
	RegisterHandlerCalled            func(handler func(topic string, hash []byte, data interface{}))
	CloseCalled                      func() error
}

// ProcessReceivedMessage -
func (is *InterceptorStub) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) error {
	if is.ProcessReceivedMessageCalled != nil {
		return is.ProcessReceivedMessageCalled(message)
	}

	return nil
}

// SetInterceptedDebugHandler -
func (is *InterceptorStub) SetInterceptedDebugHandler(debugger process.InterceptedDebugger) error {
	if is.SetInterceptedDebugHandlerCalled != nil {
		return is.SetInterceptedDebugHandlerCalled(debugger)
	}

	return nil
}

// RegisterHandler -
func (is *InterceptorStub) RegisterHandler(handler func(topic string, hash []byte, data interface{})) {
	if is.RegisterHandlerCalled != nil {
		is.RegisterHandlerCalled(handler)
	}
}

// Close -
func (is *InterceptorStub) Close() error {
	if is.CloseCalled != nil {
		return is.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (is *InterceptorStub) IsInterfaceNil() bool {
	return is == nil
}
