package components

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

var errNilEnableRoundsHandler = errors.New("nil enable rounds handler")

// ArgManualRoundHandler is the DTO used to create a new instance of manualRoundHandler
type ArgManualRoundHandler struct {
	EnableRoundsHandler       common.EnableRoundsHandler
	GenesisTimeStamp          int64
	SupernovaGenesisTimeStamp int64
	RoundDuration             time.Duration
	SupernovaRoundDuration    time.Duration
	InitialRound              int64
}

type manualRoundHandler struct {
	index                     int64
	genesisTimeStamp          int64
	supernovaGenesisTimeStamp int64
	roundDuration             time.Duration
	supernovaRoundDuration    time.Duration
	initialRound              int64
	enableRoundsHandler       common.EnableRoundsHandler
}

// NewManualRoundHandler returns a manual round handler instance
func NewManualRoundHandler(args ArgManualRoundHandler) (*manualRoundHandler, error) {
	if check.IfNil(args.EnableRoundsHandler) {
		return nil, errNilEnableRoundsHandler
	}
	return &manualRoundHandler{
		genesisTimeStamp:          args.GenesisTimeStamp,
		roundDuration:             args.RoundDuration,
		index:                     args.InitialRound,
		initialRound:              args.InitialRound,
		enableRoundsHandler:       args.EnableRoundsHandler,
		supernovaGenesisTimeStamp: args.SupernovaGenesisTimeStamp,
		supernovaRoundDuration:    args.SupernovaRoundDuration,
	}, nil
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
	roundIndex := atomic.LoadInt64(&handler.index)
	if !handler.isSupernovaActive(roundIndex) {
		timeFromGenesis := handler.roundDuration * time.Duration(roundIndex)
		timestamp := time.UnixMilli(handler.genesisTimeStamp).Add(timeFromGenesis)
		timestamp = time.UnixMilli(timestamp.UnixMilli() - handler.roundDuration.Milliseconds()*handler.initialRound)
		return timestamp
	}

	supernovaGenesisRound := handler.enableRoundsHandler.GetActivationRound(common.SupernovaRoundFlag)
	roundsDiff := uint64(roundIndex) - supernovaGenesisRound
	timeFromSupernovaGenesis := handler.supernovaRoundDuration * time.Duration(roundsDiff)
	timestamp := time.UnixMilli(handler.supernovaGenesisTimeStamp).Add(timeFromSupernovaGenesis)
	return timestamp
}

func (handler *manualRoundHandler) isSupernovaActive(roundIndex int64) bool {
	return handler.enableRoundsHandler.IsFlagEnabledInRound(common.SupernovaRoundFlag, uint64(roundIndex))
}

// TimeDuration returns the provided time duration for this instance
func (handler *manualRoundHandler) TimeDuration() time.Duration {
	roundIndex := atomic.LoadInt64(&handler.index)
	if !handler.isSupernovaActive(roundIndex) {
		return handler.roundDuration
	}

	return handler.supernovaRoundDuration
}

// RemainingTime returns the max time as the start time is not taken into account
func (handler *manualRoundHandler) RemainingTime(_ time.Time, maxTime time.Duration) time.Duration {
	return maxTime
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *manualRoundHandler) IsInterfaceNil() bool {
	return handler == nil
}
