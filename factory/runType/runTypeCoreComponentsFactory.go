package runType

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
)

type runTypeCoreComponentsFactory struct {
	epochConfig config.EpochConfig
}

// NewRunTypeCoreComponentsFactory will return a new instance of runTypeCoreComponentsFactory
func NewRunTypeCoreComponentsFactory(epochConfig config.EpochConfig) *runTypeCoreComponentsFactory {
	return &runTypeCoreComponentsFactory{
		epochConfig: epochConfig,
	}
}

// Create will create the runType core components
func (rccf *runTypeCoreComponentsFactory) Create() (*runTypeCoreComponents, error) {
	epochNotifier := forking.NewGenericEpochNotifier()
	enableEpochsHandler, err := enablers.NewEnableEpochsHandler(rccf.epochConfig.EnableEpochs, epochNotifier)
	if err != nil {
		return nil, fmt.Errorf("runTypeCoreComponentsFactory - NewEnableEpochsHandler failed: %w", err)
	}

	return &runTypeCoreComponents{
		genesisNodesSetupFactory: sharding.NewGenesisNodesSetupFactory(),
		ratingsDataFactory:       rating.NewRatingsDataFactory(),
		enableEpochsHandler:      enableEpochsHandler,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rccf *runTypeCoreComponentsFactory) IsInterfaceNil() bool {
	return rccf == nil
}
