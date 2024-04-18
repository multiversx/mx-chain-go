package runType

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/consensus"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/errors"
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	processGenesis "github.com/multiversx/mx-chain-go/genesis/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/processProxy"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/sharding"
	nodesCoord "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type runTypeComponentsFactory struct {
	coreComponents process.CoreComponentsHolder
}

// runTypeComponents struct holds the components needed for a run type
type runTypeComponents struct {
	blockChainHookHandlerCreator            hooks.BlockChainHookHandlerCreator
	epochStartBootstrapperCreator           bootstrap.EpochStartBootstrapperCreator
	bootstrapperFromStorageCreator          storageBootstrap.BootstrapperFromStorageCreator
	bootstrapperCreator                     storageBootstrap.BootstrapperCreator
	blockProcessorCreator                   processBlock.BlockProcessorCreator
	forkDetectorCreator                     sync.ForkDetectorCreator
	blockTrackerCreator                     track.BlockTrackerCreator
	requestHandlerCreator                   requestHandlers.RequestHandlerCreator
	headerValidatorCreator                  processBlock.HeaderValidatorCreator
	scheduledTxsExecutionCreator            preprocess.ScheduledTxsExecutionCreator
	transactionCoordinatorCreator           coordinator.TransactionCoordinatorCreator
	validatorStatisticsProcessorCreator     peer.ValidatorStatisticsProcessorCreator
	additionalStorageServiceCreator         process.AdditionalStorageServiceCreator
	scProcessorCreator                      scrCommon.SCProcessorCreator
	scResultPreProcessorCreator             preprocess.SmartContractResultPreProcessorCreator
	consensusModel                          consensus.ConsensusModel
	vmContainerMetaFactory                  factoryVm.VmContainerCreator
	vmContainerShardFactory                 factoryVm.VmContainerCreator
	accountsCreator                         state.AccountFactory
	outGoingOperationsPoolHandler           sovereignBlock.OutGoingOperationsPool
	dataCodecHandler                        sovereign.DataCodecHandler
	topicsCheckerHandler                    sovereign.TopicsCheckerHandler
	shardCoordinatorCreator                 sharding.ShardCoordinatorFactory
	nodesCoordinatorWithRaterFactoryCreator nodesCoord.NodesCoordinatorWithRaterFactory
	requestersContainerFactoryCreator       requesterscontainer.RequesterContainerFactoryCreator
	interceptorsContainerFactoryCreator     interceptorscontainer.InterceptorsContainerFactoryCreator
	shardResolversContainerFactoryCreator   resolverscontainer.ShardResolversContainerFactoryCreator
	txPreProcessorCreator                   preprocess.TxPreProcessorCreator
	extraHeaderSigVerifierHandler           headerCheck.ExtraHeaderSigVerifierHolder
	genesisBlockCreatorFactory              processGenesis.GenesisBlockCreatorFactory
}

// NewRunTypeComponentsFactory will return a new instance of runTypeComponentsFactory
func NewRunTypeComponentsFactory(coreComponents process.CoreComponentsHolder) (*runTypeComponentsFactory, error) {
	if check.IfNil(coreComponents) {
		return nil, errors.ErrNilCoreComponents
	}

	return &runTypeComponentsFactory{
		coreComponents: coreComponents,
	}, nil
}

