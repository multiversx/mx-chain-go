package ntp

import (
	"time"
)

// SyncTimer defines an interface for time synchronization
type SyncTimer interface {
	Close() error
	StartSyncingTime()
	ClockOffset() time.Duration
	FormattedCurrentTime() string
	CurrentTime() time.Time
	IsInterfaceNil() bool
}

// SyncTimerWithRoundHandler represents an entity able to tell the current time
type SyncTimerWithRoundHandler interface {
	SetRoundHandler(handler RoundHandler)
	IsInterfaceNil() bool
}

// RoundHandler defines the actions which should be handled by a round implementation
type RoundHandler interface {
	// UpdateRound updates the index and the time stamp of the round depending on the genesis time and the current time given
	UpdateRound(time.Time, time.Time)
	IsInterfaceNil() bool
}
