package ntp

import (
	"fmt"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

// totalRequests defines the number of requests made to determine an accurate clock offset
const totalRequests = 10

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

// StartSync method does the time synchronization at every syncPeriod time elapsed. This should be started
// as a go routine
func (s *syncTime) StartSync() {
	for {
		s.sync()
		time.Sleep(s.syncPeriod)
	}
}

// sync method does the time synchronization and sets the current offset difference between local time
// and server time with which it has done the synchronization
func (s *syncTime) sync() {
	if s.query != nil {
		clockOffsetSum := time.Duration(0)
		succeededRequests := 0

		for i := 0; i < totalRequests; i++ {
			r, err := s.query("time.google.com")

			if err != nil {
				continue
			}

			succeededRequests++
			clockOffsetSum += r.ClockOffset
		}

		if succeededRequests > 0 {
			averrageClockOffset := time.Duration(int64(clockOffsetSum) / int64(succeededRequests))
			s.setClockOffset(averrageClockOffset)
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

// FormattedCurrentTime method gets the formatted current time on which is added a given offset
func (s *syncTime) FormattedCurrentTime() string {
	return s.formatTime(s.CurrentTime())
}

// formatTime method gets the formatted time from a given time
func (s *syncTime) formatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(),
		time.Minute(), time.Second(), time.Nanosecond())
	return str
}

// CurrentTime method gets the current time on which is added the current offset
func (s *syncTime) CurrentTime() time.Time {
	return time.Now().Add(s.clockOffset)
}
