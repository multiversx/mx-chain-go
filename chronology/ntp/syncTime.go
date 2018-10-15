package ntp

import (
	"fmt"
	"sync"
	"time"
)

type SyncTimer interface {
	GetClockOffset() time.Duration
	GetFormatedCurrentTime(time.Duration) string
	GetCurrentTime(time.Duration) time.Time
}

type SyncTime struct {
	mut         sync.RWMutex
	clockOffset time.Duration
	syncPeriod  time.Duration
	query       func(host string) (*Response, error)
}

func NewSyncTime(syncPeriod time.Duration, query func(host string) (*Response, error)) *SyncTime {
	s := SyncTime{clockOffset: 0, syncPeriod: syncPeriod, query: query}
	go s.synchronize()
	return &s
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
	defer s.mut.Unlock()

	if err != nil {
		s.clockOffset = 0
	} else {
		s.clockOffset = r.ClockOffset
	}
}

func (s *SyncTime) GetClockOffset() time.Duration {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.clockOffset
}

func (s *SyncTime) GetFormatedCurrentTime(clockOffset time.Duration) string {
	return s.FormatTime(s.GetCurrentTime(clockOffset))
}

func (s *SyncTime) FormatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
}

func (s *SyncTime) GetCurrentTime(clockOffset time.Duration) time.Time {
	return time.Now().Add(clockOffset)
}
