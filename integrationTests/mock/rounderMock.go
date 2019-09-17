package mock

import "time"

type RounderMock struct {
	IndexField         int64
	TimeStampField     time.Time
	TimeDurationField  time.Duration
	RemainingTimeField time.Duration
}

func (rm *RounderMock) Index() int64 {
	return rm.IndexField
}

func (rm *RounderMock) UpdateRound(time.Time, time.Time) {
}

func (rm *RounderMock) TimeStamp() time.Time {
	return rm.TimeStampField
}

func (rm *RounderMock) TimeDuration() time.Duration {
	return rm.TimeDurationField
}

func (rm *RounderMock) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	return rm.RemainingTimeField
}

func (rm *RounderMock) IsInterfaceNil() bool {
	if rm == nil {
		return true
	}
	return false
}
