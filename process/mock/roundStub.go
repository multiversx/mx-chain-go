package mock

import (
	"time"
)

type RoundStub struct {
	IndexCalled         func() int32
	TimeDurationCalled  func() time.Duration
	TimeStampCalled     func() time.Time
	UpdateRoundCalled   func(time.Time, time.Time)
	RemainingTimeCalled func(time.Time, time.Duration) time.Duration
}

func (rnds *RoundStub) Index() int32 {
	return rnds.IndexCalled()
}

func (rnds *RoundStub) TimeDuration() time.Duration {
	return rnds.TimeDurationCalled()
}

func (rnds *RoundStub) TimeStamp() time.Time {
	return rnds.TimeStampCalled()
}

func (rnds *RoundStub) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	rnds.UpdateRoundCalled(genesisRoundTimeStamp, timeStamp)
}

func (rnds *RoundStub) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	return rnds.RemainingTimeCalled(startTime, maxTime)
}
