package mock

import "time"

// TimerHandlerStub -
type TimerHandlerStub struct {
	CreateNewTimerCalled func(duration time.Duration)
	ShouldExecuteCalled  func() <-chan time.Time
	CloseCalled          func()
}

// CreateNewTimer -
func (stub *TimerHandlerStub) CreateNewTimer(duration time.Duration) {
	if stub.CreateNewTimerCalled != nil {
		stub.CreateNewTimerCalled(duration)
	}
}

// ShouldExecute -
func (stub *TimerHandlerStub) ShouldExecute() <-chan time.Time {
	if stub.ShouldExecuteCalled != nil {
		return stub.ShouldExecuteCalled()
	}

	return nil
}

// Close -
func (stub *TimerHandlerStub) Close() {
	if stub.CloseCalled != nil {
		stub.CloseCalled()
	}
}
