package mock

import (
	"time"
)

type AlarmSchedulerStub struct {
	AddCalled    func(func(alarmID string), time.Duration, string)
	CancelCalled func(string)
	CloseCalled  func()
	ResetCalled  func(string)
}

// Add -
func (a *AlarmSchedulerStub) Add(callback func(alarmID string), duration time.Duration, alarmID string) {
	if a.AddCalled != nil {
		a.AddCalled(callback, duration, alarmID)
	}
}

// Cancel -
func (a *AlarmSchedulerStub) Cancel(alarmID string) {
	if a.CancelCalled != nil {
		a.CancelCalled(alarmID)
	}
}

// Close -
func (a *AlarmSchedulerStub) Close() {
	if a.CloseCalled != nil {
		a.CloseCalled()
	}
}

// Reset -
func (a *AlarmSchedulerStub) Reset(alarmID string) {
	if a.ResetCalled != nil {
		a.ResetCalled(alarmID)
	}
}

// IsInterfaceNil -
func (a *AlarmSchedulerStub) IsInterfaceNil() bool {
	return a == nil
}
