package heartbeat

import "time"

// RealTimer is an implementation of Timer and uses real time.now
type RealTimer struct {
}

// Now returns the time.Now() Time
func (m *RealTimer) Now() time.Time {
	return time.Now()
}

// IsInterfaceNil verifies if the interface is nil
func (m *RealTimer) IsInterfaceNil() bool {
	if m == nil {
		return true
	}
	return false
}
