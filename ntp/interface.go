package ntp

import (
	"time"
)

// SyncTimer defines an interface for time synchronization
type SyncTimer interface {
	Close() error
	ForceSync()
	StartSyncingTime()
	ClockOffset() time.Duration
	FormattedCurrentTime() string
	CurrentTime() time.Time
	IsInterfaceNil() bool
}
