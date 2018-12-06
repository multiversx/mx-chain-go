package ntp

import (
	"fmt"
	"time"
)

type LocalTime struct {
	clockOffset time.Duration
}

func (lt *LocalTime) FormatedCurrentTime(clockOffset time.Duration) string {
	return lt.FormatTime(lt.CurrentTime(clockOffset))
}

func (lt *LocalTime) CurrentTime(clockOffset time.Duration) time.Time {
	return time.Now().Add(clockOffset)
}

func (lt *LocalTime) ClockOffset() time.Duration {
	return lt.clockOffset
}

func (lt *LocalTime) SetClockOffset(clockOffset time.Duration) {
	lt.clockOffset = clockOffset
}

func (lt *LocalTime) FormatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(),
		time.Minute(), time.Second(), time.Nanosecond())
	return str
}
