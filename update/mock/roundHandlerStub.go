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
func (rhs *RoundHandlerStub) Index() int64 {
	if rhs.IndexCalled != nil {
		return rhs.IndexCalled()
	}

	return rhs.RoundIndex
}

// TimeDuration -
func (rhs *RoundHandlerStub) TimeDuration() time.Duration {
	if rhs.TimeDurationCalled != nil {
		return rhs.TimeDurationCalled()
	}

	return 4000 * time.Millisecond
}

// TimeStamp -
func (rhs *RoundHandlerStub) TimeStamp() time.Time {
	if rhs.TimeStampCalled != nil {
		return rhs.TimeStampCalled()
	}

	return time.Unix(0, 0)
}

// UpdateRound -
func (rhs *RoundHandlerStub) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	if rhs.UpdateRoundCalled != nil {
		rhs.UpdateRoundCalled(genesisRoundTimeStamp, timeStamp)
		return
	}

	rhs.RoundIndex++
}

// RemainingTime -
func (rhs *RoundHandlerStub) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	if rhs.RemainingTimeCalled != nil {
		return rhs.RemainingTimeCalled(startTime, maxTime)
	}

	return 4000 * time.Millisecond
}

// IsInterfaceNil returns true if there is no value under the interface
func (rhs *RoundHandlerStub) IsInterfaceNil() bool {
	return rhs == nil
}
