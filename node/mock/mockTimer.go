package mock

import (
	"sync"
	"time"
)

type MockTimer struct {
	secondsMutex sync.RWMutex
	seconds      int64
}

func (m *MockTimer) Now() time.Time {
	m.secondsMutex.RLock()
	currentSeconds := time.Unix(m.seconds, 0)
	m.secondsMutex.RUnlock()
	return currentSeconds
}

func (m *MockTimer) IsInterfaceNil() bool {
	if m == nil {
		return true
	}
	return false
}

func (m *MockTimer) IncrementSeconds(value int) {
	m.secondsMutex.Lock()
	m.seconds += int64(value)
	m.secondsMutex.Unlock()
}

func (m *MockTimer) SetSeconds(value int) {
	m.secondsMutex.Lock()
	m.seconds = int64(value)
	m.secondsMutex.Unlock()
}
