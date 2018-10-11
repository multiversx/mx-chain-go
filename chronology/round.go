package chronology

import (
	"time"

	"github.com/davecgh/go-spew/spew"
)

type Round struct {
	index        int
	timeStamp    time.Time
	timeDuration time.Duration
}

func NewRound(genesisRoundTimeStamp time.Time, timeStamp time.Time, roundTimeDuration time.Duration) Round {
	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	rnd := Round{}

	rnd.index = int(delta / roundTimeDuration.Nanoseconds())
	rnd.timeStamp = genesisRoundTimeStamp.Add(time.Duration(int64(rnd.index) * roundTimeDuration.Nanoseconds()))
	rnd.timeDuration = roundTimeDuration

	return rnd
}

func (rnd *Round) UpdateRound(genesisRoundTimeStamp time.Time, timeStamp time.Time) {
	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	index := int(delta / rnd.timeDuration.Nanoseconds())

	if rnd.index != index {
		rnd.index = index
		rnd.timeStamp = genesisRoundTimeStamp.Add(time.Duration(int64(index) * rnd.timeDuration.Nanoseconds()))
	}
}

func (rnd *Round) Print() {
	spew.Dump(rnd)
}
