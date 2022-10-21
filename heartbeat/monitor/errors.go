package monitor

import "errors"

var (
	errEmptyHeartbeatMessagesInstance = errors.New("programming error: empty heartbeatMessages instance")
	errInconsistentActivePidsList     = errors.New("programming error: inconsistent active pids list")
	errInconsistentActiveMap          = errors.New("programming error: inconsistent active map")
)
