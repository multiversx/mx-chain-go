package mock

import (
	"time"
)

// SyncTimerStub -
type SyncTimerStub struct {
}

// StartSyncingTime -
func (sts *SyncTimerStub) StartSyncingTime() {
}

// ClockOffset -
func (sts *SyncTimerStub) ClockOffset() time.Duration {
	return time.Second
}

// FormattedCurrentTime -
func (sts *SyncTimerStub) FormattedCurrentTime() string {
	return ""
}

// CurrentTime -
func (sts *SyncTimerStub) CurrentTime() time.Time {
	return time.Now()
}

// Close -
func (sts *SyncTimerStub) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sts *SyncTimerStub) IsInterfaceNil() bool {
	return sts == nil
}
