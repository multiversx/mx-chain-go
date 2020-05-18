package mock

import (
	"time"
)

// SyncTimerMock is a mock implementation of SyncTimer interface
type SyncTimerMock struct {
	StartSyncingTimeCalled     func()
	ClockOffsetCalled          func() time.Duration
	FormattedCurrentTimeCalled func() string
	CurrentTimeCalled          func() time.Time
}

// StartSyncingTime is a mock implementation for StartSyncingTime
func (stm *SyncTimerMock) StartSyncingTime() {
	stm.StartSyncingTimeCalled()
}

// ClockOffset is a mock implementation for ClockOffset
func (stm *SyncTimerMock) ClockOffset() time.Duration {
	return stm.ClockOffsetCalled()
}

// FormattedCurrentTime is a mock implementation for FormattedCurrentTime
func (stm *SyncTimerMock) FormattedCurrentTime() string {
	return stm.FormattedCurrentTimeCalled()
}

// CurrentTime is a mock implementation for CurrentTime
func (stm *SyncTimerMock) CurrentTime() time.Time {
	return stm.CurrentTimeCalled()
}

// Close -
func (stm *SyncTimerMock) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stm *SyncTimerMock) IsInterfaceNil() bool {
	return stm == nil
}
