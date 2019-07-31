package mock

import (
	"math"
	"time"
)

type RounderMock struct {
	RoundIndex        int64
	RoundTimeStamp    time.Time
	RoundTimeDuration time.Duration
}

func (rndm *RounderMock) Index() int64 {
	return rndm.RoundIndex
}

func (rndm *RounderMock) TimeDuration() time.Duration {
	return rndm.RoundTimeDuration
}

func (rndm *RounderMock) TimeStamp() time.Time {
	return rndm.RoundTimeStamp
}

func (rndm *RounderMock) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	index := int64(math.Floor(float64(delta) / float64(rndm.RoundTimeDuration.Nanoseconds())))

	if rndm.RoundIndex != index {
		rndm.RoundIndex = index
		rndm.RoundTimeStamp = genesisRoundTimeStamp.Add(time.Duration(int64(index) * rndm.RoundTimeDuration.Nanoseconds()))
	}
}

func (rndm *RounderMock) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	return rndm.RoundTimeDuration
}
