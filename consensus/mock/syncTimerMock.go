package mock

import (
	"time"
)

// SyncTimerMock mocks the implementation for a SyncTimer
type SyncTimerMock struct {
	ClockOffsetCalled func() time.Duration
	CurrentTimeCalled func() time.Time
}

// StartSyncingTime method does the time synchronization at every syncPeriod time elapsed. This should be started as a go routine
func (stm *SyncTimerMock) StartSyncingTime() {
	panic("implement me")
}

// ClockOffset method gets the current time offset
func (stm *SyncTimerMock) ClockOffset() time.Duration {
	if stm.ClockOffsetCalled != nil {
		return stm.ClockOffsetCalled()
	}

	return time.Duration(0)
}

// FormattedCurrentTime method gets the formatted current time on which is added a given offset
func (stm *SyncTimerMock) FormattedCurrentTime() string {
	return time.Unix(0, 0).String()
}

// CurrentTime method gets the current time on which is added the current offset
func (stm *SyncTimerMock) CurrentTime() time.Time {
	if stm.CurrentTimeCalled != nil {
		return stm.CurrentTimeCalled()
	}

	return time.Unix(0, 0)
}

// Close -
func (stm *SyncTimerMock) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stm *SyncTimerMock) IsInterfaceNil() bool {
	return stm == nil
}
