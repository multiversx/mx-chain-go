package external

import "github.com/pkg/errors"

// ErrNilSCQueryService signals that a nil SC query service has been provided
var ErrNilSCQueryService = errors.New("nil SC query service")

// ErrNilStatusMetrics signals that a nil status metrics was provided
var ErrNilStatusMetrics = errors.New("nil status metrics handler")
