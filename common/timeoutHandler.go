package common

import (
	"fmt"
	"sync"
	"time"
)

const minTimeoutNodesReceived = time.Second

type timeoutHandler struct {
	timeoutValue   time.Duration
	mutCheckpoint  sync.RWMutex
	checkpoint     time.Time
	getTimeHandler func() time.Time
}

// NewTimeoutHandler returns a new instance of the timeout handler
func NewTimeoutHandler(timeout time.Duration) (*timeoutHandler, error) {
	if timeout < minTimeoutNodesReceived {
		return nil, fmt.Errorf("%w provided: %v, minimum %v",
			ErrInvalidTimeout, timeout, minTimeoutNodesReceived)
	}

	th := &timeoutHandler{
		timeoutValue: timeout,
	}

	th.getTimeHandler = func() time.Time {
		return time.Now()
	}
	th.checkpoint = th.getTimeHandler()

	return th, nil
}

// ResetWatchdog resets the current time value
func (th *timeoutHandler) ResetWatchdog() {
	th.mutCheckpoint.Lock()
	defer th.mutCheckpoint.Unlock()

	th.checkpoint = th.getTimeHandler()
}

// IsTimeout returns true if the timeoutValue has been reached
func (th *timeoutHandler) IsTimeout() bool {
	th.mutCheckpoint.RLock()
	defer th.mutCheckpoint.RUnlock()

	now := th.getTimeHandler()

	return now.Sub(th.checkpoint) > th.timeoutValue
}

// IsInterfaceNil returns true if there is no value under the interface
func (th *timeoutHandler) IsInterfaceNil() bool {
	return th == nil
}
