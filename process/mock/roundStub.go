package mock

import (
	"time"
)

type RoundStub struct {
	IndexCalled                func() int32
	TimeDurationCalled         func() time.Duration
	TimeStampCalled            func() time.Time
	UpdateRoundCalled          func(time.Time, time.Time)
	RemainingTimeInRoundCalled func(uint32) time.Duration
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

func (rnds *RoundStub) RemainingTimeInRound(safeThresholdPercent uint32) time.Duration {
	return rnds.RemainingTimeInRoundCalled(safeThresholdPercent)
}
