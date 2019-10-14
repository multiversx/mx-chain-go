package mock

import (
	"sync"
	"time"
)

type MockTimer struct {
	seconds  int64
	mutTimer *sync.Mutex
}

func NewMockTimer() *MockTimer {
	return &MockTimer{
		seconds:  0,
		mutTimer: &sync.Mutex{},
	}
}

func (m *MockTimer) Now() time.Time {
	m.mutTimer.Lock()
	defer m.mutTimer.Unlock()
	return time.Unix(m.seconds, 0)
}

func (m *MockTimer) IsInterfaceNil() bool {
	if m == nil {
		return true
	}
	return false
}

func (m *MockTimer) IncrementSeconds(value int) {
	m.mutTimer.Lock()
	defer m.mutTimer.Unlock()
	m.seconds += int64(value)
}

func (m *MockTimer) SetSeconds(value int) {
	m.mutTimer.Lock()
	defer m.mutTimer.Unlock()
	m.seconds = int64(value)
}
