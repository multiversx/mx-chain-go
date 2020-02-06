package mock

import "time"

// RounderMock -
type RounderMock struct {
	IndexField         int64
	TimeStampField     time.Time
	TimeDurationField  time.Duration
	RemainingTimeField time.Duration
}

// Index -
func (rm *RounderMock) Index() int64 {
	return rm.IndexField
}

// UpdateRound -
func (rm *RounderMock) UpdateRound(time.Time, time.Time) {
}

// TimeStamp -
func (rm *RounderMock) TimeStamp() time.Time {
	return rm.TimeStampField
}

// TimeDuration -
func (rm *RounderMock) TimeDuration() time.Duration {
	return rm.TimeDurationField
}

// RemainingTime -
func (rm *RounderMock) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	return rm.RemainingTimeField
}

// IsInterfaceNil -
func (rm *RounderMock) IsInterfaceNil() bool {
	if rm == nil {
		return true
	}
	return false
}
