package mainFactoryMocks

import (
	"github.com/multiversx/mx-chain-go/factory"
)

// RunTypeComponentsMock -
type RunTypeComponentsMock struct {
	BlockChainHookHandlerFactory        factory.BlockChainHookHandlerCreator
	BlockProcessorFactory               factory.BlockProcessorCreator
	BlockTrackerFactory                 factory.BlockTrackerCreator
	BootstrapperFromStorageFactory      factory.BootstrapperFromStorageCreator
	EpochStartBootstrapperFactory       factory.EpochStartBootstrapperCreator
	ForkDetectorFactory                 factory.ForkDetectorCreator
	HeaderValidatorFactory              factory.HeaderValidatorCreator
	RequestHandlerFactory               factory.RequestHandlerCreator
	ScheduledTxsExecutionFactory        factory.ScheduledTxsExecutionCreator
	TransactionCoordinatorFactory       factory.TransactionCoordinatorCreator
	ValidatorStatisticsProcessorFactory factory.ValidatorStatisticsProcessorCreator
	AdditionalStorageServiceFactory     factory.AdditionalStorageServiceCreator
}

// Create -
func (r *RunTypeComponentsMock) Create() error {
	return nil
}

// Close -
func (r *RunTypeComponentsMock) Close() error {
	return nil
}

// CheckSubcomponents -
func (r *RunTypeComponentsMock) CheckSubcomponents() error {
	return nil
}

// BlockChainHookHandlerCreator -
func (r *RunTypeComponentsMock) String() string {
	return ""
}

// BlockChainHookHandlerCreator -
func (r *RunTypeComponentsMock) BlockChainHookHandlerCreator() factory.BlockChainHookHandlerCreator {
	return r.BlockChainHookHandlerFactory
}

// BlockProcessorCreator -
func (r *RunTypeComponentsMock) BlockProcessorCreator() factory.BlockProcessorCreator {
	return r.BlockProcessorFactory
}

// BlockTrackerCreator -
func (r *RunTypeComponentsMock) BlockTrackerCreator() factory.BlockTrackerCreator {
	return r.BlockTrackerFactory
}

// BootstrapperFromStorageCreator -
func (r *RunTypeComponentsMock) BootstrapperFromStorageCreator() factory.BootstrapperFromStorageCreator {
	return r.BootstrapperFromStorageFactory
}

// EpochStartBootstrapperCreator -
func (r *RunTypeComponentsMock) EpochStartBootstrapperCreator() factory.EpochStartBootstrapperCreator {
	return r.EpochStartBootstrapperFactory
}

// ForkDetectorCreator -
func (r *RunTypeComponentsMock) ForkDetectorCreator() factory.ForkDetectorCreator {
	return r.ForkDetectorFactory
}

// HeaderValidatorCreator -
func (r *RunTypeComponentsMock) HeaderValidatorCreator() factory.HeaderValidatorCreator {
	return r.HeaderValidatorFactory
}

// RequestHandlerCreator -
func (r *RunTypeComponentsMock) RequestHandlerCreator() factory.RequestHandlerCreator {
	return r.RequestHandlerFactory
}

// ScheduledTxsExecutionCreator -
func (r *RunTypeComponentsMock) ScheduledTxsExecutionCreator() factory.ScheduledTxsExecutionCreator {
	return r.ScheduledTxsExecutionFactory
}

// TransactionCoordinatorCreator -
func (r *RunTypeComponentsMock) TransactionCoordinatorCreator() factory.TransactionCoordinatorCreator {
	return r.TransactionCoordinatorFactory
}

// ValidatorStatisticsProcessorCreator -
func (r *RunTypeComponentsMock) ValidatorStatisticsProcessorCreator() factory.ValidatorStatisticsProcessorCreator {
	return r.ValidatorStatisticsProcessorFactory
}

// AdditionalStorageServiceCreator -
func (r *RunTypeComponentsMock) AdditionalStorageServiceCreator() factory.AdditionalStorageServiceCreator {
	return r.AdditionalStorageServiceFactory
}

// IsInterfaceNil -
func (r *RunTypeComponentsMock) IsInterfaceNil() bool {
	return r == nil
}
