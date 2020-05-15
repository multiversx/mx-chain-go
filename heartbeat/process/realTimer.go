package process

import "time"

// RealTimer is an implementation of Timer and uses real time.now
type RealTimer struct {
}

// Now returns the time.Now() Time
func (rt *RealTimer) Now() time.Time {
	return time.Now()
}

// IsInterfaceNil verifies if the interface is nil
func (rt *RealTimer) IsInterfaceNil() bool {
	return rt == nil
}
