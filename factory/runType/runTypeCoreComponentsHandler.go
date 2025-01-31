package runType

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
)

var _ factory.ComponentHandler = (*managedRunTypeCoreComponents)(nil)
var _ factory.RunTypeCoreComponentsHandler = (*managedRunTypeCoreComponents)(nil)
var _ factory.RunTypeCoreComponentsHolder = (*managedRunTypeCoreComponents)(nil)

type managedRunTypeCoreComponents struct {
	*runTypeCoreComponents
	factory                  runTypeCoreComponentsCreator
	mutRunTypeCoreComponents sync.RWMutex
}

// NewManagedRunTypeCoreComponents returns a news instance of managed runType core components
func NewManagedRunTypeCoreComponents(rcf runTypeCoreComponentsCreator) (*managedRunTypeCoreComponents, error) {
	if rcf == nil {
		return nil, errors.ErrNilRunTypeCoreComponentsFactory
	}

	return &managedRunTypeCoreComponents{
		runTypeCoreComponents: nil,
		factory:               rcf,
	}, nil
}

// Create will create the managed components
func (mrcc *managedRunTypeCoreComponents) Create() error {
	rtc := mrcc.factory.Create()

	mrcc.mutRunTypeCoreComponents.Lock()
	mrcc.runTypeCoreComponents = rtc
	mrcc.mutRunTypeCoreComponents.Unlock()

	return nil
}

// Close will close all underlying subcomponents
func (mrcc *managedRunTypeCoreComponents) Close() error {
	mrcc.mutRunTypeCoreComponents.Lock()
	defer mrcc.mutRunTypeCoreComponents.Unlock()

	if check.IfNil(mrcc.runTypeCoreComponents) {
		return nil
	}

	err := mrcc.runTypeCoreComponents.Close()
	if err != nil {
		return err
	}
	mrcc.runTypeCoreComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mrcc *managedRunTypeCoreComponents) CheckSubcomponents() error {
	mrcc.mutRunTypeCoreComponents.RLock()
	defer mrcc.mutRunTypeCoreComponents.RUnlock()

	if check.IfNil(mrcc.runTypeCoreComponents) {
		return errors.ErrNilRunTypeCoreComponents
	}
	if check.IfNil(mrcc.genesisNodesSetupFactory) {
		return errors.ErrNilGenesisBlockFactory
	}
	if check.IfNil(mrcc.ratingsDataFactory) {
		return errors.ErrNilRatingsDataFactory
	}
	if check.IfNil(mrcc.enableEpochsFactory) {
		return enablers.ErrNilEnableEpochsFactory
	}
	return nil
}

// GenesisNodesSetupFactoryCreator returns the genesis nodes setup factory creator
func (mrcc *managedRunTypeCoreComponents) GenesisNodesSetupFactoryCreator() sharding.GenesisNodesSetupFactory {
	mrcc.mutRunTypeCoreComponents.RLock()
	defer mrcc.mutRunTypeCoreComponents.RUnlock()

	if check.IfNil(mrcc.runTypeCoreComponents) {
		return nil
	}

	return mrcc.runTypeCoreComponents.genesisNodesSetupFactory
}

// RatingsDataFactoryCreator returns the ratings data factory creator
func (mrcc *managedRunTypeCoreComponents) RatingsDataFactoryCreator() rating.RatingsDataFactory {
	mrcc.mutRunTypeCoreComponents.RLock()
	defer mrcc.mutRunTypeCoreComponents.RUnlock()

	if check.IfNil(mrcc.runTypeCoreComponents) {
		return nil
	}

	return mrcc.runTypeCoreComponents.ratingsDataFactory
}

// EnableEpochsFactoryCreator returns the enable epochs factory creator
func (mrcc *managedRunTypeCoreComponents) EnableEpochsFactoryCreator() enablers.EnableEpochsFactory {
	mrcc.mutRunTypeCoreComponents.RLock()
	defer mrcc.mutRunTypeCoreComponents.RUnlock()

	if check.IfNil(mrcc.runTypeCoreComponents) {
		return nil
	}

	return mrcc.runTypeCoreComponents.enableEpochsFactory
}

// IsInterfaceNil returns true if the interface is nil
func (mrcc *managedRunTypeCoreComponents) IsInterfaceNil() bool {
	return mrcc == nil
}

// String returns the name of the component
func (mrcc *managedRunTypeCoreComponents) String() string {
	return factory.RunTypeCoreComponentsName
}
