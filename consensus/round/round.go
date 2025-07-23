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

// ArgsRound defines the arguments needed to create a new round handler component
type ArgsRound struct {
	GenesisTimeStamp          time.Time
	SupernovaGenesisTimeStamp time.Time
	CurrentTimeStamp          time.Time
	RoundTimeDuration         time.Duration
	SupernovaTimeDuration     time.Duration
	SyncTimer                 ntp.SyncTimer
	StartRound                int64
	SupernovaStartRound       int64
	EnableRoundsHandler       common.EnableRoundsHandler
}

// round defines the data needed by the roundHandler
type round struct {
	index                     int64         // represents the index of the round in the current chronology (current time - genesis time) / round duration
	timeStamp                 time.Time     // represents the start time of the round in the current chronology genesis time + round index * round duration
	supernovaGenesisTimeStamp time.Time     // time duration between genesis and the time duration change
	timeDuration              time.Duration // represents the duration of the round in current chronology
	supernovaTimeDuration     time.Duration
	syncTimer                 ntp.SyncTimer
	startRound                int64
	supernovaStartRound       int64

	*sync.RWMutex

	enableRoundsHandler common.EnableRoundsHandler
}

// NewRound defines a new round object
func NewRound(args ArgsRound) (*round, error) {
	log.Debug("creating round handler..")

	if check.IfNil(args.SyncTimer) {
		return nil, ErrNilSyncTimer
	}
	if check.IfNil(args.EnableRoundsHandler) {
		return nil, errors.ErrNilEnableRoundsHandler
	}

	rnd := round{
		timeDuration:              args.RoundTimeDuration,
		supernovaTimeDuration:     args.SupernovaTimeDuration,
		timeStamp:                 args.GenesisTimeStamp,
		supernovaGenesisTimeStamp: args.SupernovaGenesisTimeStamp,
		syncTimer:                 args.SyncTimer,
		startRound:                args.StartRound,
		supernovaStartRound:       args.SupernovaStartRound,
		RWMutex:                   &sync.RWMutex{},
		enableRoundsHandler:       args.EnableRoundsHandler,
	}
	rnd.UpdateRound(args.GenesisTimeStamp, args.CurrentTimeStamp)

	log.Debug("updated initial round..")

	return &rnd, nil
}

// UpdateRound updates the index and the time stamp of the round depending on the genesis time and the current time given
func (rnd *round) UpdateRound(genesisTimeStamp time.Time, currentTimeStamp time.Time) {
	baseTimeStamp := rnd.supernovaGenesisTimeStamp
	roundDuration := rnd.supernovaTimeDuration
	startRound := rnd.supernovaStartRound

	supernovaActivated := rnd.isSupernovaActivated(currentTimeStamp)

	if !supernovaActivated {
		baseTimeStamp = genesisTimeStamp
		roundDuration = rnd.timeDuration
		startRound = rnd.startRound
	}

	rnd.updateRound(baseTimeStamp, currentTimeStamp, startRound, roundDuration)
}

func (rnd *round) isSupernovaRoundActivated() bool {
	rnd.RLock()
	index := rnd.index
	rnd.RUnlock()

	if index < 0 {
		return false
	}

	return rnd.enableRoundsHandler.IsFlagEnabledInRound(common.SupernovaRoundFlag, uint64(index))
}

func (rnd *round) isSupernovaActivated(currentTimeStamp time.Time) bool {
	supernovaActivated := rnd.isSupernovaRoundActivated()
	if supernovaActivated {
		return supernovaActivated
	}

	currentTimeAfterSupernova := currentTimeStamp.UnixMilli() >= rnd.supernovaGenesisTimeStamp.UnixMilli()

	if currentTimeAfterSupernova {
		log.Debug("isSupernovaActivated: force set supernovaActivated",
			"currentTimeAfterSupernova", currentTimeAfterSupernova,
		)
		supernovaActivated = true
	}

	return supernovaActivated
}

func (rnd *round) updateRound(
	genesisTimeStamp time.Time,
	currentTimeStamp time.Time,
	startRound int64,
	roundDuration time.Duration,
) {
	delta := currentTimeStamp.Sub(genesisTimeStamp).Nanoseconds()

	index := int64(math.Floor(float64(delta)/float64(roundDuration.Nanoseconds()))) + startRound

	log.Debug("round.updateRound",
		"delta", delta,
		"index", index,
		"rnd.timeStamp", rnd.timeStamp.UnixMilli(),
		"currentTimeStamp", currentTimeStamp.UnixMilli(),
	)

	rnd.Lock()
	if rnd.index != index {
		rnd.index = index
		rnd.timeStamp = genesisTimeStamp.Add(time.Duration((index - startRound) * roundDuration.Nanoseconds()))
	}

	log.Trace("round.updateRound",
		"delta", delta,
		"index", index,
		"startRound", startRound,
		"genesisTimeStamp", genesisTimeStamp.UnixMilli(),
		"rnd.timeStamp", rnd.timeStamp.UnixMilli(),
		"currentTimeStamp", currentTimeStamp.UnixMilli(),
	)

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
	return rnd.getTimeDuration()
}

func (rnd *round) getTimeDuration() time.Duration {
	if rnd.isSupernovaRoundActivated() {
		return rnd.supernovaTimeDuration
	}

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

// RevertOneRound reverts the round index and time stamp by one round, used in case of a transition to new consensus
func (rnd *round) RevertOneRound() {
	timeDuration := rnd.getTimeDuration()

	rnd.Lock()

	rnd.index--
	rnd.timeStamp = rnd.timeStamp.Add(-timeDuration)

	rnd.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rnd *round) IsInterfaceNil() bool {
	return rnd == nil
}
