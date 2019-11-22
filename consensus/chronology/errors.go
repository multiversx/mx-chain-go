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

// ErrNilEpochStartTrigger is raised when a valid start of epoch trigger is expected but nil used
var ErrNilEpochStartTrigger = errors.New("nil start of epoch trigger")
