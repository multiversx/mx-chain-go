package enablers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go/config"
)

// TODO uncomment all from this file and from the test file and replace Example with something else
// const exampleName = "Example"

type enableRoundsHandler struct {
	*roundFlagsHolder
}

// NewEnableRoundsHandler creates a new enable rounds handler instance
func NewEnableRoundsHandler(args config.RoundConfig) (*enableRoundsHandler, error) {
	// example, err := getRoundConfig(args, exampleName)
	// if err != nil {
	//	return nil, err
	// }

	return &enableRoundsHandler{
		roundFlagsHolder: &roundFlagsHolder{
			//	example: example,
		},
	}, nil
}

func getRoundConfig(args config.RoundConfig, configName string) (*roundFlag, error) {
	activationRound, found := args.RoundActivations[configName]
	if !found {
		return nil, fmt.Errorf("%w for config %s", errMissingRoundActivation, configName)
	}

	return &roundFlag{
		Flag:    &atomic.Flag{},
		round:   activationRound.Round,
		options: activationRound.Options,
	}, nil
}

// CheckRound should be called whenever a new round is known. It will trigger the updating of all containing round flags
func (handler *enableRoundsHandler) CheckRound(round uint64) {
	// handler.example.SetValue(handler.example.round <= round)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *enableRoundsHandler) IsInterfaceNil() bool {
	return handler == nil
}
