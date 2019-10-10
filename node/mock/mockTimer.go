package mock

import "time"

type MockTimer struct {
	seconds int64
}

func (m *MockTimer) Now() time.Time {
	return time.Unix(m.seconds, 0)
}

func (m *MockTimer) IsInterfaceNil() bool {
	if m == nil {
		return true
	}
	return false
}

func (m *MockTimer) IncrementSeconds(value int) {
	m.seconds += int64(value)
}

func (m *MockTimer) SetSeconds(value int) {
	m.seconds = int64(value)
}
