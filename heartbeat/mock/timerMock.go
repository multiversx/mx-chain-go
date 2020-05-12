package mock

import (
	"sync"
	"time"
)

// TimerMock -
type TimerMock struct {
	secondsMutex *sync.RWMutex
	seconds      int64
}

// NewTimerMock -
func NewTimerMock() *TimerMock {
	return &TimerMock{
		seconds:      0,
		secondsMutex: &sync.RWMutex{},
	}
}

// Now -
func (tm *TimerMock) Now() time.Time {
	tm.secondsMutex.RLock()
	currentSeconds := time.Unix(tm.seconds, 0)
	tm.secondsMutex.RUnlock()
	return currentSeconds
}

// IsInterfaceNil -
func (tm *TimerMock) IsInterfaceNil() bool {
	return tm == nil
}

// IncrementSeconds -
func (tm *TimerMock) IncrementSeconds(value int) {
	tm.secondsMutex.Lock()
	tm.seconds += int64(value)
	tm.secondsMutex.Unlock()
}

// SetSeconds -
func (tm *TimerMock) SetSeconds(value int) {
	tm.secondsMutex.Lock()
	tm.seconds = int64(value)
	tm.secondsMutex.Unlock()
}
