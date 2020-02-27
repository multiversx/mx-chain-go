package mock

import (
	"time"
)

// SyncStub -
type SyncStub struct {
}

// StartSync -
func (ss *SyncStub) StartSync() {
}

// ClockOffset -
func (ss *SyncStub) ClockOffset() time.Duration {
	return time.Second
}

// FormattedCurrentTime -
func (ss *SyncStub) FormattedCurrentTime() string {
	return ""
}

// CurrentTime -
func (ss *SyncStub) CurrentTime() time.Time {
	return time.Now()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *SyncStub) IsInterfaceNil() bool {
	return ss == nil
}
