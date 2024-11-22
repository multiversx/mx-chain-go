package runType

import (
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
)

type sovereignRunTypeCoreComponentsFactory struct {
}

// NewSovereignRunTypeCoreComponentsFactory will return a new instance of sovereign runType core components factory
func NewSovereignRunTypeCoreComponentsFactory() *sovereignRunTypeCoreComponentsFactory {
	return &sovereignRunTypeCoreComponentsFactory{}
}

// Create will return a new instance of runType core components
func (srccf *sovereignRunTypeCoreComponentsFactory) Create() *runTypeCoreComponents {
	return &runTypeCoreComponents{
		genesisNodesSetupFactory: sharding.NewSovereignGenesisNodesSetupFactory(),
		ratingsDataFactory:       rating.NewSovereignRatingsDataFactory(),
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (srccf *sovereignRunTypeCoreComponentsFactory) IsInterfaceNil() bool {
	return srccf == nil
}
