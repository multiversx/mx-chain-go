package watchdog

import "time"

// DisabledWatchdog represents a disabled WatchdogTimer implementation
type DisabledWatchdog struct {
}

// Set does nothing
func (dw *DisabledWatchdog) Set(_ func(alarmID string), _ time.Duration, _ string) {
}

// SetDefault does nothing
func (dw *DisabledWatchdog) SetDefault(_ time.Duration, _ string) {
}

// Stop does nothing
func (dw *DisabledWatchdog) Stop(_ string) {
}

// Reset does nothing
func (dw *DisabledWatchdog) Reset(_ string) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (dw *DisabledWatchdog) IsInterfaceNil() bool {
	return dw == nil
}
