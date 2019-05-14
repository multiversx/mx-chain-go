package mock

import (
	"time"
)

type RounderMock struct {
	RoundIndex int32

	IndexCalled         func() int32
	TimeDurationCalled  func() time.Duration
	TimeStampCalled     func() time.Time
	UpdateRoundCalled   func(time.Time, time.Time)
	RemainingTimeCalled func(startTime time.Time, maxTime time.Duration) time.Duration
}

func (rndm *RounderMock) Index() int32 {
	if rndm.IndexCalled != nil {
		return rndm.IndexCalled()
	}

	return rndm.RoundIndex
}

func (rndm *RounderMock) TimeDuration() time.Duration {
	if rndm.TimeDurationCalled != nil {
		return rndm.TimeDurationCalled()
	}

	return time.Duration(4000 * time.Millisecond)
}

func (rndm *RounderMock) TimeStamp() time.Time {
	if rndm.TimeStampCalled != nil {
		return rndm.TimeStampCalled()
	}

	return time.Unix(0, 0)
}

func (rndm *RounderMock) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	if rndm.UpdateRoundCalled != nil {
		rndm.UpdateRoundCalled(genesisRoundTimeStamp, timeStamp)
		return
	}

	rndm.RoundIndex++
}

func (rndm *RounderMock) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	if rndm.RemainingTimeCalled != nil {
		return rndm.RemainingTimeCalled(startTime, maxTime)
	}

	return time.Duration(4000 * time.Millisecond)
}
