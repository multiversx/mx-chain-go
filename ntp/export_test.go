package ntp

import (
	"time"

	"github.com/beevik/ntp"
)

// NumRequestsFromHost -
var NumRequestsFromHost = numRequestsFromHost

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

// TriggerSync -
func (s *syncTime) TriggerSync() {
	s.triggerSync()
}

// GetMedianOffset -
func (s *syncTime) GetMedianOffset(clockOffsets []time.Duration) (time.Duration, error) {
	return s.getMedianOffset(clockOffsets)
}

// GetSleepTime -
func (s *syncTime) GetSleepTime() time.Duration {
	return s.getSleepTime()
}

// SetLastSyncTime -
func (s *syncTime) SetLastSyncTime(t time.Time) {
	s.mut.Lock()
	s.lastSyncTime = t
	s.mut.Unlock()
}

// SyncCooldownDuration -
var SyncCooldownDuration = syncCooldown
