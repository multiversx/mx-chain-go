package runType

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader"
	"github.com/multiversx/mx-chain-go/sovereignnode/dataCodec"
	"github.com/multiversx/mx-sdk-abi-go/abi"
)

const (
	separator = "@"
)

// CreateSovereignArgsRunTypeComponents creates the args for run type component
func CreateSovereignArgsRunTypeComponents(
	argsRunType runType.ArgsRunTypeComponents,
	configs config.SovereignConfig,
) (*runType.ArgsSovereignRunTypeComponents, error) {
	runTypeComponentsFactory, err := runType.NewRunTypeComponentsFactory(argsRunType)
	if err != nil {
		return nil, fmt.Errorf("NewRunTypeComponentsFactory failed: %w", err)
	}

	serializer, err := abi.NewSerializer(abi.ArgsNewSerializer{
		PartsSeparator: separator,
	})
	if err != nil {
		return nil, err
	}

	dataCodecHandler, err := dataCodec.NewDataCodec(serializer)
	if err != nil {
		return nil, err
	}

	return &runType.ArgsSovereignRunTypeComponents{
		RunTypeComponentsFactory: runTypeComponentsFactory,
		Config:                   configs,
		DataCodec:                dataCodecHandler,
		TopicsChecker:            incomingHeader.NewTopicsChecker(),
	}, nil
}
