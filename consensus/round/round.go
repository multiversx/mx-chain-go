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
	EnableEpochsHandler       common.EnableEpochsHandler
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

	enableEpochsHandler common.EnableEpochsHandler
	enableRoundsHandler common.EnableRoundsHandler
}

// NewRound defines a new round object
func NewRound(args ArgsRound) (*round, error) {
	log.Debug("creating round handler..")

	if check.IfNil(args.SyncTimer) {
		return nil, ErrNilSyncTimer
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
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
		enableEpochsHandler:       args.EnableEpochsHandler,
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

func (rnd *round) isSupernovaActivated(currentTimeStamp time.Time) bool {
	supernovaActivated := common.IsSupernovaRoundActivated(rnd.enableEpochsHandler, rnd.enableRoundsHandler)
	if supernovaActivated {
		return supernovaActivated
	}

	currentTimeAfterSupernova := currentTimeStamp.UnixMilli() > rnd.supernovaGenesisTimeStamp.UnixMilli()
	defaultEpoch := rnd.enableEpochsHandler.GetCurrentEpoch() == 0
	defaultRound := rnd.enableRoundsHandler.GetCurrentRound() == 0

	if currentTimeAfterSupernova && (defaultEpoch || defaultRound) {
		log.Debug("isSupernovaActivated: force set supernovaActivated",
			"currentTimeAfterSupernova", currentTimeAfterSupernova,
			"enableEpochsHandler current epoch", rnd.enableEpochsHandler.GetCurrentEpoch(),
			"enableRoundsHandler current round", rnd.enableRoundsHandler.GetCurrentRound(),
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

	rnd.Lock()
	if rnd.index != index {
		rnd.index = index
		rnd.timeStamp = genesisTimeStamp.Add(time.Duration((index - startRound) * roundDuration.Nanoseconds()))
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
	rnd.RLock()
	defer rnd.RUnlock()

	return rnd.getTimeDuration()
}

// this should be called under mutex protection
func (rnd *round) getTimeDuration() time.Duration {
	// TODO: analysize adding here also forced activation based on current timestamp
	if common.IsSupernovaRoundActivated(rnd.enableEpochsHandler, rnd.enableRoundsHandler) {
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
	rnd.Lock()

	rnd.index--
	rnd.timeStamp = rnd.timeStamp.Add(-rnd.getTimeDuration())

	rnd.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rnd *round) IsInterfaceNil() bool {
	return rnd == nil
}
