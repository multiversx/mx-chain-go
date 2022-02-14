package mock

import "time"

// SenderHandlerStub -
type SenderHandlerStub struct {
	ShouldExecuteCalled func() <-chan time.Time
	ExecuteCalled       func()
	CloseCalled         func()
}

// ShouldExecute -
func (stub *SenderHandlerStub) ShouldExecute() <-chan time.Time {
	if stub.ShouldExecuteCalled != nil {
		return stub.ShouldExecuteCalled()
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
