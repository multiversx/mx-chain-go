package watchdog

import "errors"

// ErrNilAlarmScheduler is raised when a valid alarm scheduler is expected but nil is used
var ErrNilAlarmScheduler = errors.New("nil alarm scheduler")

// ErrNilEndProcessChan is raised when a valid end process chan is expected but nil is used
var ErrNilEndProcessChan = errors.New("nil end process chan")
