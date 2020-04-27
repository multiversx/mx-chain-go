package ntp

import (
	"time"

	"github.com/beevik/ntp"
)

func (s *syncTime) Query() func(options NTPOptions, hostIndex int) (*ntp.Response, error) {
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

func (s *syncTime) GetClockOffsetsWithoutEdges(clockOffsets []time.Duration) []time.Duration {
	return s.getClockOffsetsWithoutEdges(clockOffsets)
}

func (s *syncTime) GetHarmonicMean(clockOffsets []time.Duration) time.Duration {
	return s.getHarmonicMean(clockOffsets)
}

func (s *syncTime) GetSleepTime() time.Duration {
	return s.getSleepTime()
}
