package mock

import (
	"time"
)

// RoundHandlerStub -
type RoundHandlerStub struct {
	RoundIndex int64

	IndexCalled         func() int64
	TimeDurationCalled  func() time.Duration
	TimeStampCalled     func() time.Time
	UpdateRoundCalled   func(time.Time, time.Time)
	RemainingTimeCalled func(startTime time.Time, maxTime time.Duration) time.Duration
}

// Index -
func (rndm *RoundHandlerStub) Index() int64 {
	if rndm.IndexCalled != nil {
		return rndm.IndexCalled()
	}

	return rndm.RoundIndex
}

// TimeDuration -
func (rndm *RoundHandlerStub) TimeDuration() time.Duration {
	if rndm.TimeDurationCalled != nil {
		return rndm.TimeDurationCalled()
	}

	return 4000 * time.Millisecond
}

// TimeStamp -
func (rndm *RoundHandlerStub) TimeStamp() time.Time {
	if rndm.TimeStampCalled != nil {
		return rndm.TimeStampCalled()
	}

	return time.Unix(0, 0)
}

// UpdateRound -
func (rndm *RoundHandlerStub) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	if rndm.UpdateRoundCalled != nil {
		rndm.UpdateRoundCalled(genesisRoundTimeStamp, timeStamp)
		return
	}

	rndm.RoundIndex++
}

// RemainingTime -
func (rndm *RoundHandlerStub) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	if rndm.RemainingTimeCalled != nil {
		return rndm.RemainingTimeCalled(startTime, maxTime)
	}

	return 4000 * time.Millisecond
}

// IsInterfaceNil returns true if there is no value under the interface
func (rndm *RoundHandlerStub) IsInterfaceNil() bool {
	return rndm == nil
}
