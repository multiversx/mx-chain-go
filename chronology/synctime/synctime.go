package synctime

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"sync"
	"time"
)

type SyncTime struct {
	mut         sync.RWMutex
	clockOffset time.Duration
	syncPeriod  time.Duration
}

func New(syncPeriod time.Duration) SyncTime {
	s := SyncTime{clockOffset: 0, syncPeriod: syncPeriod}
	go s.synchronize()
	return s
}

func (s *SyncTime) synchronize() {
	for {
		r, err := ntp.Query("time.google.com")

		s.mut.Lock()

		if err != nil {
			s.clockOffset = 0
		} else {
			s.clockOffset = r.ClockOffset
		}

		s.mut.Unlock()
		time.Sleep(s.syncPeriod)
	}
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

func (s *SyncTime) GetCurrentTime() time.Time {
	return time.Now().Add(s.GetClockOffset())
}
