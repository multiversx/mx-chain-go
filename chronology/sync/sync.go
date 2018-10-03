package sync

import (
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
