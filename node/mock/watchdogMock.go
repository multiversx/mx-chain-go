package mock

import "time"

// WatchdogMock -
type WatchdogMock struct {
}

// Set -
func (w *WatchdogMock) Set(callback func(alarmID string), duration time.Duration, alarmID string) {
}

// SetDefault -
func (w *WatchdogMock) SetDefault(duration time.Duration, alarmID string) {
}

// Stop -
func (w *WatchdogMock) Stop(alarmID string) {
}

// Reset -
func (w *WatchdogMock) Reset(alarmID string) {
}

// IsInterfaceNil -
func (w *WatchdogMock) IsInterfaceNil() bool {
	return false
}
