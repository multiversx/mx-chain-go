package statistics

import (
	"errors"
)

// ErrNilNetworkStatisticsProvider signals that a nil network statistics provider was provided
var ErrNilNetworkStatisticsProvider = errors.New("nil network statistics provider")

// ErrInvalidRefreshIntervalValue signals that an invalid value for the refresh interval was provided
var ErrInvalidRefreshIntervalValue = errors.New("invalid refresh interval value")

// ErrNilStateStatsHandler signals that a nil state statistics handler was provided
var ErrNilStateStatsHandler = errors.New("nil state statistics handler")
