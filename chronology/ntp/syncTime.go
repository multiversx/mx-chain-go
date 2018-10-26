package ntp

import (
	"fmt"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

type SyncTimer interface {
	ClockOffset() time.Duration
	FormatedCurrentTime(time.Duration) string
	CurrentTime(time.Duration) time.Time
}

type SyncTime struct {
	mut         sync.RWMutex
	clockOffset time.Duration
	syncPeriod  time.Duration
	query       func(host string) (*ntp.Response, error)
}

func NewSyncTime(syncPeriod time.Duration, query func(host string) (*ntp.Response, error)) *SyncTime {
	s := SyncTime{clockOffset: 0, syncPeriod: syncPeriod, query: query}
	go s.synchronize()
	return &s
}

func (s *SyncTime) synchronize() {
	for {
		s.DoSync()
		time.Sleep(s.syncPeriod)
	}
}

func (s *SyncTime) DoSync() {
	r, err := s.query("time.google.com")

	if err != nil {
		s.SetClockOffset(0)
	} else {
		s.SetClockOffset(r.ClockOffset)
	}
}

func (s *SyncTime) Query() func(host string) (*ntp.Response, error) {
	return s.query
}

func (s *SyncTime) SetQuery(query func(host string) (*ntp.Response, error)) {
	s.query = query
}

func (s *SyncTime) SyncPeriod() time.Duration {
	return s.syncPeriod
}

func (s *SyncTime) SetSyncPeriod(syncPeriod time.Duration) {
	s.syncPeriod = syncPeriod
}

func (s *SyncTime) ClockOffset() time.Duration {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.clockOffset
}

func (s *SyncTime) SetClockOffset(clockOffset time.Duration) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.clockOffset = clockOffset
}

func (s *SyncTime) FormatedCurrentTime(clockOffset time.Duration) string {
	return s.FormatTime(s.CurrentTime(clockOffset))
}

func (s *SyncTime) FormatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(),
		time.Minute(), time.Second(), time.Nanosecond())
	return str
}

func (s *SyncTime) CurrentTime(clockOffset time.Duration) time.Time {
	return time.Now().Add(clockOffset)
}
