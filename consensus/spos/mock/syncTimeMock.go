package mock

import (
	"time"
)

type SyncTimerMock struct {
}

func (stm *SyncTimerMock) StartSync() {
	panic("implement me")
}

func (stm *SyncTimerMock) ClockOffset() time.Duration {
	return time.Duration(0)
}

func (stm *SyncTimerMock) FormattedCurrentTime(time.Duration) string {
	panic("implement me")
}

func (stm *SyncTimerMock) CurrentTime(time.Duration) time.Time {
	panic("implement me")
}
