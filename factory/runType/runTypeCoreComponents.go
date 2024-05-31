package runType

import (
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
)

type runTypeCoreComponentsFactory struct {
}

type runTypeCoreComponents struct {
	genesisNodesSetupFactory sharding.GenesisNodesSetupFactory
	ratingsDataFactory       rating.RatingsDataFactory
}

// NewRunTypeCoreComponentsFactory will return a new instance of runTypeCoreComponentsFactory
func NewRunTypeCoreComponentsFactory() *runTypeCoreComponentsFactory {
	return &runTypeCoreComponentsFactory{}
}

func (rccf *runTypeCoreComponentsFactory) Create() *runTypeCoreComponents {
	return &runTypeCoreComponents{
		genesisNodesSetupFactory: sharding.NewGenesisNodesSetupFactory(),
		ratingsDataFactory:       rating.NewRatingsDataFactory(),
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (rccf *runTypeCoreComponentsFactory) IsInterfaceNil() bool {
	return rccf == nil
}

// Close does nothing
func (rcc *runTypeCoreComponents) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcc *runTypeCoreComponents) IsInterfaceNil() bool {
	return rcc == nil
}
