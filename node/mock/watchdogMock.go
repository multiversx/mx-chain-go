package mock

import "time"

// WatchdogMock -
type WatchdogMock struct {
}

// Set -
func (w *WatchdogMock) Set(_ func(alarmID string), _ time.Duration, _ string) {
}

// SetDefault -
func (w *WatchdogMock) SetDefault(_ time.Duration, _ string) {
}

// Stop -
func (w *WatchdogMock) Stop(_ string) {
}

// Reset -
func (w *WatchdogMock) Reset(_ string) {
}

// IsInterfaceNil -
func (w *WatchdogMock) IsInterfaceNil() bool {
	return false
}
