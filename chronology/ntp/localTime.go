package ntp

import (
	"fmt"
	"time"
)

type LocalTime struct {
	ClockOffset time.Duration
}

func (lt *LocalTime) GetFormatedCurrentTime(clockOffset time.Duration) string {
	return lt.FormatTime(lt.GetCurrentTime(clockOffset))
}

func (lt *LocalTime) GetCurrentTime(clockOffset time.Duration) time.Time {
	return time.Now().Add(clockOffset)
}

func (lt *LocalTime) GetClockOffset() time.Duration {
	return lt.ClockOffset
}

func (lt *LocalTime) FormatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
}
