package mock

import (
	"time"
)

// SyncTimerStub is a mock implementation of SyncTimer interface
type SyncTimerStub struct {
	StartSyncCalled            func()
	ClockOffsetCalled          func() time.Duration
	FormattedCurrentTimeCalled func() string
	CurrentTimeCalled          func() time.Time
}

// StartSync is a mock implementation for StartSync
func (stm *SyncTimerStub) StartSync() {
	stm.StartSyncCalled()
}

// ClockOffset is a mock implementation for ClockOffset
func (stm *SyncTimerStub) ClockOffset() time.Duration {
	return stm.ClockOffsetCalled()
}

// FormattedCurrentTime is a mock implementation for FormattedCurrentTime
func (stm *SyncTimerStub) FormattedCurrentTime() string {
	return stm.FormattedCurrentTimeCalled()
}

// CurrentTime is a mock implementation for CurrentTime
func (stm *SyncTimerStub) CurrentTime() time.Time {
	return stm.CurrentTimeCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (stm *SyncTimerStub) IsInterfaceNil() bool {
	return stm == nil
}
