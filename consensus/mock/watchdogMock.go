package mock

import (
	"time"
)

// WatchdogMock -
type WatchdogMock struct {
	SetCalled  func(callback func(alarmID string), duration time.Duration, alarmID string)
	StopCalled func(alarmID string)
}

// Set -
func (w *WatchdogMock) Set(callback func(alarmID string), duration time.Duration, alarmID string) {
	if w.SetCalled != nil {
		w.SetCalled(callback, duration, alarmID)
	}
}

// SetDefault -
func (w *WatchdogMock) SetDefault(duration time.Duration, alarmID string) {
}

// Stop -
func (w *WatchdogMock) Stop(alarmID string) {
	if w.StopCalled != nil {
		w.StopCalled(alarmID)
	}
}

// Reset -
func (w *WatchdogMock) Reset(alarmID string) {
}

// IsInterfaceNil -
func (w *WatchdogMock) IsInterfaceNil() bool {
	return false
}
