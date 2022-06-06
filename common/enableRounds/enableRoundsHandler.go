package enableRounds

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go/config"
)

const checkValueOnExecByCallerName = "CheckValueOnExecByCaller"

type enableRoundsHandler struct {
	*flagsHolder
}

// NewEnableRoundsHandler creates a new enable rounds handler instance
func NewEnableRoundsHandler(args config.RoundConfig) (*enableRoundsHandler, error) {
	checkValueOnExecByCaller, err := getRoundConfig(args, checkValueOnExecByCallerName)
	if err != nil {
		return nil, err
	}

	return &enableRoundsHandler{
		flagsHolder: &flagsHolder{
			checkValueOnExecByCaller: checkValueOnExecByCaller,
		},
	}, nil
}

func getRoundConfig(args config.RoundConfig, configName string) (*roundFlag, error) {
	for _, activationRound := range args.RoundActivations {
		if activationRound.Name == configName {
			return &roundFlag{
				round:   activationRound.Round,
				flag:    &atomic.Flag{},
				options: activationRound.Options,
			}, nil
		}
	}

	return nil, fmt.Errorf("%w for config %s", errMissingRoundActivation, configName)
}

// CheckRound should be called whenever a new round is known. It will trigger the updating of all containing round flags
func (handler *enableRoundsHandler) CheckRound(round uint64) {
	handler.flagsHolder.checkValueOnExecByCaller.flag.SetValue(handler.checkValueOnExecByCaller.round <= round)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *enableRoundsHandler) IsInterfaceNil() bool {
	return handler == nil
}
