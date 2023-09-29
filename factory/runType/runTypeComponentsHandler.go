package runType

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	processSync "github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
)

var _ factory.ComponentHandler = (*managedRunTypeComponents)(nil)
var _ factory.RunTypeComponentsHandler = (*managedRunTypeComponents)(nil)
var _ factory.RunTypeComponentsHolder = (*managedRunTypeComponents)(nil)

type managedRunTypeComponents struct {
	*runTypeComponents
	factory            runTypeComponentsCreator
	mutStateComponents sync.RWMutex
}

// NewManagedRunTypeComponents returns a news instance of managedRunTypeComponents
func NewManagedRunTypeComponents(rcf runTypeComponentsCreator) (*managedRunTypeComponents, error) {
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
	rtc, err := mrc.factory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrRunTypeComponentsFactoryCreate, err)
	}

	mrc.mutStateComponents.Lock()
	mrc.runTypeComponents = rtc
	mrc.mutStateComponents.Unlock()

	return nil
}

// Close will close all underlying subcomponents
func (mrc *managedRunTypeComponents) Close() error {
	mrc.mutStateComponents.Lock()
	defer mrc.mutStateComponents.Unlock()

	if check.IfNil(mrc.runTypeComponents) {
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

	if check.IfNil(mrc.runTypeComponents) {
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
	if check.IfNil(mrc.scProcessorCreator) {
		return errors.ErrNilSCProcessorCreator
	}
	return nil
}

// AdditionalStorageServiceCreator returns the additional storage service creator
func (mrc *managedRunTypeComponents) AdditionalStorageServiceCreator() process.AdditionalStorageServiceCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.additionalStorageServiceCreator
}

// BlockProcessorCreator returns the block processor creator
func (mrc *managedRunTypeComponents) BlockProcessorCreator() processBlock.BlockProcessorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.blockProcessorCreator
}

// BlockChainHookHandlerCreator returns the blockchain hook handler creator
func (mrc *managedRunTypeComponents) BlockChainHookHandlerCreator() hooks.BlockChainHookHandlerCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.blockChainHookHandlerCreator
}

// BootstrapperFromStorageCreator returns the bootstrapper from storage creator
func (mrc *managedRunTypeComponents) BootstrapperFromStorageCreator() storageBootstrap.BootstrapperFromStorageCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.bootstrapperFromStorageCreator
}

// BlockTrackerCreator returns the block tracker creator
func (mrc *managedRunTypeComponents) BlockTrackerCreator() track.BlockTrackerCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.blockTrackerCreator
}

// EpochStartBootstrapperCreator returns the epoch start bootstrapper creator
func (mrc *managedRunTypeComponents) EpochStartBootstrapperCreator() bootstrap.EpochStartBootstrapperCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.epochStartBootstrapperCreator
}

// ForkDetectorCreator returns the fork detector creator
func (mrc *managedRunTypeComponents) ForkDetectorCreator() processSync.ForkDetectorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.forkDetectorCreator
}

// HeaderValidatorCreator returns the header validator creator
func (mrc *managedRunTypeComponents) HeaderValidatorCreator() processBlock.HeaderValidatorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.headerValidatorCreator
}

// RequestHandlerCreator returns the request handler creator
func (mrc *managedRunTypeComponents) RequestHandlerCreator() requestHandlers.RequestHandlerCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.requestHandlerCreator
}

// ScheduledTxsExecutionCreator returns the scheduled transactions execution creator
func (mrc *managedRunTypeComponents) ScheduledTxsExecutionCreator() preprocess.ScheduledTxsExecutionCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.scheduledTxsExecutionCreator
}

// TransactionCoordinatorCreator returns the transaction coordinator creator
func (mrc *managedRunTypeComponents) TransactionCoordinatorCreator() coordinator.TransactionCoordinatorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.transactionCoordinatorCreator
}

// ValidatorStatisticsProcessorCreator returns the validator statistics processor creator
func (mrc *managedRunTypeComponents) ValidatorStatisticsProcessorCreator() peer.ValidatorStatisticsProcessorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.validatorStatisticsProcessorCreator
}

// SCProcessorCreator returns the smart contract processor creator
func (mrc *managedRunTypeComponents) SCProcessorCreator() scrCommon.SCProcessorCreator {
	mrc.mutStateComponents.RLock()
	defer mrc.mutStateComponents.RUnlock()

	if check.IfNil(mrc.runTypeComponents) {
		return nil
	}

	return mrc.runTypeComponents.scProcessorCreator
}

// IsInterfaceNil returns true if the interface is nil
func (mrc *managedRunTypeComponents) IsInterfaceNil() bool {
	return mrc == nil
}

// String returns the name of the component
func (mrc *managedRunTypeComponents) String() string {
	return factory.RunTypeComponentsName
}
