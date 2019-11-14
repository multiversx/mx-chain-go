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
func (s *SyncTimerStub) StartSync() {
	s.StartSyncCalled()
}

// ClockOffset is a mock implementation for ClockOffset
func (s *SyncTimerStub) ClockOffset() time.Duration {
	return s.ClockOffsetCalled()
}

// FormattedCurrentTime is a mock implementation for FormattedCurrentTime
func (s *SyncTimerStub) FormattedCurrentTime() string {
	return s.FormattedCurrentTimeCalled()
}

// CurrentTime is a mock implementation for CurrentTime
func (s *SyncTimerStub) CurrentTime() time.Time {
	return s.CurrentTimeCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (stm *SyncTimerStub) IsInterfaceNil() bool {
	if stm == nil {
		return true
	}
	return false
}
