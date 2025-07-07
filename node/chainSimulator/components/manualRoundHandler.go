package components

import (
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/round"
)

type manualRoundHandler struct {
	consensus.RoundHandler
	index            int64
	genesisTimeStamp int64
	roundDuration    time.Duration
	initialRound     int64
}

// NewManualRoundHandler returns a manual round handler instance
func NewManualRoundHandler(genesisTimeStamp int64, roundDuration time.Duration, initialRound int64) *manualRoundHandler {
	roundArgs := round.ArgsRound{
		GenesisTimeStamp:          genesisTimeStamp,
		SupernovaGenesisTimeStamp: genesisTimeStamp,
		CurrentTimeStamp:          genesisTimeStamp, // not used here
		RoundTimeDuration:         roundDuration,
		SupernovaTimeDuration:     0,
		SyncTimer:                 nil,
		StartRound:                0,
		SupernovaStartRound:       0,
		EnableEpochsHandler:       nil,
		EnableRoundsHandler:       nil,
	}
	roundHandler, err := round.NewRound()

	return &manualRoundHandler{
		genesisTimeStamp: genesisTimeStamp,
		roundDuration:    roundDuration,
		index:            initialRound,
		initialRound:     initialRound,
	}
}

// IncrementIndex will increment the current round index
func (handler *manualRoundHandler) IncrementIndex() {
	atomic.AddInt64(&handler.index, 1)
}

// RevertOneRound -
func (handler *manualRoundHandler) RevertOneRound() {
	atomic.AddInt64(&handler.index, -1)
}

// Index returns the current index
func (handler *manualRoundHandler) Index() int64 {
	return atomic.LoadInt64(&handler.index)
}

// BeforeGenesis returns false
func (handler *manualRoundHandler) BeforeGenesis() bool {
	return false
}

// UpdateRound does nothing as this implementation does not work with real timers
func (handler *manualRoundHandler) UpdateRound(_ time.Time, _ time.Time) {
}

// TimeStamp returns the time based of the genesis timestamp and the current round
func (handler *manualRoundHandler) TimeStamp() time.Time {
	rounds := atomic.LoadInt64(&handler.index)
	timeFromGenesis := handler.roundDuration * time.Duration(rounds)
	timestamp := time.UnixMilli(handler.genesisTimeStamp).Add(timeFromGenesis)
	timestamp = time.UnixMilli(timestamp.UnixMilli() - int64(handler.roundDuration.Milliseconds())*handler.initialRound)
	return timestamp
}

// TimeDuration returns the provided time duration for this instance
func (handler *manualRoundHandler) TimeDuration() time.Duration {
	return handler.roundDuration
}

// RemainingTime returns the max time as the start time is not taken into account
func (handler *manualRoundHandler) RemainingTime(_ time.Time, maxTime time.Duration) time.Duration {
	return maxTime
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *manualRoundHandler) IsInterfaceNil() bool {
	return handler == nil
}
