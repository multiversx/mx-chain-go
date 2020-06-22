package health

import (
	"time"
)

var _ clock = (*realClock)(nil)

type realClock struct {
}

func (_ *realClock) now() time.Time {
	return time.Now()
}

func (_ *realClock) after(d time.Duration) <-chan time.Time {
	return time.After(d)
}
