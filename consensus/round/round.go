package round

import (
	"math"
	"time"
)

// Round defines the data needed by the rounder
type Round struct {
	index        int32
	timeStamp    time.Time
	timeDuration time.Duration
}

// NewRound defines a new Round object
func NewRound(genesisRoundTimeStamp time.Time, timeStamp time.Time, roundTimeDuration time.Duration) *Round {
	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	rnd := Round{}

	rnd.index = int32(math.Floor(float64(delta) / float64(roundTimeDuration.Nanoseconds())))
	rnd.timeStamp = genesisRoundTimeStamp.Add(time.Duration(int64(rnd.index) * roundTimeDuration.Nanoseconds()))
	rnd.timeDuration = roundTimeDuration

	return &rnd
}

// UpdateRound updates the index and the time stamp of the rounder depending of the genesis time and the current time
// given
func (rnd *Round) UpdateRound(genesisTimeStamp time.Time, timeStamp time.Time) {
	delta := timeStamp.Sub(genesisTimeStamp).Nanoseconds()

	index := int32(math.Floor(float64(delta) / float64(rnd.timeDuration.Nanoseconds())))

	if rnd.index != index {
		rnd.index = index
		rnd.timeStamp = genesisTimeStamp.Add(time.Duration(int64(index) * rnd.timeDuration.Nanoseconds()))
	}
}

// Index returns the index of the rounder in current epoch
func (rnd *Round) Index() int32 {
	return rnd.index
}

// TimeStamp returns the time stamp of the rounder
func (rnd *Round) TimeStamp() time.Time {
	return rnd.timeStamp
}

// TimeDuration returns the duration of the rounder
func (rnd *Round) TimeDuration() time.Duration {
	return rnd.timeDuration
}
