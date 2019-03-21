package mock

import (
	"time"
)

type SyncTimerMock struct {
	ClockOffsetCalled func() time.Duration
	CurrentTimeCalled func() time.Time
}

func (stm *SyncTimerMock) StartSync() {
	panic("implement me")
}

func (stm *SyncTimerMock) ClockOffset() time.Duration {
	if stm.ClockOffsetCalled != nil {
		return stm.ClockOffsetCalled()
	}

	return time.Duration(0)
}

func (stm *SyncTimerMock) FormattedCurrentTime() string {
	return time.Unix(0, 0).String()
}

func (stm *SyncTimerMock) CurrentTime() time.Time {
	if stm.CurrentTimeCalled != nil {
		return stm.CurrentTimeCalled()
	}

	return time.Unix(0, 0)
}
