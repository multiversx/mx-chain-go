package heartbeat

import "time"

type RealTimer struct {
}

func (m *RealTimer) Now() time.Time {
	return time.Now()
}

func (m *RealTimer) IsInterfaceNil() bool {
	if m == nil {
		return true
	}
	return false
}
