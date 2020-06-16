package mock

import (
	"time"
)

// RounderMock -
type RounderMock struct {
	index int64

	IndexCalled         func() int64
	TimeDurationCalled  func() time.Duration
	TimeStampCalled     func() time.Time
	UpdateRoundCalled   func(time.Time, time.Time)
	RemainingTimeCalled func(startTime time.Time, maxTime time.Duration) time.Duration
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
func (rndm *RounderMock) Index() int64 {
	if rndm.IndexCalled != nil {
		return rndm.IndexCalled()
	}

	return rndm.index
}

// TimeDuration -
func (rndm *RounderMock) TimeDuration() time.Duration {
	if rndm.TimeDurationCalled != nil {
		return rndm.TimeDurationCalled()
	}

	return 4000 * time.Millisecond
}

// TimeStamp -
func (rndm *RounderMock) TimeStamp() time.Time {
	if rndm.TimeStampCalled != nil {
		return rndm.TimeStampCalled()
	}

	return time.Unix(0, 0)
}

// UpdateRound -
func (rndm *RounderMock) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	if rndm.UpdateRoundCalled != nil {
		rndm.UpdateRoundCalled(genesisRoundTimeStamp, timeStamp)
		return
	}

	rndm.index++
}

// RemainingTime -
func (rndm *RounderMock) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	if rndm.RemainingTimeCalled != nil {
		return rndm.RemainingTimeCalled(startTime, maxTime)
	}

	return 4000 * time.Millisecond
}

// IsInterfaceNil returns true if there is no value under the interface
func (rndm *RounderMock) IsInterfaceNil() bool {
	return rndm == nil
}
