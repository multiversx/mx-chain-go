package integrationTests

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/provider"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
)

func (tpn *TestProcessorNode) addGenesisBlocksIntoStorage() {
	for shardId, header := range tpn.GenesisBlocks {
		buffHeader, _ := TestMarshalizer.Marshal(header)
		headerHash := TestHasher.Compute(string(buffHeader))

		if shardId == core.MetachainShardId {
			metablockStorer := tpn.Storage.GetStorer(dataRetriever.MetaBlockUnit)
			_ = metablockStorer.Put(headerHash, buffHeader)
		} else {
			shardblockStorer := tpn.Storage.GetStorer(dataRetriever.BlockHeaderUnit)
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
	coreComponents.EnableEpochsHandlerField = &testscommon.EnableEpochsHandlerStub{}

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

	argumentsBase := block.ArgBaseProcessor{
		CoreComponents:      coreComponents,
		DataComponents:      dataComponents,
		BootstrapComponents: bootstrapComponents,
		StatusComponents:    statusComponents,
		Config:              triesConfig,
		AccountsDB:          accountsDb,
		ForkDetector:        nil,
		NodesCoordinator:    tpn.NodesCoordinator,
		FeeHandler:          tpn.FeeAccumulator,
		RequestHandler:      tpn.RequestHandler,
		BlockChainHook:      &testscommon.BlockChainHookStub{},
		EpochStartTrigger:   &mock.EpochStartTriggerStub{},
		HeaderValidator:     tpn.HeaderValidator,
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
		ReceiptsRepository:             &testscommon.ReceiptsRepositoryStub{},
	}

	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		tpn.ForkDetector, _ = sync.NewMetaForkDetector(tpn.RoundHandler, tpn.BlockBlackListHandler, tpn.BlockTracker, 0)
		argumentsBase.ForkDetector = tpn.ForkDetector
		argumentsBase.TxCoordinator = &mock.TransactionCoordinatorMock{}
		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:             argumentsBase,
			SCToProtocol:                 &mock.SCToProtocolStub{},
			PendingMiniBlocksHandler:     &mock.PendingMiniBlocksHandlerStub{},
			EpochStartDataCreator:        &mock.EpochStartDataCreatorStub{},
			EpochEconomics:               &mock.EpochEconomicsStub{},
			EpochRewardsCreator:          &mock.EpochRewardsCreatorStub{},
			EpochValidatorInfoCreator:    &mock.EpochValidatorInfoCreatorStub{},
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorStub{},
			EpochSystemSCProcessor:       &mock.EpochStartSystemSCStub{},
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)
	} else {
		tpn.ForkDetector, _ = sync.NewShardForkDetector(tpn.RoundHandler, tpn.BlockBlackListHandler, tpn.BlockTracker, 0)
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
		OutportHandler:               mock.NewNilOutport(),
		AccountsDBSyncer:             &mock.AccountsDBSyncerStub{},
		CurrentEpochProvider:         &testscommon.CurrentEpochProviderStub{},
		IsInImportMode:               false,
		HistoryRepo:                  &dblookupext.HistoryRepositoryStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessWaitTime:              tpn.RoundHandler.TimeDuration(),
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
		OutportHandler:               mock.NewNilOutport(),
		AccountsDBSyncer:             &mock.AccountsDBSyncerStub{},
		CurrentEpochProvider:         &testscommon.CurrentEpochProviderStub{},
		IsInImportMode:               false,
		HistoryRepo:                  &dblookupext.HistoryRepositoryStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessWaitTime:              tpn.RoundHandler.TimeDuration(),
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
	arg := provider.ArgMiniBlockProvider{
		MiniBlockPool:    tpn.DataPool.MiniBlocks(),
		MiniBlockStorage: tpn.Storage.GetStorer(dataRetriever.MiniBlockUnit),
		Marshalizer:      TestMarshalizer,
	}

	miniblockGetter, err := provider.NewMiniBlockProvider(arg)
	log.LogIfError(err)

	tpn.MiniblocksProvider = miniblockGetter
}
