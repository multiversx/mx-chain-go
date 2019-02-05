package mock

import (
	"time"
)

// SyncTimerMock is a mock implementation of SyncTimer interface
type SyncTimerMock struct {
	StartSyncCalled            func()
	ClockOffsetCalled          func() time.Duration
	FormattedCurrentTimeCalled func(time.Duration) string
	CurrentTimeCalled          func(time.Duration) time.Time
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
func (s *SyncTimerMock) FormattedCurrentTime(t time.Duration) string {
	return s.FormattedCurrentTimeCalled(t)
}

// CurrentTime is a mock implementation for CurrentTime
func (s *SyncTimerMock) CurrentTime(t time.Duration) time.Time {
	return s.CurrentTimeCalled(t)
}
