package mock

import "time"

// SyncTimerStub -
type SyncTimerStub struct {
	CurrentTimeCalled func() time.Time
}

// CurrentTime -
func (sts *SyncTimerStub) CurrentTime() time.Time {
	if sts.CurrentTimeCalled != nil {
		return sts.CurrentTimeCalled()
	}

	return time.Time{}
}

// IsInterfaceNil -
func (sts *SyncTimerStub) IsInterfaceNil() bool {
	return sts == nil
}
