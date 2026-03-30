package mock

import (
	"time"
)

// RoundStub -
type RoundStub struct {
	IndexCalled                func() int64
	TimeDurationCalled         func() time.Duration
	TimeStampCalled            func() time.Time
	UpdateRoundCalled          func(time.Time, time.Time)
	RemainingTimeCalled        func(time.Time, time.Duration) time.Duration
	GetTimeStampForRoundCalled func(round uint64) uint64
}

// Index -
func (rnds *RoundStub) Index() int64 {
	return rnds.IndexCalled()
}

// TimeDuration -
func (rnds *RoundStub) TimeDuration() time.Duration {
	return rnds.TimeDurationCalled()
}

// TimeStamp -
func (rnds *RoundStub) TimeStamp() time.Time {
	return rnds.TimeStampCalled()
}

// UpdateRound -
func (rnds *RoundStub) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	rnds.UpdateRoundCalled(genesisRoundTimeStamp, timeStamp)
}

// RemainingTime -
func (rnds *RoundStub) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	return rnds.RemainingTimeCalled(startTime, maxTime)
}

// GetTimeStampForRound -
func (rnds *RoundStub) GetTimeStampForRound(round uint64) uint64 {
	if rnds.GetTimeStampForRoundCalled != nil {
		return rnds.GetTimeStampForRoundCalled(round)
	}

	return uint64(time.Unix(0, 0).UnixMilli())
}

// IsInterfaceNil --
func (rnds *RoundStub) IsInterfaceNil() bool {
	return rnds == nil
}
