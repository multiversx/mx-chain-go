package mock

import (
	"time"
)

// SyncTimerMock mocks the implementation for a SyncTimer
type SyncTimerMock struct {
}

// StartSync method does the time synchronization at every syncPeriod time elapsed. This should be started as a go routine
func (stm *SyncTimerMock) StartSync() {
	panic("implement me")
}

// ClockOffset method gets the current time offset
func (stm *SyncTimerMock) ClockOffset() time.Duration {
	return time.Duration(0)
}

// FormattedCurrentTime method gets the formatted current time on which is added a given offset
func (stm *SyncTimerMock) FormattedCurrentTime(time.Duration) string {
	panic("implement me")
}

// CurrentTime method gets the current time on which is added the current offset
func (stm *SyncTimerMock) CurrentTime(time.Duration) time.Time {
	panic("implement me")
}
