package ntp

import (
	"time"
)

// SyncTimer defines an interface for time synchronization
type SyncTimer interface {
	StartSync()
	ClockOffset() time.Duration
	FormattedCurrentTime() string
	CurrentTime() time.Time
}
