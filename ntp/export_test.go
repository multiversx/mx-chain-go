package ntp

import (
	"time"

	"github.com/beevik/ntp"
)

func (s *syncTime) Query() func(host string) (*ntp.Response, error) {
	return s.query
}

func (s *syncTime) SyncPeriod() time.Duration {
	return s.syncPeriod
}

func (s *syncTime) SetClockOffset(clockOffset time.Duration) {
	s.setClockOffset(clockOffset)
}

func (s *syncTime) Sync() {
	s.sync()
}
