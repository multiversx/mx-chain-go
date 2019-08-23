package round

import (
	"math"
	"time"

	"github.com/ElrondNetwork/elrond-go/ntp"
)

// round defines the data needed by the rounder
type round struct {
	index        int64         // represents the index of the round in the current chronology (current time - genesis time) / round duration
	timeStamp    time.Time     // represents the start time of the round in the current chronology genesis time + round index * round duration
	timeDuration time.Duration // represents the duration of the round in current chronology
	syncTimer    ntp.SyncTimer
}

// NewRound defines a new round object
func NewRound(
	genesisTimeStamp time.Time,
	currentTimeStamp time.Time,
	roundTimeDuration time.Duration,
	syncTimer ntp.SyncTimer,
) (*round, error) {

	if syncTimer == nil {
		return nil, ErrNilSyncTimer
	}

	rnd := round{timeDuration: roundTimeDuration, timeStamp: genesisTimeStamp, syncTimer: syncTimer}
	rnd.UpdateRound(genesisTimeStamp, currentTimeStamp)
	return &rnd, nil
}

// UpdateRound updates the index and the time stamp of the round depending of the genesis time and the current time given
func (rnd *round) UpdateRound(genesisTimeStamp time.Time, currentTimeStamp time.Time) {
	delta := currentTimeStamp.Sub(genesisTimeStamp).Nanoseconds()

	index := int64(math.Floor(float64(delta) / float64(rnd.timeDuration.Nanoseconds())))

	if rnd.index != index {
		rnd.index = index
		rnd.timeStamp = genesisTimeStamp.Add(time.Duration(int64(index) * rnd.timeDuration.Nanoseconds()))
	}
}

// Index returns the index of the round in current epoch
func (rnd *round) Index() int64 {
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

// RemainingTime returns the remaining time in the current round given by the current time, round start time and
// safe threshold percent
func (rnd *round) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	currentTime := rnd.syncTimer.CurrentTime()
	elapsedTime := currentTime.Sub(startTime)
	remainingTime := maxTime - elapsedTime

	return remainingTime
}

// IsInterfaceNil returns true if there is no value under the interface
func (rnd *round) IsInterfaceNil() bool {
	if rnd == nil {
		return true
	}
	return false
}
