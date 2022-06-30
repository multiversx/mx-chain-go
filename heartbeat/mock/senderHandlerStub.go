package mock

import "time"

// SenderHandlerStub -
type SenderHandlerStub struct {
	ExecutionReadyChannelCalled func() <-chan time.Time
	ExecuteCalled               func()
	CloseCalled                 func()
}

// ExecutionReadyChannel -
func (stub *SenderHandlerStub) ExecutionReadyChannel() <-chan time.Time {
	if stub.ExecutionReadyChannelCalled != nil {
		return stub.ExecutionReadyChannelCalled()
	}

	return nil
}

// Execute -
func (stub *SenderHandlerStub) Execute() {
	if stub.ExecuteCalled != nil {
		stub.ExecuteCalled()
	}
}

// Close -
func (stub *SenderHandlerStub) Close() {
	if stub.CloseCalled != nil {
		stub.CloseCalled()
	}
}

// IsInterfaceNil -
func (stub *SenderHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
