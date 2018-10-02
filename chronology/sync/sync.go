package sync

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"sync"
	"time"
)

type SyncTime struct {
	mut         sync.RWMutex
	clockOffset time.Duration
}

func New() SyncTime {
	s := SyncTime{}
	go s.synchronize()
	return s
}

func (s *SyncTime) synchronize() {
	for {
		time.Sleep(100 * time.Millisecond)
		r, err := ntp.Query("time.google.com")

		s.mut.Lock()

		if err != nil {
			s.clockOffset = 0
		} else {
			s.clockOffset = r.ClockOffset
		}

		s.mut.Unlock()
	}
}

func (s *SyncTime) GetClockOffset() time.Duration {
	s.mut.RLock()
	clockOffset := s.clockOffset
	s.mut.RUnlock()

	return clockOffset
}
