package common

import "time"

// NewTimeoutHandlerWithHandlerFunc -
func NewTimeoutHandlerWithHandlerFunc(timeout time.Duration, handler func() time.Time) *timeoutHandler {
	th := &timeoutHandler{
		timeoutValue:   timeout,
		getTimeHandler: handler,
	}
	th.checkpoint = th.getTimeHandler()

	return th
}
