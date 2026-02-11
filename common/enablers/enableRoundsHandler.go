package enablers

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

const (
	conversionBase    = 10
	bitConversionSize = 64
)

type roundFlag struct {
	round   uint64
	options []string
}

type flagEnabledInRound = func(round uint64) bool

type roundFlagHandler struct {
	isActiveInRound flagEnabledInRound
	activationRound uint64
}

type enableRoundsHandler struct {
	allFlagsDefined map[common.EnableRoundFlag]roundFlagHandler
	currentRound    uint64
	roundMut        sync.RWMutex
}

// NewEnableRoundsHandler creates a new enable rounds handler instance
func NewEnableRoundsHandler(roundsConfig config.RoundConfig, roundNotifier process.RoundNotifier) (*enableRoundsHandler, error) {
	handler := &enableRoundsHandler{}
	err := handler.createAllFlagsMap(roundsConfig)
	if err != nil {
		return nil, err
	}

	roundNotifier.RegisterNotifyHandler(handler)

	return handler, nil
}

func getRoundConfig(args config.RoundConfig, configName string) (*roundFlag, error) {
	activationRound, found := args.RoundActivations[configName]
	if !found {
		return nil, fmt.Errorf("%w for config %s", errMissingRoundActivation, configName)
	}

	round, err := strconv.ParseUint(activationRound.Round, conversionBase, bitConversionSize)
	if err != nil {
		return nil, fmt.Errorf("%w while trying to convert %s for the round config %s",
			err, activationRound.Round, configName)
	}

	log.Debug("loaded round config",
		"name", configName, "round", round,
		"options", fmt.Sprintf("[%s]", strings.Join(activationRound.Options, ", ")))

	return &roundFlag{
		round:   round,
		options: activationRound.Options,
	}, nil
}

func (handler *enableRoundsHandler) createAllFlagsMap(conf config.RoundConfig) error {
	disableAsyncCallV1, err := getRoundConfig(conf, string(common.DisableAsyncCallV1Flag))
	if err != nil {
		return fmt.Errorf("%w while trying to get round config for %s", err, string(common.DisableAsyncCallV1Flag))
	}

	supernovaEnableRound, err := getRoundConfig(conf, string(common.SupernovaRoundFlag))
	if err != nil {
		return fmt.Errorf("%w while trying to get round config for %s", err, string(common.SupernovaRoundFlag))
	}

	handler.allFlagsDefined = map[common.EnableRoundFlag]roundFlagHandler{
		common.DisableAsyncCallV1Flag: {
			isActiveInRound: func(round uint64) bool {
				return round >= disableAsyncCallV1.round
			},
			activationRound: disableAsyncCallV1.round,
		},
		common.SupernovaRoundFlag: {
			isActiveInRound: func(round uint64) bool {
				return round >= supernovaEnableRound.round
			},
			activationRound: supernovaEnableRound.round,
		},
	}

	return nil
}

// RoundConfirmed is called whenever a new round is confirmed
func (handler *enableRoundsHandler) RoundConfirmed(round uint64, _ uint64) {
	handler.roundMut.Lock()
	handler.currentRound = round
	handler.roundMut.Unlock()
}

// GetCurrentRound returns the current round
func (handler *enableRoundsHandler) GetCurrentRound() uint64 {
	handler.roundMut.RLock()
	currentRound := handler.currentRound
	handler.roundMut.RUnlock()

	return currentRound
}

// IsFlagDefined returns true if the provided flag is enabled in the current round
func (handler *enableRoundsHandler) IsFlagDefined(flag common.EnableRoundFlag) bool {
	_, found := handler.allFlagsDefined[flag]
	if found {
		return true
	}

	log.Warn("flag is not defined",
		"flag", flag,
		"stack trace", string(debug.Stack()),
	)

	return false
}

// IsFlagEnabled returns true if the provided flag is enabled in the current round
func (handler *enableRoundsHandler) IsFlagEnabled(flag common.EnableRoundFlag) bool {
	handler.roundMut.RLock()
	currentRound := handler.currentRound
	handler.roundMut.RUnlock()

	return handler.IsFlagEnabledInRound(flag, currentRound)
}

// IsFlagEnabledInRound returns true if the provided flag is enabled in the provided round
func (handler *enableRoundsHandler) IsFlagEnabledInRound(flag common.EnableRoundFlag, round uint64) bool {
	fh, found := handler.allFlagsDefined[flag]
	if !found {
		log.Warn("IsFlagEnabledInRound: got unknown flag",
			"flag", flag,
			"round", round,
			"stack trace", string(debug.Stack()),
		)

		return false
	}

	return fh.isActiveInRound(round)
}

// GetActivationRound returns the activation round of the provided flag
func (handler *enableRoundsHandler) GetActivationRound(flag common.EnableRoundFlag) uint64 {
	fh, found := handler.allFlagsDefined[flag]
	if !found {
		log.Warn("GetActivationRound: got unknown flag",
			"flag", flag,
			"stack trace", string(debug.Stack()),
		)

		return 0
	}

	return fh.activationRound
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *enableRoundsHandler) IsInterfaceNil() bool {
	return handler == nil
}
