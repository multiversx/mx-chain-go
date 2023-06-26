package integrationTests

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/provider"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/outport/disabled"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/shardingmock"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func (tpn *TestProcessorNode) addGenesisBlocksIntoStorage() {
	for shardId, header := range tpn.GenesisBlocks {
		buffHeader, _ := TestMarshalizer.Marshal(header)
		headerHash := TestHasher.Compute(string(buffHeader))

		if shardId == core.MetachainShardId {
			metablockStorer, _ := tpn.Storage.GetStorer(dataRetriever.MetaBlockUnit)
			_ = metablockStorer.Put(headerHash, buffHeader)
		} else {
			shardblockStorer, _ := tpn.Storage.GetStorer(dataRetriever.BlockHeaderUnit)
			_ = shardblockStorer.Put(headerHash, buffHeader)
		}
	}
}

func (tpn *TestProcessorNode) initBlockProcessorWithSync() {
	var err error

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = tpn.AccntState
	accountsDb[state.PeerAccountsState] = tpn.PeerState

	coreComponents := GetDefaultCoreComponents()
	coreComponents.InternalMarshalizerField = TestMarshalizer
	coreComponents.HasherField = TestHasher
	coreComponents.Uint64ByteSliceConverterField = TestUint64Converter
	coreComponents.EpochNotifierField = tpn.EpochNotifier
	coreComponents.EnableEpochsHandlerField = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		RefactorPeersMiniBlocksEnableEpochField: UnreachableEpoch,
	}

	dataComponents := GetDefaultDataComponents()
	dataComponents.Store = tpn.Storage
	dataComponents.DataPool = tpn.DataPool
	dataComponents.BlockChain = tpn.BlockChain

	bootstrapComponents := getDefaultBootstrapComponents(tpn.ShardCoordinator)
	bootstrapComponents.HdrIntegrityVerifier = tpn.HeaderIntegrityVerifier

	statusComponents := GetDefaultStatusComponents()

	triesConfig := config.Config{
		StateTriesConfig: config.StateTriesConfig{
			CheckpointRoundsModulus: stateCheckpointModulus,
		},
	}

	statusCoreComponents := &factory.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}

	argumentsBase := block.ArgBaseProcessor{
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		BootstrapComponents:  bootstrapComponents,
		StatusComponents:     statusComponents,
		StatusCoreComponents: statusCoreComponents,
		Config:               triesConfig,
		AccountsDB:           accountsDb,
		ForkDetector:         nil,
		NodesCoordinator:     tpn.NodesCoordinator,
		FeeHandler:           tpn.FeeAccumulator,
		RequestHandler:       tpn.RequestHandler,
		BlockChainHook:       &testscommon.BlockChainHookStub{},
		EpochStartTrigger:    &mock.EpochStartTriggerStub{},
		HeaderValidator:      tpn.HeaderValidator,
		BootStorer: &mock.BoostrapStorerMock{
			PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
				return nil
			},
		},
		BlockTracker:                 tpn.BlockTracker,
		BlockSizeThrottler:           TestBlockSizeThrottler,
		HistoryRepository:            tpn.HistoryRepository,
		EnableRoundsHandler:          coreComponents.EnableRoundsHandler(),
		GasHandler:                   tpn.GasHandler,
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		ReceiptsRepository:           &testscommon.ReceiptsRepositoryStub{},
		OutportDataProvider:          &outport.OutportDataProviderStub{},
		BlockProcessingCutoffHandler: &testscommon.BlockProcessingCutoffStub{},
		ChainParametersHandler: &shardingmock.ChainParametersHandlerStub{
			ChainParametersForEpochCalled: func(_ uint32) (config.ChainParametersByEpochConfig, error) {
				return config.ChainParametersByEpochConfig{ShardFinality: 4, MetaFinality: 4}, nil
			},
		},
	}

	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		tpn.ForkDetector, _ = sync.NewMetaForkDetector(
			tpn.RoundHandler,
			tpn.BlockBlackListHandler,
			tpn.BlockTracker,
			&shardingmock.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(_ uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{ShardFinality: 4, MetaFinality: 4}, nil
				},
			},
			0)
		argumentsBase.ForkDetector = tpn.ForkDetector
		argumentsBase.TxCoordinator = &mock.TransactionCoordinatorMock{}
		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:          argumentsBase,
			SCToProtocol:              &mock.SCToProtocolStub{},
			PendingMiniBlocksHandler:  &mock.PendingMiniBlocksHandlerStub{},
			EpochStartDataCreator:     &mock.EpochStartDataCreatorStub{},
			EpochEconomics:            &mock.EpochEconomicsStub{},
			EpochRewardsCreator:       &mock.EpochRewardsCreatorStub{},
			EpochValidatorInfoCreator: &mock.EpochValidatorInfoCreatorStub{},
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorStub{
				UpdatePeerStateCalled: func(header data.MetaHeaderHandler) ([]byte, error) {
					return []byte("validator stats root hash"), nil
				},
			},
			EpochSystemSCProcessor: &mock.EpochStartSystemSCStub{},
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)
	} else {
		tpn.ForkDetector, _ = sync.NewShardForkDetector(
			tpn.RoundHandler,
			tpn.BlockBlackListHandler,
			tpn.BlockTracker,
			&shardingmock.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(_ uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{ShardFinality: 4, MetaFinality: 4}, nil
				},
			},
			0)
		argumentsBase.ForkDetector = tpn.ForkDetector
		argumentsBase.BlockChainHook = tpn.BlockchainHook
		argumentsBase.TxCoordinator = tpn.TxCoordinator
		argumentsBase.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{}
		arguments := block.ArgShardProcessor{
			ArgBaseProcessor: argumentsBase,
		}

		tpn.BlockProcessor, err = block.NewShardProcessor(arguments)
	}

	if err != nil {
		panic(fmt.Sprintf("Error creating blockprocessor: %s", err.Error()))
	}
}

