package chronology

import (
	"time"

	"github.com/davecgh/go-spew/spew"
)

// Round defines the data needed by the round
type Round struct {
	index        int
	timeStamp    time.Time
	timeDuration time.Duration
}

// NewRound defines a new Round object
func NewRound(genesisRoundTimeStamp time.Time, timeStamp time.Time, roundTimeDuration time.Duration) *Round {
	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	rnd := Round{}

	rnd.index = int(delta / roundTimeDuration.Nanoseconds())
	rnd.timeStamp = genesisRoundTimeStamp.Add(time.Duration(int64(rnd.index) * roundTimeDuration.Nanoseconds()))
	rnd.timeDuration = roundTimeDuration

	return &rnd
}

// UpdateRound updates the index and the time stamp of the round depending of the genesis time and the current time
// given
func (rnd *Round) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	index := int(delta / rnd.timeDuration.Nanoseconds())

	if rnd.index != index {
		rnd.index = index
		rnd.timeStamp = genesisRoundTimeStamp.Add(time.Duration(int64(index) * rnd.timeDuration.Nanoseconds()))
	}
}

// Index returns the index of the round in current epoch
func (rnd *Round) Index() int {
	return rnd.index
}

// TimeStamp returns the time stamp of the round
func (rnd *Round) TimeStamp() time.Time {
	return rnd.timeStamp
}

// TimeDuration returns the duration of the round
func (rnd *Round) TimeDuration() time.Duration {
	return rnd.timeDuration
}

// Print method just spew to the console the Round object in some pretty format
func (rnd *Round) Print() {
	spew.Dump(rnd)
}
