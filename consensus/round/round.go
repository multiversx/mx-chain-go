package round

import (
	"math"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/ntp"
)

var log = logger.GetOrCreate("consensus/round")

var _ consensus.RoundHandler = (*round)(nil)

// round defines the data needed by the roundHandler
type round struct {
	index               int64         // represents the index of the round in the current chronology (current time - genesis time) / round duration
	timeStamp           time.Time     // represents the start time of the round in the current chronology genesis time + round index * round duration
	newGenesisTimeStamp time.Time     // time duration between genesis and the time duration change
	timeDuration        time.Duration // represents the duration of the round in current chronology
	newTimeDuration     time.Duration // represents the duration of the round in current chronology
	syncTimer           ntp.SyncTimer
	startRound          int64
	newStartRound       int64

	*sync.RWMutex

	enableRoundsHandler common.EnableRoundsHandler
}

// NewRound defines a new round object
func NewRound(
	genesisTimeStamp time.Time,
	currentTimeStamp time.Time,
	roundTimeDuration time.Duration,
	syncTimer ntp.SyncTimer,
	startRound int64,
	enableRoundsHandler common.EnableRoundsHandler,
) (*round, error) {
	log.Debug("creating round handler..")

	if check.IfNil(syncTimer) {
		return nil, ErrNilSyncTimer
	}
	if check.IfNil(enableRoundsHandler) {
		return nil, errors.ErrNilEnableRoundsHandler
	}

	newStartRound := int64(enableRoundsHandler.SupernovaActivationRound())
	newGenesisTimeStamp := genesisTimeStamp.Add(time.Duration(newStartRound * roundTimeDuration.Nanoseconds()))
	newTimeDuration := time.Duration(4000) * time.Millisecond

	rnd := round{
		timeDuration:        roundTimeDuration,
		newTimeDuration:     newTimeDuration,
		timeStamp:           genesisTimeStamp,
		newGenesisTimeStamp: newGenesisTimeStamp,
		syncTimer:           syncTimer,
		startRound:          startRound,
		newStartRound:       newStartRound,
		RWMutex:             &sync.RWMutex{},
		enableRoundsHandler: enableRoundsHandler,
	}
	rnd.UpdateRound(genesisTimeStamp, currentTimeStamp)

	log.Debug("updated initial round..")

	return &rnd, nil
}

// UpdateRound updates the index and the time stamp of the round depending on the genesis time and the current time given
func (rnd *round) UpdateRound(genesisTimeStamp time.Time, currentTimeStamp time.Time) {
	// TODO: determine if there is any issue here on how genesisTimeStamp is passed at transitions, related to initial granularity

	// log.Debug("UpdateRound",
	// 	"genesisTimeStamp", genesisTimeStamp,
	// 	"genesisTimeStamp unix", genesisTimeStamp.Unix(),
	// 	"genesisTimeStamp unix milli", genesisTimeStamp.UnixMilli(),
	// 	"currentTimeStamp", currentTimeStamp,
	// 	"currentTimeStamp unix", currentTimeStamp.Unix(),
	// 	"currentTimeStamp unix milli", currentTimeStamp.UnixMilli(),
	// 	"old rnd.timeStamp", rnd.timeStamp.UnixMilli(),
	// )

	if !rnd.enableRoundsHandler.SupernovaEnableRoundEnabled() {
		rnd.updateRoundLegacy(genesisTimeStamp, currentTimeStamp)
		return
	}

	delta := currentTimeStamp.Sub(rnd.newGenesisTimeStamp).Nanoseconds()

	startRound := rnd.startRound + rnd.newStartRound

	index := int64(math.Floor(float64(delta)/float64(rnd.getTimeDuration().Nanoseconds()))) + startRound

	rnd.Lock()
	if rnd.index != index {
		rnd.index = index
		rnd.timeStamp = rnd.newGenesisTimeStamp.Add(time.Duration((index - startRound) * rnd.getTimeDuration().Nanoseconds()))
	}
	rnd.Unlock()

	log.Trace("UpdateRound",
		"index", index,
		"newGenesisTimeStamp milli", rnd.newGenesisTimeStamp.UnixMilli(),
		"rnd.timeStamp unix milli", rnd.timeStamp.UnixMilli(),
	)
}

func (rnd *round) updateRoundLegacy(genesisTimeStamp time.Time, currentTimeStamp time.Time) {
	delta := currentTimeStamp.Sub(genesisTimeStamp).Nanoseconds()

	index := int64(math.Floor(float64(delta)/float64(rnd.getTimeDuration().Nanoseconds()))) + rnd.startRound

	rnd.Lock()
	if rnd.index != index {
		rnd.index = index
		rnd.timeStamp = genesisTimeStamp.Add(time.Duration((index - rnd.startRound) * rnd.getTimeDuration().Nanoseconds()))
	}

	// log.Debug("updateRoundLegacy",
	// 	"delta", delta,
	// 	"genesisTimeStamp.UnixMilli", genesisTimeStamp.UnixMilli(),
	// 	"genesisTimeStamp.Unix", genesisTimeStamp.Unix(),
	// 	"currentTimeStamp.UnixMilli", currentTimeStamp.UnixMilli(),
	// 	"currentTimeStamp.Unix", currentTimeStamp.Unix(),
	// 	"index", index,
	// )

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
	rnd.RLock()
	defer rnd.RUnlock()

	return rnd.getTimeDuration()
}

// this should be called under mutex protection
func (rnd *round) getTimeDuration() time.Duration {
	if rnd.enableRoundsHandler.SupernovaEnableRoundEnabled() {
		return rnd.newTimeDuration
	}

	return rnd.timeDuration
}

// SetTimeDuration sets the duration of the round
func (rnd *round) SetTimeDuration(timeDuration time.Duration) {
}

func (rnd *round) SetNewTimeStamp(genesisTimeStamp time.Time, currentTimeStamp time.Time) {
}

// RemainingTime returns the remaining time in the current round given by the current time, round start time and
// safe threshold percent
func (rnd *round) RemainingTime(startTime time.Time, maxTime time.Duration) time.Duration {
	currentTime := rnd.syncTimer.CurrentTime()
	elapsedTime := currentTime.Sub(startTime)
	remainingTime := maxTime - elapsedTime

	return remainingTime
}

// RevertOneRound reverts the round index and time stamp by one round, used in case of a transition to new consensus
// TODO: handle revert for supernova new timestamp field
func (rnd *round) RevertOneRound() {
	rnd.Lock()

	rnd.index--
	rnd.timeStamp = rnd.timeStamp.Add(-rnd.getTimeDuration())
	// index := rnd.index - 1
	// timeStamp := rnd.timeStamp.Add(-rnd.getTimeDuration())

	log.Debug("RevertOneRound fake",
		"index", rnd.index,
		"timeStamp milli", rnd.timeStamp.UnixMilli(),
	)

	rnd.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rnd *round) IsInterfaceNil() bool {
	return rnd == nil
}
