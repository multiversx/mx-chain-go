package mock

import "time"

// TimerHandlerStub -
type TimerHandlerStub struct {
	CreateNewTimerCalled func(duration time.Duration)
}

// CreateNewTimer -
func (stub *TimerHandlerStub) CreateNewTimer(duration time.Duration) {
	if stub.CreateNewTimerCalled != nil {
		stub.CreateNewTimerCalled(duration)
	}
}
