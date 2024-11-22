package runType

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
)

type sovereignRunTypeCoreComponentsFactory struct {
	*runTypeCoreComponentsFactory
}

// NewSovereignRunTypeCoreComponentsFactory will return a new instance of sovereign runType core components factory
func NewSovereignRunTypeCoreComponentsFactory(rtcc *runTypeCoreComponentsFactory) (*sovereignRunTypeCoreComponentsFactory, error) {
	if check.IfNil(rtcc) {
		return nil, errorsMx.ErrNilRunTypeCoreComponentsFactory
	}

	return &sovereignRunTypeCoreComponentsFactory{
		runTypeCoreComponentsFactory: rtcc,
	}, nil
}

// Create will return a new instance of runType core components
func (srccf *sovereignRunTypeCoreComponentsFactory) Create() (*runTypeCoreComponents, error) {
	epochNotifier := forking.NewGenericEpochNotifier()
	sovEnableEpochsHandler, err := enablers.NewSovereignEnableEpochsHandler(srccf.epochConfig.EnableEpochs, srccf.epochConfig.SovereignEnableEpochs, srccf.epochConfig.SovereignChainSpecificEnableEpochs, epochNotifier)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeCoreComponentsFactory - NewSovereignEnableEpochsHandler failed: %w", err)
	}

	return &runTypeCoreComponents{
		genesisNodesSetupFactory: sharding.NewSovereignGenesisNodesSetupFactory(),
		ratingsDataFactory:       rating.NewSovereignRatingsDataFactory(),
		enableEpochsHandler:      sovEnableEpochsHandler,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (srccf *sovereignRunTypeCoreComponentsFactory) IsInterfaceNil() bool {
	return srccf == nil
}
