package facade

import "github.com/pkg/errors"

// ErrHeartbeatsNotActive signals that the heartbeat system is not active
var ErrHeartbeatsNotActive = errors.New("heartbeat system not active")
