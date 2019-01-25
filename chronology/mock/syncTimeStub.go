package mock

import (
	"time"
)

type SyncTimeStub struct {
	StartSyncCalled            func()
	ClockOffsetCalled          func() time.Duration
	FormattedCurrentTimeCalled func(time.Duration) string
	CurrentTimeCalled          func(time.Duration) time.Time
}

func (sts *SyncTimeStub) StartSync() {
	sts.StartSyncCalled()
}

func (sts *SyncTimeStub) ClockOffset() time.Duration {
	return sts.ClockOffsetCalled()
}

func (sts *SyncTimeStub) FormattedCurrentTime(t time.Duration) string {
	return sts.FormattedCurrentTimeCalled(t)
}

func (sts *SyncTimeStub) CurrentTime(t time.Duration) time.Time {
	return sts.CurrentTimeCalled(t)
}
