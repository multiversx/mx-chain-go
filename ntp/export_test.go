package ntp

import (
	"time"

	"github.com/beevik/ntp"
)

// OutOfBoundsDuration -
const OutOfBoundsDuration = outOfBoundsDuration

// Query -
func (s *syncTime) Query() func(options NTPOptions, hostIndex int) (*ntp.Response, error) {
	return s.query
}

// SyncPeriod -
func (s *syncTime) SyncPeriod() time.Duration {
	return s.syncPeriod
}

// SetClockOffset -
func (s *syncTime) SetClockOffset(clockOffset time.Duration) {
	s.setClockOffset(clockOffset)
}

// Sync -
func (s *syncTime) Sync() {
	s.sync()
}

// GetClockOffsetsWithoutEdges -
func (s *syncTime) GetClockOffsetsWithoutEdges(clockOffsets []time.Duration) []time.Duration {
	return s.getClockOffsetsWithoutEdges(clockOffsets)
}

// GetHarmonicMean -
func (s *syncTime) GetHarmonicMean(clockOffsets []time.Duration) time.Duration {
	return s.getHarmonicMean(clockOffsets)
}

// GetSleepTime -
func (s *syncTime) GetSleepTime() time.Duration {
	return s.getSleepTime()
}
