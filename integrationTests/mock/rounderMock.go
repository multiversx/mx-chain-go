package mock

import "time"

// RounderMock -
type RounderMock struct {
	IndexField          int64
	TimeStampField      time.Time
	TimeDurationField   time.Duration
	RemainingTimeField  time.Duration
	BeforeGenesisCalled func() bool
}

// BeforeGenesis -
func (rndm *RounderMock) BeforeGenesis() bool {
	if rndm.BeforeGenesisCalled != nil {
		return rndm.BeforeGenesisCalled()
	}
	return false
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
func (rm *RounderMock) RemainingTime(_ time.Time, _ time.Duration) time.Duration {
	return rm.RemainingTimeField
}

// IsInterfaceNil -
func (rm *RounderMock) IsInterfaceNil() bool {
	return rm == nil
}
