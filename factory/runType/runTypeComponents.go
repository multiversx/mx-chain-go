package runType

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
)

type runTypeComponentsFactory struct {
}

// runTypeComponents struct holds the components needed for a run type
type runTypeComponents struct {
	blockChainHookHandlerCreator        hooks.BlockChainHookHandlerCreator
	epochStartBootstrapperCreator       bootstrap.EpochStartBootstrapperCreator
	bootstrapperFromStorageCreator      storageBootstrap.BootstrapperFromStorageCreator
	blockProcessorCreator               processBlock.BlockProcessorCreator
	forkDetectorCreator                 sync.ForkDetectorCreator
	blockTrackerCreator                 track.BlockTrackerCreator
	requestHandlerCreator               requestHandlers.RequestHandlerCreator
	headerValidatorCreator              processBlock.HeaderValidatorCreator
	scheduledTxsExecutionCreator        preprocess.ScheduledTxsExecutionCreator
	transactionCoordinatorCreator       coordinator.TransactionCoordinatorCreator
	validatorStatisticsProcessorCreator peer.ValidatorStatisticsProcessorCreator
	additionalStorageServiceCreator     process.AdditionalStorageServiceCreator
}

// NewRunTypeComponentsFactory will return a new instance of runTypeComponentsFactory
func NewRunTypeComponentsFactory() (*runTypeComponentsFactory, error) {
	return &runTypeComponentsFactory{}, nil
}

// Create creates the runType components
func (rcf *runTypeComponentsFactory) Create() (*runTypeComponents, error) {
	blockChainHookHandlerFactory, err := hooks.NewBlockChainHookFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewBlockChainHookFactory failed: %w", err)
	}

	epochStartBootstrapperFactory := bootstrap.NewEpochStartBootstrapperFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewEpochStartBootstrapperFactory failed: %w", err)
	}

	bootstrapperFromStorageFactory, err := storageBootstrap.NewShardStorageBootstrapperFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardStorageBootstrapperFactory failed: %w", err)
	}

	blockProcessorFactory, err := block.NewShardBlockProcessorFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardBlockProcessorFactory failed: %w", err)
	}

	forkDetectorFactory, err := sync.NewShardForkDetectorFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardForkDetectorFactory failed: %w", err)
	}

	blockTrackerFactory, err := track.NewShardBlockTrackerFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardBlockTrackerFactory failed: %w", err)
	}

	requestHandlerFactory := requestHandlers.NewResolverRequestHandlerFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewResolverRequestHandlerFactory failed: %w", err)
	}

	headerValidatorFactory, err := block.NewShardHeaderValidatorFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardHeaderValidatorFactory failed: %w", err)
	}

	scheduledTxsExecutionFactory, err := preprocess.NewShardScheduledTxsExecutionFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewSovereignScheduledTxsExecutionFactory failed: %w", err)
	}

	transactionCoordinatorFactory, err := coordinator.NewShardTransactionCoordinatorFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardTransactionCoordinatorFactory failed: %w", err)
	}

	validatorStatisticsProcessorFactory, err := peer.NewValidatorStatisticsProcessorFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardBlockProcessorFactory failed: %w", err)
	}

	additionalStorageServiceCreator, err := storageFactory.NewShardAdditionalStorageServiceFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardAdditionalStorageServiceFactory failed: %w", err)
	}

	return &runTypeComponents{
		blockChainHookHandlerCreator:        blockChainHookHandlerFactory,
		epochStartBootstrapperCreator:       epochStartBootstrapperFactory,
		bootstrapperFromStorageCreator:      bootstrapperFromStorageFactory,
		blockProcessorCreator:               blockProcessorFactory,
		forkDetectorCreator:                 forkDetectorFactory,
		blockTrackerCreator:                 blockTrackerFactory,
		requestHandlerCreator:               requestHandlerFactory,
		headerValidatorCreator:              headerValidatorFactory,
		scheduledTxsExecutionCreator:        scheduledTxsExecutionFactory,
		transactionCoordinatorCreator:       transactionCoordinatorFactory,
		validatorStatisticsProcessorCreator: validatorStatisticsProcessorFactory,
		additionalStorageServiceCreator:     additionalStorageServiceCreator,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rc *runTypeComponentsFactory) IsInterfaceNil() bool {
	return rc == nil
}

// Close does nothing
func (rc *runTypeComponents) Close() error {
	return nil
}
