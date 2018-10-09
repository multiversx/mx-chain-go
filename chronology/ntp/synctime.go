package ntp

import (
	"fmt"
	"sync"
	"time"
)

type SyncTimer interface {
	GetClockOffset() time.Duration
	GetFormatedCurrentTime() string
	FormatTime(time time.Time) string
	GetCurrentTime() time.Time
}

type SyncTime struct {
	mut         sync.RWMutex
	clockOffset time.Duration
	syncPeriod  time.Duration
	query       func(host string) (*Response, error)
}

func NewSyncTime(syncPeriod time.Duration, query func(host string) (*Response, error)) SyncTime {
	s := SyncTime{clockOffset: 0, syncPeriod: syncPeriod, query: query}
	go s.synchronize()
	return s
}

func (s *SyncTime) synchronize() {
	for {
		s.doSync()
		time.Sleep(s.syncPeriod)
	}
}

func (s *SyncTime) doSync() {
	r, err := s.query("time.google.com")

	s.mut.Lock()

	if err != nil {
		s.clockOffset = 0
	} else {
		s.clockOffset = r.ClockOffset
	}

	s.mut.Unlock()
}

func (s *SyncTime) GetClockOffset() time.Duration {
	s.mut.RLock()
	clockOffset := s.clockOffset
	s.mut.RUnlock()

	return clockOffset
}

func (s *SyncTime) GetFormatedCurrentTime() string {
	return s.FormatTime(s.GetCurrentTime())
}

func (s *SyncTime) FormatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
}

func (s *SyncTime) getCurrentTime(ref time.Time) time.Time {
	return ref.Add(s.GetClockOffset())
}

func (s *SyncTime) GetCurrentTime() time.Time {
	return s.getCurrentTime(time.Now())
}
