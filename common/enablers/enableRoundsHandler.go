package enablers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

const (
	conversionBase    = 10
	bitConversionSize = 64

	disableAsyncCallV1 = "DisableAsyncCallV1"
)

type enableRoundsHandler struct {
	*roundFlagsHolder
}

// NewEnableRoundsHandler creates a new enable rounds handler instance
func NewEnableRoundsHandler(args config.RoundConfig, roundNotifier process.RoundNotifier) (*enableRoundsHandler, error) {
	disableAsyncCallV1, err := getRoundConfig(args, disableAsyncCallV1)
	if err != nil {
		return nil, err
	}

	handler := &enableRoundsHandler{
		roundFlagsHolder: &roundFlagsHolder{
			disableAsyncCallV1: disableAsyncCallV1,
		},
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
		Flag:    &atomic.Flag{},
		round:   round,
		options: activationRound.Options,
	}, nil
}

// RoundConfirmed is called whenever a new round is confirmed
func (handler *enableRoundsHandler) RoundConfirmed(round uint64, timestamp uint64) {
	handler.disableAsyncCallV1.SetValue(handler.disableAsyncCallV1.round <= round)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *enableRoundsHandler) IsInterfaceNil() bool {
	return handler == nil
}
