package chronology

import (
	"errors"
)

// ErrNilRoundHandler is raised when a valid roundHandler is expected but nil used
var ErrNilRoundHandler = errors.New("roundHandler is nil")

// ErrNilSyncTimer is raised when a valid sync timer is expected but nil used
var ErrNilSyncTimer = errors.New("sync timer is nil")

// ErrNilAppStatusHandler is raised when the AppStatusHandler is nil when setting it
var ErrNilAppStatusHandler = errors.New("nil AppStatusHandler")

// ErrNilWatchdog signals that a nil watchdog has been provided
var ErrNilWatchdog = errors.New("nil watchdog")
