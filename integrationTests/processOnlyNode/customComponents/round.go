package customComponents

import (
	"sync"
	"time"
)

// round defines the data needed by the roundHandler
type round struct {
	index        int64         // represents the index of the round in the current chronology (current time - genesis time) / round duration
	timeStamp    time.Time     // represents the start time of the round in the current chronology genesis time + round index * round duration
	timeDuration time.Duration // represents the duration of the round in current chronology

	*sync.RWMutex
}

// NewRound defines a new round object
func NewRound(
	genesisTimeStamp time.Time,
	roundTimeDuration time.Duration,
) *round {

	rnd := &round{
		timeDuration: roundTimeDuration,
		timeStamp:    genesisTimeStamp,
		RWMutex:      &sync.RWMutex{},
	}
	return rnd
}

// UpdateRound does nothing
func (rnd *round) UpdateRound(_ time.Time, _ time.Time) {}

// Index returns the index of the round in current epoch
func (rnd *round) Index() int64 {
	rnd.RLock()
	defer rnd.RUnlock()

	return rnd.index
}

// BeforeGenesis returns false, all calls are considered to be done after the genesis
func (rnd *round) BeforeGenesis() bool {
	return false
}

// TimeStamp returns the time stamp of the round
func (rnd *round) TimeStamp() time.Time {
	rnd.RLock()
	defer rnd.RUnlock()

	return rnd.timeStamp
}

// TimeDuration returns the duration of the round
func (rnd *round) TimeDuration() time.Duration {
	return rnd.timeDuration
}

// RemainingTime hardcoded to return the round time
func (rnd *round) RemainingTime(_ time.Time, _ time.Duration) time.Duration {
	return rnd.timeDuration
}

// IncrementRound will increment the round & add the round time duration to the current timestamp
func (rnd *round) IncrementRound() {
	rnd.Lock()
	defer rnd.Unlock()

	rnd.index++
	rnd.timeStamp = rnd.timeStamp.Add(rnd.timeDuration)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rnd *round) IsInterfaceNil() bool {
	return rnd == nil
}
