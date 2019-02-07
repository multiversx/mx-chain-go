package mock

import (
	"time"
)

type SyncTimeMock struct {
	CurrentTimeCalled func(time.Duration) time.Time
}

func (stm *SyncTimeMock) StartSync() {
}

func (stm *SyncTimeMock) ClockOffset() time.Duration {
	return time.Duration(0)
}

func (stm *SyncTimeMock) FormattedCurrentTime(t time.Duration) string {
	return "formatted time"
}

func (stm *SyncTimeMock) CurrentTime(t time.Duration) time.Time {
	return stm.CurrentTimeCalled(t)
}
