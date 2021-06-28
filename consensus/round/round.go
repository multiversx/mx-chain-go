package round

import (
	"math"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/ntp"
)

var _ consensus.RoundHandler = (*round)(nil)

// round defines the data needed by the roundHandler
type round struct {
	index        int64         // represents the index of the round in the current chronology (current time - genesis time) / round duration
	timeStamp    time.Time     // represents the start time of the round in the current chronology genesis time + round index * round duration
	timeDuration time.Duration // represents the duration of the round in current chronology
	syncTimer    ntp.SyncTimer
	startRound   int64

	*sync.RWMutex
}

// NewRound defines a new round object
func NewRound(
	genesisTimeStamp time.Time,
	currentTimeStamp time.Time,
	roundTimeDuration time.Duration,
	syncTimer ntp.SyncTimer,
	startRound int64,
) (*round, error) {

	if check.IfNil(syncTimer) {
		return nil, ErrNilSyncTimer
	}

	rnd := round{
		timeDuration: roundTimeDuration,
		timeStamp:    genesisTimeStamp,
		syncTimer:    syncTimer,
		startRound:   startRound,
		RWMutex:      &sync.RWMutex{},
	}
	rnd.UpdateRound(genesisTimeStamp, currentTimeStamp)
	return &rnd, nil
}

// UpdateRound updates the index and the time stamp of the round depending of the genesis time and the current time given
func (rnd *round) UpdateRound(genesisTimeStamp time.Time, currentTimeStamp time.Time) {
	delta := currentTimeStamp.Sub(genesisTimeStamp).Nanoseconds()

	index := int64(math.Floor(float64(delta)/float64(rnd.timeDuration.Nanoseconds()))) + rnd.startRound

	rnd.Lock()
	if rnd.index != index {
		rnd.index = index
		rnd.timeStamp = genesisTimeStamp.Add(time.Duration((index - rnd.startRound) * rnd.timeDuration.Nanoseconds()))
	}
	rnd.Unlock()
}

// Index returns the index of the round in current epoch
func (rnd *round) Index() int64 {
	rnd.RLock()
	defer rnd.RUnlock()

	return rnd.index
}

// BeforeGenesis returns true if round index is before start round
func (rnd *round) BeforeGenesis() bool {
	rnd.RLock()
	defer rnd.RUnlock()

	return rnd.index <= rnd.startRound
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
	return rnd == nil
}
