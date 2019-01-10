package mock

import (
	"time"
)

type SyncStub struct {
}

func (ss *SyncStub) StartSync() {
	panic("implement me")
}

func (ss *SyncStub) ClockOffset() time.Duration {
	panic("implement me")
}

func (ss *SyncStub) FormattedCurrentTime(time.Duration) string {
	panic("implement me")
}

func (ss *SyncStub) CurrentTime(time.Duration) time.Time {
	panic("implement me")
}
