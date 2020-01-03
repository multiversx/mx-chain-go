package mock

import (
	"time"
)

// SyncTimerMock is a mock implementation of SyncTimer interface
type SyncTimerMock struct {
	StartSyncCalled            func()
	ClockOffsetCalled          func() time.Duration
	FormattedCurrentTimeCalled func() string
	CurrentTimeCalled          func() time.Time
}

// StartSync is a mock implementation for StartSync
func (stm *SyncTimerMock) StartSync() {
	stm.StartSyncCalled()
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

// IsInterfaceNil returns true if there is no value under the interface
func (stm *SyncTimerMock) IsInterfaceNil() bool {
	return stm == nil
}
