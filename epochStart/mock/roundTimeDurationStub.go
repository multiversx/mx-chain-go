package mock

import "time"

// RoundTimeDurationHandler -
type RoundTimeDurationHandler struct {
	TimeDurationCalled func() time.Duration
}

// TimeDuration -
func (r *RoundTimeDurationHandler) TimeDuration() time.Duration {
	if r.TimeDurationCalled != nil {
		return r.TimeDurationCalled()
	}

	return 4000 * time.Millisecond
}

// IsInterfaceNil -
func (r *RoundTimeDurationHandler) IsInterfaceNil() bool {
	return r == nil
}
