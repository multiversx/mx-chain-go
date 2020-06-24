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

// ErrNilAlarmScheduler is raised when a valid alarm scheduler is expected but nil is used
var ErrNilAlarmScheduler = errors.New("nil alarm scheduler")

// ErrNilEndProcessChan is raised when a valid end process chan is expected but nil is used
var ErrNilEndProcessChan = errors.New("nil end process chan")
