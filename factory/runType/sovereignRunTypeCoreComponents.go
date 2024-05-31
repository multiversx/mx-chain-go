package runType

import (
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type sovereignRunTypeCoreComponentsFactory struct {
	*runTypeCoreComponentsFactory
}

// NewSovereignRunTypeCoreComponentsFactory will return a new instance of sovereign runTypeCoreComponentsFactory
func NewSovereignRunTypeCoreComponentsFactory(rccf *runTypeCoreComponentsFactory) (*sovereignRunTypeCoreComponentsFactory, error) {
	if check.IfNil(rccf) {
		return nil, errors.ErrNilRunTypeCoreComponentsFactory
	}

	return &sovereignRunTypeCoreComponentsFactory{
		runTypeCoreComponentsFactory: rccf,
	}, nil
}

func (srccf *sovereignRunTypeCoreComponentsFactory) Create() *runTypeCoreComponents {
	return &runTypeCoreComponents{
		genesisNodesSetupFactory: sharding.NewSovereignGenesisNodesSetupFactory(),
		ratingsDataFactory:       rating.NewSovereignRatingsDataFactory(),
	}
}
