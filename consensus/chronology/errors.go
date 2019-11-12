package chronology

import (
	"errors"
)

// ErrNilRounder is raised when a valid rounder is expected but nil used
var ErrNilRounder = errors.New("rounder is nil")

// ErrNilSyncTimer is raised when a valid sync timer is expected but nil used
var ErrNilSyncTimer = errors.New("sync timer is nil")

// ErrNilAppStatusHandler is raised when the AppStatusHandler is nil when setting it
var ErrNilAppStatusHandler = errors.New("nil AppStatusHandler")

// ErrNilEndOfEpochTrigger is raised when a valid end of epoch trigger is expected but nil used
var ErrNilEndOfEpochTrigger = errors.New("nil end of epoch trigger")
