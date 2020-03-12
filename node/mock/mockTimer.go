package mock

import (
	"sync"
	"time"
)

// MockTimer -
type MockTimer struct {
	secondsMutex *sync.RWMutex
	seconds      int64
}

// NewMockTimer -
func NewMockTimer() *MockTimer {
	return &MockTimer{
		seconds:      0,
		secondsMutex: &sync.RWMutex{},
	}
}

// Now -
func (m *MockTimer) Now() time.Time {
	m.secondsMutex.RLock()
	currentSeconds := time.Unix(m.seconds, 0)
	m.secondsMutex.RUnlock()
	return currentSeconds
}

// IsInterfaceNil -
func (m *MockTimer) IsInterfaceNil() bool {
	return m == nil
}

// IncrementSeconds -
func (m *MockTimer) IncrementSeconds(value int) {
	m.secondsMutex.Lock()
	m.seconds += int64(value)
	m.secondsMutex.Unlock()
}

// SetSeconds -
func (m *MockTimer) SetSeconds(value int) {
	m.secondsMutex.Lock()
	m.seconds = int64(value)
	m.secondsMutex.Unlock()
}
