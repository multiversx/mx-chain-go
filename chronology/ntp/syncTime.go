package ntp

import (
	"fmt"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

// SyncTimer defines an interface for time synchronization
type SyncTimer interface {
	StartSync()
	ClockOffset() time.Duration
	FormattedCurrentTime(time.Duration) string
	CurrentTime(time.Duration) time.Time
}

// syncTime defines an object for time synchronization
type syncTime struct {
	mut         sync.RWMutex
	clockOffset time.Duration
	syncPeriod  time.Duration
	query       func(host string) (*ntp.Response, error)
}

// NewSyncTime creates a syncTime object
func NewSyncTime(syncPeriod time.Duration, query func(host string) (*ntp.Response, error)) *syncTime {
	s := syncTime{clockOffset: 0, syncPeriod: syncPeriod, query: query}
	return &s
}

// StartSync method does the time syncronization at every syncPeriod time elapsed. This should be started
// as a go routine
func (s *syncTime) StartSync() {
	for {
		s.sync()
		time.Sleep(s.syncPeriod)
	}
}

// sync method does the time syncronization and sets the current offset difference between local time
// and server time with wich it has done the syncronization
func (s *syncTime) sync() {
	if s.query != nil {
		r, err := s.query("time.google.com")

		if err != nil {
			s.setClockOffset(0)
		} else {
			s.setClockOffset(r.ClockOffset)
		}
	}
}

// ClockOffset method gets the current time offset
func (s *syncTime) ClockOffset() time.Duration {
	s.mut.RLock()
	clockOffset := s.clockOffset
	s.mut.RUnlock()

	return clockOffset
}

func (s *syncTime) setClockOffset(clockOffset time.Duration) {
	s.mut.Lock()
	s.clockOffset = clockOffset
	s.mut.Unlock()
}

// FormattedCurrentTime method gets the formatted current time on wich is added a given offset
func (s *syncTime) FormattedCurrentTime(clockOffset time.Duration) string {
	return s.formatTime(s.CurrentTime(clockOffset))
}

// formatTime method gets the formatted time from a given time
func (s *syncTime) formatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(),
		time.Minute(), time.Second(), time.Nanosecond())
	return str
}

// CurrentTime method gets the current time on which is added the current offset
func (s *syncTime) CurrentTime(clockOffset time.Duration) time.Time {
	return time.Now().Add(clockOffset)
}
