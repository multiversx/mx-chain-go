package old

import (
	"fmt"
	"time"
)

type SyncTimeMock struct {
	ClockOffset time.Duration
	RefTime     time.Time
}

func NewSyncTimeMock(clockOffset time.Duration, refTime time.Time) *SyncTimeMock {
	return &SyncTimeMock{ClockOffset: clockOffset, RefTime: refTime}
}

func (stm *SyncTimeMock) GetClockOffset() time.Duration {
	return stm.ClockOffset
}

func (stm *SyncTimeMock) GetFormatedCurrentTime() string {
	return stm.FormatTime(stm.GetCurrentTime())
}

func (stm *SyncTimeMock) FormatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
}

func (stm *SyncTimeMock) GetCurrentTime() time.Time {
	return stm.RefTime.Add(stm.ClockOffset)
}
