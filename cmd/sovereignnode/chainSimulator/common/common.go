package common

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	sovRunType "github.com/multiversx/mx-chain-go/sovereignnode/runType"
)

// GetCurrentSovereignHeader returns current sovereign chain block handler from blockchain hook
func GetCurrentSovereignHeader(nodeHandler process.NodeHandler) data.SovereignChainHeaderHandler {
	return nodeHandler.GetChainHandler().GetCurrentBlockHeader().(data.SovereignChainHeaderHandler)
}

// CreateSovereignRunTypeComponents will create sovereign run type components
func CreateSovereignRunTypeComponents(args runType.ArgsRunTypeComponents, sovereignExtraConfig config.SovereignConfig) (factory.RunTypeComponentsHolder, error) {
	argsSovRunType, err := sovRunType.CreateSovereignArgsRunTypeComponents(args, sovereignExtraConfig)
	if err != nil {
		return nil, err
	}

	sovereignComponentsFactory, err := runType.NewSovereignRunTypeComponentsFactory(*argsSovRunType)
	if err != nil {
		return nil, err
	}

	managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(sovereignComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedRunTypeComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedRunTypeComponents, nil
}
