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
func (s *SyncTimerMock) StartSync() {
	s.StartSyncCalled()
}

// ClockOffset is a mock implementation for ClockOffset
func (s *SyncTimerMock) ClockOffset() time.Duration {
	return s.ClockOffsetCalled()
}

// FormattedCurrentTime is a mock implementation for FormattedCurrentTime
func (s *SyncTimerMock) FormattedCurrentTime() string {
	return s.FormattedCurrentTimeCalled()
}

// CurrentTime is a mock implementation for CurrentTime
func (s *SyncTimerMock) CurrentTime() time.Time {
	return s.CurrentTimeCalled()
}
