package mock

import (
	"time"
)

// SyncTimerStub is a mock implementation of SyncTimer interface
type SyncTimerStub struct {
	StartSyncingTimeCalled     func()
	ClockOffsetCalled          func() time.Duration
	FormattedCurrentTimeCalled func() string
	CurrentTimeCalled          func() time.Time
}

// StartSyncingTime is a mock implementation for StartSyncingTime
func (sts *SyncTimerStub) StartSyncingTime() {
	sts.StartSyncingTimeCalled()
}

// ClockOffset is a mock implementation for ClockOffset
func (sts *SyncTimerStub) ClockOffset() time.Duration {
	return sts.ClockOffsetCalled()
}

// FormattedCurrentTime is a mock implementation for FormattedCurrentTime
func (sts *SyncTimerStub) FormattedCurrentTime() string {
	return sts.FormattedCurrentTimeCalled()
}

// CurrentTime is a mock implementation for CurrentTime
func (sts *SyncTimerStub) CurrentTime() time.Time {
	return sts.CurrentTimeCalled()
}

// Close -
func (sts *SyncTimerStub) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sts *SyncTimerStub) IsInterfaceNil() bool {
	return sts == nil
}
