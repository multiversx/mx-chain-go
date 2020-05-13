package ntp

import (
	"io"
	"time"
)

// SyncTimer defines an interface for time synchronization
type SyncTimer interface {
	io.Closer
	StartSync()
	ClockOffset() time.Duration
	FormattedCurrentTime() string
	CurrentTime() time.Time
	IsInterfaceNil() bool
}
