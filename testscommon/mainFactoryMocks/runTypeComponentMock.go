package mainFactoryMocks

import (
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
)

// RunTypeComponentsMock -
type RunTypeComponentsMock struct {
	BlockChainHookHandlerFactory        hooks.BlockChainHookHandlerCreator
	BlockProcessorFactory               block.BlockProcessorCreator
	BlockTrackerFactory                 track.BlockTrackerCreator
	BootstrapperFromStorageFactory      storageBootstrap.BootstrapperFromStorageCreator
	EpochStartBootstrapperFactory       bootstrap.EpochStartBootstrapperCreator
	ForkDetectorFactory                 sync.ForkDetectorCreator
	HeaderValidatorFactory              block.HeaderValidatorCreator
	RequestHandlerFactory               requestHandlers.RequestHandlerCreator
	ScheduledTxsExecutionFactory        preprocess.ScheduledTxsExecutionCreator
	TransactionCoordinatorFactory       coordinator.TransactionCoordinatorCreator
	ValidatorStatisticsProcessorFactory peer.ValidatorStatisticsProcessorCreator
	AdditionalStorageServiceFactory     process.AdditionalStorageServiceCreator
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
func (r *RunTypeComponentsMock) BlockChainHookHandlerCreator() hooks.BlockChainHookHandlerCreator {
	return r.BlockChainHookHandlerFactory
}

// BlockProcessorCreator -
func (r *RunTypeComponentsMock) BlockProcessorCreator() block.BlockProcessorCreator {
	return r.BlockProcessorFactory
}

// BlockTrackerCreator -
func (r *RunTypeComponentsMock) BlockTrackerCreator() track.BlockTrackerCreator {
	return r.BlockTrackerFactory
}

// BootstrapperFromStorageCreator -
func (r *RunTypeComponentsMock) BootstrapperFromStorageCreator() storageBootstrap.BootstrapperFromStorageCreator {
	return r.BootstrapperFromStorageFactory
}

// EpochStartBootstrapperCreator -
func (r *RunTypeComponentsMock) EpochStartBootstrapperCreator() bootstrap.EpochStartBootstrapperCreator {
	return r.EpochStartBootstrapperFactory
}

// ForkDetectorCreator -
func (r *RunTypeComponentsMock) ForkDetectorCreator() sync.ForkDetectorCreator {
	return r.ForkDetectorFactory
}

// HeaderValidatorCreator -
func (r *RunTypeComponentsMock) HeaderValidatorCreator() block.HeaderValidatorCreator {
	return r.HeaderValidatorFactory
}

// RequestHandlerCreator -
func (r *RunTypeComponentsMock) RequestHandlerCreator() requestHandlers.RequestHandlerCreator {
	return r.RequestHandlerFactory
}

// ScheduledTxsExecutionCreator -
func (r *RunTypeComponentsMock) ScheduledTxsExecutionCreator() preprocess.ScheduledTxsExecutionCreator {
	return r.ScheduledTxsExecutionFactory
}

// TransactionCoordinatorCreator -
func (r *RunTypeComponentsMock) TransactionCoordinatorCreator() coordinator.TransactionCoordinatorCreator {
	return r.TransactionCoordinatorFactory
}

// ValidatorStatisticsProcessorCreator -
func (r *RunTypeComponentsMock) ValidatorStatisticsProcessorCreator() peer.ValidatorStatisticsProcessorCreator {
	return r.ValidatorStatisticsProcessorFactory
}

// AdditionalStorageServiceCreator -
func (r *RunTypeComponentsMock) AdditionalStorageServiceCreator() process.AdditionalStorageServiceCreator {
	return r.AdditionalStorageServiceFactory
}

// IsInterfaceNil -
func (r *RunTypeComponentsMock) IsInterfaceNil() bool {
	return r == nil
}
