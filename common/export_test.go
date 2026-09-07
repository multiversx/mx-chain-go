package common

import "time"

// MinRoundDurationMS -
var MinRoundDurationMS = minRoundDurationMS

// MinRoundDurationSec -
var MinRoundDurationSec = minRoundDurationSec

// NewTimeoutHandlerWithHandlerFunc -
func NewTimeoutHandlerWithHandlerFunc(timeout time.Duration, handler func() time.Time) *timeoutHandler {
	th := &timeoutHandler{
		timeoutValue:   timeout,
		getTimeHandler: handler,
	}
	th.checkpoint = th.getTimeHandler()

	return th
}

// TimeDurationToUnix -
func TimeDurationToUnix(
	duration time.Duration,
	enableEpochsHandler EnableEpochsHandler,
	epoch uint32,
) int64 {
	return timeDurationToUnix(duration, enableEpochsHandler, epoch)
}
