package runType

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/sovereignnode/dataCodec"
	"github.com/multiversx/mx-chain-go/sovereignnode/incomingHeader"

	"github.com/multiversx/mx-sdk-abi-incubator/golang/abi"
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

	codec := abi.NewDefaultCodec()
	argsDataCodec := dataCodec.ArgsDataCodec{
		Serializer: abi.NewSerializer(codec),
	}

	dataCodecHandler, err := dataCodec.NewDataCodec(argsDataCodec)
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
