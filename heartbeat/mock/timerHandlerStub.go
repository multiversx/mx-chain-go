package mock

import "time"

// TimerHandlerStub -
type TimerHandlerStub struct {
	CreateNewTimerCalled        func(duration time.Duration)
	ExecutionReadyChannelCalled func() <-chan time.Time
	CloseCalled                 func()
}

// CreateNewTimer -
func (stub *TimerHandlerStub) CreateNewTimer(duration time.Duration) {
	if stub.CreateNewTimerCalled != nil {
		stub.CreateNewTimerCalled(duration)
	}
}

// ExecutionReadyChannel -
func (stub *TimerHandlerStub) ExecutionReadyChannel() <-chan time.Time {
	if stub.ExecutionReadyChannelCalled != nil {
		return stub.ExecutionReadyChannelCalled()
	}

	return nil
}

// Close -
func (stub *TimerHandlerStub) Close() {
	if stub.CloseCalled != nil {
		stub.CloseCalled()
	}
}
