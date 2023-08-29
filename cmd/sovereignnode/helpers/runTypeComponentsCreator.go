package helpers

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
)

func NewRunTypeComponentsFactoryArgs() (runType.RunTypeComponentsFactoryArgs, error) {
	shardBlockChainHookHandlerFactory, err := hooks.NewBlockChainHookFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewBlockChainHookFactory failed: %w", err)
	}
	blockChainHookHandlerFactory, err := hooks.NewSovereignBlockChainHookFactory(shardBlockChainHookHandlerFactory)
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignBlockChainHookFactory failed: %w", err)
	}

	shardEpochStartBootstrapperFactory, err := bootstrap.NewEpochStartBootstrapperFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewEpochStartBootstrapperFactory failed: %w", err)
	}
	epochStartBootstrapperFactory, err := bootstrap.NewSovereignEpochStartBootstrapperFactory(shardEpochStartBootstrapperFactory)
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignEpochStartBootstrapperFactory failed: %w", err)
	}

	shardBootstrapperFromStorageFactory, err := storageBootstrap.NewShardStorageBootstrapperFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewShardStorageBootstrapperFactory failed: %w", err)
	}
	bootstrapperFromStorageFactory, err := storageBootstrap.NewSovereignShardStorageBootstrapperFactory(shardBootstrapperFromStorageFactory)
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignShardStorageBootstrapperFactory failed: %w", err)
	}

	shardBlockProcessorFactory, err := block.NewShardBlockProcessorFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewShardBlockProcessorFactory failed: %w", err)
	}
	blockProcessorFactory, err := block.NewSovereignBlockProcessorFactory(shardBlockProcessorFactory)
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignBlockProcessorFactory failed: %w", err)
	}

	shardForkDetectorFactory, err := sync.NewShardForkDetectorFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewShardForkDetectorFactory failed: %w", err)
	}
	forkDetectorFactory, err := sync.NewSovereignForkDetectorFactory(shardForkDetectorFactory)
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignForkDetectorFactory failed: %w", err)
	}

	shardBlockTrackerFactory, err := track.NewShardBlockTrackerFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewShardBlockTrackerFactory failed: %w", err)
	}
	blockTrackerFactory, err := track.NewSovereignBlockTrackerFactory(shardBlockTrackerFactory)
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignBlockTrackerFactory failed: %w", err)
	}

	shardRequestHandlerFactory, err := requestHandlers.NewResolverRequestHandlerFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewResolverRequestHandlerFactory failed: %w", err)
	}
	requestHandlerFactory, err := requestHandlers.NewSovereignResolverRequestHandlerFactory(shardRequestHandlerFactory)
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignResolverRequestHandlerFactory failed: %w", err)
	}

	shardHeaderValidatorFactory, err := block.NewShardHeaderValidatorFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewShardHeaderValidatorFactory failed: %w", err)
	}
	headerValidatorFactory, err := block.NewSovereignHeaderValidatorFactory(shardHeaderValidatorFactory)
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignHeaderValidatorFactory failed: %w", err)
	}

	scheduledTxsExecutionFactory, err := preprocess.NewSovereignScheduledTxsExecutionFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignScheduledTxsExecutionFactory failed: %w", err)
	}

	shardTransactionCoordinatorFactory, err := coordinator.NewShardTransactionCoordinatorFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewShardTransactionCoordinatorFactory failed: %w", err)
	}
	transactionCoordinatorFactory, err := coordinator.NewSovereignTransactionCoordinatorFactory(shardTransactionCoordinatorFactory)
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignTransactionCoordinatorFactory failed: %w", err)
	}

	shardValidatorStatisticsProcessorFactory, err := peer.NewValidatorStatisticsProcessorFactory()
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewShardBlockProcessorFactory failed: %w", err)
	}
	validatorStatisticsProcessorFactory, err := peer.NewSovereignValidatorStatisticsProcessorFactory(shardValidatorStatisticsProcessorFactory)
	if err != nil {
		return runType.RunTypeComponentsFactoryArgs{}, fmt.Errorf("createManagedRunTypeComponents - NewSovereignValidatorStatisticsProcessorFactory failed: %w", err)
	}

	return runType.RunTypeComponentsFactoryArgs{
		BlockChainHookHandlerCreator:        blockChainHookHandlerFactory,
		EpochStartBootstrapperCreator:       epochStartBootstrapperFactory,
		BootstrapperFromStorageCreator:      bootstrapperFromStorageFactory,
		BlockProcessorCreator:               blockProcessorFactory,
		ForkDetectorCreator:                 forkDetectorFactory,
		BlockTrackerCreator:                 blockTrackerFactory,
		RequestHandlerCreator:               requestHandlerFactory,
		HeaderValidatorCreator:              headerValidatorFactory,
		ScheduledTxsExecutionCreator:        scheduledTxsExecutionFactory,
		TransactionCoordinatorCreator:       transactionCoordinatorFactory,
		ValidatorStatisticsProcessorCreator: validatorStatisticsProcessorFactory,
	}, nil
}