func (tpn *TestProcessorNode) createShardBootstrapper() (TestBootstrapper, error) {
	argsBaseBootstrapper := sync.ArgBaseBootstrapper{
		PoolsHolder:                  tpn.DataPool,
		Store:                        tpn.Storage,
		ChainHandler:                 tpn.BlockChain,
		RoundHandler:                 tpn.RoundHandler,
		BlockProcessor:               tpn.BlockProcessor,
		WaitTime:                     tpn.RoundHandler.TimeDuration(),
		Hasher:                       TestHasher,
		Marshalizer:                  TestMarshalizer,
		ForkDetector:                 tpn.ForkDetector,
		RequestHandler:               tpn.RequestHandler,
		ShardCoordinator:             tpn.ShardCoordinator,
		Accounts:                     tpn.AccntState,
		BlackListHandler:             tpn.BlockBlackListHandler,
		NetworkWatcher:               tpn.Messenger,
		BootStorer:                   tpn.BootstrapStorer,
		StorageBootstrapper:          tpn.StorageBootstrapper,
		EpochHandler:                 tpn.EpochStartTrigger,
		MiniblocksProvider:           tpn.MiniblocksProvider,
		Uint64Converter:              TestUint64Converter,
		AppStatusHandler:             TestAppStatusHandler,
		OutportHandler:               disabled.NewDisabledOutport(),
		AccountsDBSyncer:             &mock.AccountsDBSyncerStub{},
		CurrentEpochProvider:         &testscommon.CurrentEpochProviderStub{},
		IsInImportMode:               false,
		HistoryRepo:                  &dblookupext.HistoryRepositoryStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessWaitTime:              tpn.RoundHandler.TimeDuration(),
		RepopulateTokensSupplies:     false,
	}

	argsShardBootstrapper := sync.ArgShardBootstrapper{
		ArgBaseBootstrapper: argsBaseBootstrapper,
	}

	bootstrap, err := sync.NewShardBootstrap(argsShardBootstrapper)
	if err != nil {
		return nil, err
	}

	return &sync.TestShardBootstrap{
		ShardBootstrap: bootstrap,
	}, nil
}

func (tpn *TestProcessorNode) createMetaChainBootstrapper() (TestBootstrapper, error) {
	argsBaseBootstrapper := sync.ArgBaseBootstrapper{
		PoolsHolder:                  tpn.DataPool,
		Store:                        tpn.Storage,
		ChainHandler:                 tpn.BlockChain,
		RoundHandler:                 tpn.RoundHandler,
		BlockProcessor:               tpn.BlockProcessor,
		WaitTime:                     tpn.RoundHandler.TimeDuration(),
		Hasher:                       TestHasher,
		Marshalizer:                  TestMarshalizer,
		ForkDetector:                 tpn.ForkDetector,
		RequestHandler:               tpn.RequestHandler,
		ShardCoordinator:             tpn.ShardCoordinator,
		Accounts:                     tpn.AccntState,
		BlackListHandler:             tpn.BlockBlackListHandler,
		NetworkWatcher:               tpn.Messenger,
		BootStorer:                   tpn.BootstrapStorer,
		StorageBootstrapper:          tpn.StorageBootstrapper,
		EpochHandler:                 tpn.EpochStartTrigger,
		MiniblocksProvider:           tpn.MiniblocksProvider,
		Uint64Converter:              TestUint64Converter,
		AppStatusHandler:             TestAppStatusHandler,
		OutportHandler:               disabled.NewDisabledOutport(),
		AccountsDBSyncer:             &mock.AccountsDBSyncerStub{},
		CurrentEpochProvider:         &testscommon.CurrentEpochProviderStub{},
		IsInImportMode:               false,
		HistoryRepo:                  &dblookupext.HistoryRepositoryStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessWaitTime:              tpn.RoundHandler.TimeDuration(),
		RepopulateTokensSupplies:     false,
	}

	argsMetaBootstrapper := sync.ArgMetaBootstrapper{
		ArgBaseBootstrapper:         argsBaseBootstrapper,
		EpochBootstrapper:           tpn.EpochStartTrigger,
		ValidatorAccountsDB:         tpn.PeerState,
		ValidatorStatisticsDBSyncer: &mock.AccountsDBSyncerStub{},
	}

	bootstrap, err := sync.NewMetaBootstrap(argsMetaBootstrapper)
	if err != nil {
		return nil, err
	}

	return &sync.TestMetaBootstrap{
		MetaBootstrap: bootstrap,
	}, nil
}

func (tpn *TestProcessorNode) initBootstrapper() {
	tpn.createMiniblocksProvider()

	if tpn.ShardCoordinator.SelfId() < tpn.ShardCoordinator.NumberOfShards() {
		tpn.Bootstrapper, _ = tpn.createShardBootstrapper()
	} else {
		tpn.Bootstrapper, _ = tpn.createMetaChainBootstrapper()
	}
}

func (tpn *TestProcessorNode) createMiniblocksProvider() {
	storer, _ := tpn.Storage.GetStorer(dataRetriever.MiniBlockUnit)
	arg := provider.ArgMiniBlockProvider{
		MiniBlockPool:    tpn.DataPool.MiniBlocks(),
		MiniBlockStorage: storer,
		Marshalizer:      TestMarshalizer,
	}

	miniblockGetter, err := provider.NewMiniBlockProvider(arg)
	log.LogIfError(err)

	tpn.MiniblocksProvider = miniblockGetter
}
