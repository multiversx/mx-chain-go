package mock

import (
	"time"
)

type SyncStub struct {
}

func (ss *SyncStub) StartSync() {
}

func (ss *SyncStub) ClockOffset() time.Duration {
	return time.Second
}

func (ss *SyncStub) FormattedCurrentTime() string {
	return ""
}

func (ss *SyncStub) CurrentTime() time.Time {
	return time.Now()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *SyncStub) IsInterfaceNil() bool {
	return ss == nil
}