// Create creates the runType components
func (rcf *runTypeComponentsFactory) Create() (*runTypeComponents, error) {
	blockChainHookHandlerFactory, err := hooks.NewBlockChainHookFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewBlockChainHookFactory failed: %w", err)
	}

	epochStartBootstrapperFactory := bootstrap.NewEpochStartBootstrapperFactory()

	bootstrapperFromStorageFactory, err := storageBootstrap.NewShardStorageBootstrapperFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardStorageBootstrapperFactory failed: %w", err)
	}

	shardBootstrapFactory, err := storageBootstrap.NewShardBootstrapFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardBootstrapFactory failed: %w", err)
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

	headerValidatorFactory, err := block.NewShardHeaderValidatorFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewShardHeaderValidatorFactory failed: %w", err)
	}

	scheduledTxsExecutionFactory, err := preprocess.NewShardScheduledTxsExecutionFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewSovereignScheduledTxsExecutionFactory failed: %w", err)
	}

	scResultsPreProcessorCreator, err := preprocess.NewSmartContractResultPreProcessorFactory()
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewSmartContractResultPreProcessorFactory failed: %w", err)
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

	scProcessorCreator := processProxy.NewSCProcessProxyFactory()

	vmContainerMetaCreator, err := factoryVm.NewVmContainerMetaFactory(blockChainHookHandlerFactory)
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewVmContainerMetaFactory failed: %w", err)
	}

	vmContainerShardCreator, err := factoryVm.NewVmContainerShardFactory(blockChainHookHandlerFactory)
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewVmContainerShardFactory failed: %w", err)
	}

	accountsCreator, err := factory.NewAccountCreator(factory.ArgsAccountCreator{
		Hasher:              rcf.coreComponents.Hasher(),
		Marshaller:          rcf.coreComponents.InternalMarshalizer(),
		EnableEpochsHandler: rcf.coreComponents.EnableEpochsHandler(),
	})
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewAccountCreator failed: %w", err)
	}

	return &runTypeComponents{
		blockChainHookHandlerCreator:            blockChainHookHandlerFactory,
		epochStartBootstrapperCreator:           epochStartBootstrapperFactory,
		bootstrapperFromStorageCreator:          bootstrapperFromStorageFactory,
		bootstrapperCreator:                     shardBootstrapFactory,
		blockProcessorCreator:                   blockProcessorFactory,
		forkDetectorCreator:                     forkDetectorFactory,
		blockTrackerCreator:                     blockTrackerFactory,
		requestHandlerCreator:                   requestHandlerFactory,
		headerValidatorCreator:                  headerValidatorFactory,
		scheduledTxsExecutionCreator:            scheduledTxsExecutionFactory,
		transactionCoordinatorCreator:           transactionCoordinatorFactory,
		validatorStatisticsProcessorCreator:     validatorStatisticsProcessorFactory,
		additionalStorageServiceCreator:         additionalStorageServiceCreator,
		scProcessorCreator:                      scProcessorCreator,
		scResultPreProcessorCreator:             scResultsPreProcessorCreator,
		consensusModel:                          consensus.ConsensusModelV1,
		vmContainerMetaFactory:                  vmContainerMetaCreator,
		vmContainerShardFactory:                 vmContainerShardCreator,
		accountsCreator:                         accountsCreator,
		outGoingOperationsPoolHandler:           disabled.NewDisabledOutGoingOperationPool(),
		dataCodecHandler:                        disabled.NewDisabledDataCodec(),
		topicsCheckerHandler:                    disabled.NewDisabledTopicsChecker(),
		shardCoordinatorCreator:                 sharding.NewMultiShardCoordinatorFactory(),
		nodesCoordinatorWithRaterFactoryCreator: nodesCoord.NewIndexHashedNodesCoordinatorWithRaterFactory(),
		requestersContainerFactoryCreator:       requesterscontainer.NewShardRequestersContainerFactoryCreator(),
		interceptorsContainerFactoryCreator:     interceptorscontainer.NewShardInterceptorsContainerFactoryCreator(),
		shardResolversContainerFactoryCreator:   resolverscontainer.NewShardResolversContainerFactoryCreator(),
		txPreProcessorCreator:                   preprocess.NewTxPreProcessorCreator(),
		extraHeaderSigVerifierHandler:           headerCheck.NewExtraHeaderSigVerifierHolder(),
		genesisBlockCreatorFactory:              processGenesis.NewGenesisBlockCreatorFactory(),
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

// IsInterfaceNil returns true if there is no value under the interface
func (rc *runTypeComponents) IsInterfaceNil() bool {
	return rc == nil
}

//reset github actions
