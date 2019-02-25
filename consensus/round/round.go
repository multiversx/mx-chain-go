package round

import (
	"math"
	"time"
)

// round defines the data needed by the rounder
type round struct {
	index        int32         // represents the index of the round in the current chronology (current time - genesis time) / round duration
	timeStamp    time.Time     // represents the start time of the round in the current chronology genesis time + round index * round duration
	timeDuration time.Duration // represents the duration of the round in current chronology
}

// NewRound defines a new round object
func NewRound(genesisTimeStamp time.Time, currentTimeStamp time.Time, roundTimeDuration time.Duration) *round {
	rnd := round{timeDuration: roundTimeDuration, timeStamp: genesisTimeStamp}
	rnd.UpdateRound(genesisTimeStamp, currentTimeStamp)
	return &rnd
}

// UpdateRound updates the index and the time stamp of the round depending of the genesis time and the current time given
func (rnd *round) UpdateRound(genesisTimeStamp time.Time, currentTimeStamp time.Time) {
	delta := currentTimeStamp.Sub(genesisTimeStamp).Nanoseconds()

	index := int32(math.Floor(float64(delta) / float64(rnd.timeDuration.Nanoseconds())))

	if rnd.index != index {
		rnd.index = index
		rnd.timeStamp = genesisTimeStamp.Add(time.Duration(int64(index) * rnd.timeDuration.Nanoseconds()))
	}
}

// Index returns the index of the round in current epoch
func (rnd *round) Index() int32 {
	return rnd.index
}

// TimeStamp returns the time stamp of the round
func (rnd *round) TimeStamp() time.Time {
	return rnd.timeStamp
}

// TimeDuration returns the duration of the round
func (rnd *round) TimeDuration() time.Duration {
	return rnd.timeDuration
}

// RemainingTimeInRound returns the remaining time in the current round given by the current time, round start time and
// safe threshold percent
func (rnd *round) RemainingTimeInRound(currentTime time.Time, safeThresholdPercent uint32) time.Duration {
	roundStartTime := rnd.timeStamp
	elapsedTime := currentTime.Sub(roundStartTime)
	remainingTime := rnd.timeDuration*time.Duration(safeThresholdPercent)/100 - elapsedTime

	return remainingTime
}
