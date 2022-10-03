package statistics

import (
	"errors"
)

// ErrNilNetworkStatistics signals that a nil network statistics was provided
var ErrNilNetworkStatistics = errors.New("nil network statistics")

// ErrInvalidRefreshIntervalValue signals that an invalid value for the refresh interval was provided
var ErrInvalidRefreshIntervalValue = errors.New("invalid refresh interval value")
