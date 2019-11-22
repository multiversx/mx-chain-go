package mock

import (
	"time"
)

type RounderStub struct {
	RoundIndex int64

	IndexCalled         func() int64
	TimeDurationCalled  func() time.Duration
	TimeStampCalled     func() time.Time
	UpdateRoundCalled   func(time.Time, time.Time)
	RemainingTimeCalled func(startTime time.Time, maxTime time.Duration) time.Duration
}

func (rndm *RounderStub) Index() int64 {
	if rndm.IndexCalled != nil {
		return rndm.IndexCalled()
	}

	return rndm.RoundIndex
}

func (rndm *RounderStub) TimeDuration() time.Duration {
	if rndm.TimeDurationCalled != nil {
		return rndm.TimeDurationCalled()
	}

	return time.Duration(4000 * time.Millisecond)
}

func (rndm *RounderStub) TimeStamp() time.Time {
	if rndm.TimeStampCalled != nil {
		return rndm.TimeStampCalled()
	}

	return time.Unix(0, 0)
}

func (rndm *RounderStub) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	if rndm.UpdateRoundCalled != nil {
		rndm.UpdateRoundCalled(genesisRoundTimeStamp, timeStamp)
		return
	}

	rndm.RoundIndex++
}

func (rndm *RounderStub) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	if rndm.RemainingTimeCalled != nil {
		return rndm.RemainingTimeCalled(startTime, maxTime)
	}

	return time.Duration(4000 * time.Millisecond)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rndm *RounderStub) IsInterfaceNil() bool {
	if rndm == nil {
		return true
	}
	return false
}
