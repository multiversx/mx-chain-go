package runType

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
)

var _ factory.ComponentHandler = (*managedRunTypeComponents)(nil)
var _ factory.RunTypeComponentsHandler = (*managedRunTypeComponents)(nil)
var _ factory.RunTypeComponentsHolder = (*managedRunTypeComponents)(nil)

type managedRunTypeComponents struct {
	*runTypeComponents
	factory            *runTypeComponentsFactory
	mutStateComponents sync.RWMutex
}

// NewManagedRunTypeComponents returns a news instance of managedRunTypeComponents
func NewManagedRunTypeComponents(rcf *runTypeComponentsFactory) (*managedRunTypeComponents, error) {
	if rcf == nil {
		return nil, errors.ErrNilRunTypeComponentsFactory
	}

	return &managedRunTypeComponents{
		runTypeComponents: nil,
		factory:           rcf,
	}, nil
}

// Create will create the managed components
func (mrc *managedRunTypeComponents) Create() error {
	sc, err := mrc.factory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrRunTypeComponentsFactoryCreate, err)
	}

	mrc.mutStateComponents.Lock()
	mrc.runTypeComponents = sc
	mrc.mutStateComponents.Unlock()

	return nil
}

// Close will close all underlying subcomponents
func (mrc *managedRunTypeComponents) Close() error {
	mrc.mutStateComponents.Lock()
	defer mrc.mutStateComponents.Unlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	err := mrc.runTypeComponents.Close()
	if err != nil {
		return err
	}
	mrc.runTypeComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mrc *managedRunTypeComponents) CheckSubcomponents() error {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return errors.ErrNilRunTypeComponents
	}
	if check.IfNil(mrc.blockProcessorCreator) {
		return errors.ErrNilBlockProcessorCreator
	}
	if check.IfNil(mrc.blockChainHookHandlerCreator) {
		return errors.ErrNilBlockChainHookHandlerCreator
	}
	if check.IfNil(mrc.bootstrapperFromStorageCreator) {
		return errors.ErrNilBootstrapperFromStorageCreator
	}
	if check.IfNil(mrc.blockTrackerCreator) {
		return errors.ErrNilBlockTrackerCreator
	}
	if check.IfNil(mrc.epochStartBootstrapperCreator) {
		return errors.ErrNilEpochStartBootstrapperCreator
	}
	if check.IfNil(mrc.forkDetectorCreator) {
		return errors.ErrNilForkDetectorCreator
	}
	if check.IfNil(mrc.headerValidatorCreator) {
		return errors.ErrNilHeaderValidatorCreator
	}
	if check.IfNil(mrc.requestHandlerCreator) {
		return errors.ErrNilRequestHandlerCreator
	}
	if check.IfNil(mrc.scheduledTxsExecutionCreator) {
		return errors.ErrNilScheduledTxsExecutionCreator
	}
	if check.IfNil(mrc.transactionCoordinatorCreator) {
		return errors.ErrNilTransactionCoordinatorCreator
	}
	if check.IfNil(mrc.validatorStatisticsProcessorCreator) {
		return errors.ErrNilValidatorStatisticsProcessorCreator
	}
	if check.IfNil(mrc.additionalStorageServiceCreator) {
		return errors.ErrNilAdditionalStorageServiceCreator
	}
	return nil
}

// BlockProcessorCreator returns the block processor creator
func (mrc *managedRunTypeComponents) BlockProcessorCreator() factory.BlockProcessorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.blockProcessorCreator
}

// BlockChainHookHandlerCreator returns the blockchain hook handler creator
func (mrc *managedRunTypeComponents) BlockChainHookHandlerCreator() factory.BlockChainHookHandlerCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.blockChainHookHandlerCreator
}

// BootstrapperFromStorageCreator returns the bootstrapper from storage creator
func (mrc *managedRunTypeComponents) BootstrapperFromStorageCreator() factory.BootstrapperFromStorageCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.bootstrapperFromStorageCreator
}

// BlockTrackerCreator returns the block tracker creator
func (mrc *managedRunTypeComponents) BlockTrackerCreator() factory.BlockTrackerCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.blockTrackerCreator
}

// EpochStartBootstrapperCreator returns the epoch start bootstrapper creator
func (mrc *managedRunTypeComponents) EpochStartBootstrapperCreator() factory.EpochStartBootstrapperCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.epochStartBootstrapperCreator
}

// ForkDetectorCreator returns the fork detector creator
func (mrc *managedRunTypeComponents) ForkDetectorCreator() factory.ForkDetectorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.forkDetectorCreator
}

// HeaderValidatorCreator returns the header validator creator
func (mrc *managedRunTypeComponents) HeaderValidatorCreator() factory.HeaderValidatorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.headerValidatorCreator
}

// RequestHandlerCreator returns the request handler creator
func (mrc *managedRunTypeComponents) RequestHandlerCreator() factory.RequestHandlerCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.requestHandlerCreator
}

// ScheduledTxsExecutionCreator returns the scheduled transactions execution creator
func (mrc *managedRunTypeComponents) ScheduledTxsExecutionCreator() factory.ScheduledTxsExecutionCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.scheduledTxsExecutionCreator
}

// TransactionCoordinatorCreator returns the transaction coordinator creator
func (mrc *managedRunTypeComponents) TransactionCoordinatorCreator() factory.TransactionCoordinatorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.transactionCoordinatorCreator
}

// ValidatorStatisticsProcessorCreator returns the validator statistics processor creator
func (mrc *managedRunTypeComponents) ValidatorStatisticsProcessorCreator() factory.ValidatorStatisticsProcessorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.validatorStatisticsProcessorCreator
}

// AdditionalStorageServiceCreator returns the additional storage service creator
func (mrc *managedRunTypeComponents) AdditionalStorageServiceCreator() factory.AdditionalStorageServiceCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if mrc.runTypeComponents == nil {
		return nil
	}

	return mrc.runTypeComponents.additionalStorageServiceCreator
}

// IsInterfaceNil returns true if the interface is nil
func (mrc *managedRunTypeComponents) IsInterfaceNil() bool {
	return mrc == nil
}

// String returns the name of the component
func (mrc *managedRunTypeComponents) String() string {
	return factory.RunTypeComponentsName
}
