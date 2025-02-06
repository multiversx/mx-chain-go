package runType

import (
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
)

type runTypeCoreComponentsFactory struct {
}

// NewRunTypeCoreComponentsFactory will return a new instance of runTypeCoreComponentsFactory
func NewRunTypeCoreComponentsFactory() *runTypeCoreComponentsFactory {
	return &runTypeCoreComponentsFactory{}
}

// Create will create the runType core components
func (rccf *runTypeCoreComponentsFactory) Create() *runTypeCoreComponents {
	return &runTypeCoreComponents{
		genesisNodesSetupFactory: sharding.NewGenesisNodesSetupFactory(),
		ratingsDataFactory:       rating.NewRatingsDataFactory(),
		enableEpochsFactory:      enablers.NewEnableEpochsFactory(),
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (rccf *runTypeCoreComponentsFactory) IsInterfaceNil() bool {
	return rccf == nil
}
