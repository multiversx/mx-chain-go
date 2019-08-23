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

func (ss *SyncStub) FormattedCurrentTime() string {
	panic("implement me")
}

func (ss *SyncStub) CurrentTime() time.Time {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *SyncStub) IsInterfaceNil() bool {
	if ss == nil {
		return true
	}
	return false
}
