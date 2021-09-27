package testscommon

import (
	"sync"
	"time"
)

type timeoutHandlerMock struct {
	timeoutValue  time.Duration
	mutCheckpoint sync.RWMutex
	checkpoint    time.Time
}

// NewTimeoutHandlerMock -
func NewTimeoutHandlerMock(timeout time.Duration) *timeoutHandlerMock {
	return &timeoutHandlerMock{
		timeoutValue: timeout,
		checkpoint:   time.Now(),
	}
}

// ResetWatchdog -
func (thm *timeoutHandlerMock) ResetWatchdog() {
	thm.mutCheckpoint.Lock()
	defer thm.mutCheckpoint.Unlock()

	thm.checkpoint = time.Now()
}

// IsTimeout returns true if the timeoutValue has been reached
func (thm *timeoutHandlerMock) IsTimeout() bool {
	thm.mutCheckpoint.RLock()
	defer thm.mutCheckpoint.RUnlock()

	now := time.Now()

	return now.Sub(thm.checkpoint) > thm.timeoutValue
}

// IsInterfaceNil returns true if there is no value under the interface
func (thm *timeoutHandlerMock) IsInterfaceNil() bool {
	return thm == nil
}
