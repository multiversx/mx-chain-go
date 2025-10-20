package integrationTests

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/executionTrack"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
	"github.com/multiversx/mx-chain-go/process/block/headerForBlock"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/estimator"
	"github.com/multiversx/mx-chain-go/process/missingData"

	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
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

	if tpn.EpochNotifier == nil {
		tpn.EpochNotifier = forking.NewGenericEpochNotifier()
	}
	if tpn.EnableEpochsHandler == nil {
		tpn.EnableEpochsHandler, _ = enablers.NewEnableEpochsHandler(CreateEnableEpochsConfig(), tpn.EpochNotifier)
	}
	coreComponents := GetDefaultCoreComponents(tpn.EnableEpochsHandler, tpn.EpochNotifier)
	coreComponents.InternalMarshalizerField = TestMarshalizer
	coreComponents.HasherField = TestHasher
	coreComponents.Uint64ByteSliceConverterField = TestUint64Converter
	coreComponents.EpochNotifierField = tpn.EpochNotifier
	coreComponents.RoundNotifierField = tpn.RoundNotifier

	dataComponents := GetDefaultDataComponents()
	dataComponents.Store = tpn.Storage
	dataComponents.DataPool = tpn.DataPool
	dataComponents.BlockChain = tpn.BlockChain

	bootstrapComponents := getDefaultBootstrapComponents(tpn.ShardCoordinator, tpn.EnableEpochsHandler)
	bootstrapComponents.HdrIntegrityVerifier = tpn.HeaderIntegrityVerifier

	statusComponents := GetDefaultStatusComponents()

	statusCoreComponents := &factory.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}

	argsHeadersForBlock := headerForBlock.ArgHeadersForBlock{
		DataPool:            tpn.DataPool,
		RequestHandler:      tpn.RequestHandler,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
		ShardCoordinator:    tpn.ShardCoordinator,
		BlockTracker:        tpn.BlockTracker,
		TxCoordinator:       tpn.TxCoordinator,
		RoundHandler:        tpn.RoundHandler,
		ExtraDelayForRequestBlockInfoInMilliseconds: 100,
		GenesisNonce: tpn.GenesisBlocks[tpn.ShardCoordinator.SelfId()].GetNonce(),
	}
	hdrsForBlock, err := headerForBlock.NewHeadersForBlock(argsHeadersForBlock)
	if err != nil {
		log.Error("initBlockProcessor NewHeadersForBlock", "error", err)
	}

	blockDataRequesterArgs := coordinator.BlockDataRequestArgs{
		RequestHandler:      tpn.RequestHandler,
		MiniBlockPool:       tpn.DataPool.MiniBlocks(),
		PreProcessors:       tpn.PreProcessorsProposal,
		ShardCoordinator:    tpn.ShardCoordinator,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
	}
	// second instance for proposal missing data fetching to avoid interferences
	proposalBlockDataRequester, err := coordinator.NewBlockDataRequester(blockDataRequesterArgs)
	if err != nil {
		log.LogIfError(err)
	}

	mbSelectionSession, err := block.NewMiniBlocksSelectionSession(
		tpn.ShardCoordinator.SelfId(),
		TestMarshalizer,
		TestHasher,
	)
	if err != nil {
		log.LogIfError(err)
	}

	executionResultsTracker := executionTrack.NewExecutionResultsTracker()
	err = process.SetBaseExecutionResult(executionResultsTracker, tpn.BlockChain)
	if err != nil {
		log.LogIfError(err)
	}

	execResultsVerifier, err := block.NewExecutionResultsVerifier(tpn.BlockChain, executionResultsTracker)
	if err != nil {
		log.LogIfError(err)
	}

	inclusionEstimator := estimator.NewExecutionResultInclusionEstimator(
		config.ExecutionResultInclusionEstimatorConfig{
			SafetyMargin:       110,
			MaxResultsPerBlock: 20,
		},

		tpn.RoundHandler,
	)

	missingDataArgs := missingData.ResolverArgs{
		HeadersPool:        tpn.DataPool.Headers(),
		ProofsPool:         tpn.DataPool.Proofs(),
		RequestHandler:     tpn.RequestHandler,
		BlockDataRequester: proposalBlockDataRequester,
	}
	missingDataResolver, err := missingData.NewMissingDataResolver(missingDataArgs)
	if err != nil {
		log.LogIfError(err)
	}

	argsGasConsumption := block.ArgsGasConsumption{
		EconomicsFee:                      tpn.EconomicsData,
		ShardCoordinator:                  tpn.ShardCoordinator,
		GasHandler:                        tpn.GasHandler,
		BlockCapacityOverestimationFactor: 200,
		PercentDecreaseLimitsStep:         10,
	}
	gasConsumption, err := block.NewGasConsumption(argsGasConsumption)
	if err != nil {
		log.LogIfError(err)
	}

	argumentsBase := block.ArgBaseProcessor{
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		BootstrapComponents:  bootstrapComponents,
		StatusComponents:     statusComponents,
		StatusCoreComponents: statusCoreComponents,
		Config:               config.Config{},
		AccountsDB:           accountsDb,
		AccountsProposal:     tpn.AccntStateProposal,
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
		BlockTracker:                       tpn.BlockTracker,
		BlockSizeThrottler:                 TestBlockSizeThrottler,
		HistoryRepository:                  tpn.HistoryRepository,
		GasHandler:                         tpn.GasHandler,
		ScheduledTxsExecutionHandler:       &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:         &testscommon.ProcessedMiniBlocksTrackerStub{},
		ReceiptsRepository:                 &testscommon.ReceiptsRepositoryStub{},
		OutportDataProvider:                &outport.OutportDataProviderStub{},
		BlockProcessingCutoffHandler:       &testscommon.BlockProcessingCutoffStub{},
		ManagedPeersHolder:                 &testscommon.ManagedPeersHolderStub{},
		SentSignaturesTracker:              &testscommon.SentSignatureTrackerStub{},
		HeadersForBlock:                    hdrsForBlock,
		MiniBlocksSelectionSession:         mbSelectionSession,
		ExecutionResultsVerifier:           execResultsVerifier,
		MissingDataResolver:                missingDataResolver,
		ExecutionResultsInclusionEstimator: inclusionEstimator,
		ExecutionResultsTracker:            executionResultsTracker,
		GasComputation:                     gasConsumption,
	}

	tpn.BlocksQueue = queue.NewBlocksQueue()

	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		tpn.ForkDetector, _ = sync.NewMetaForkDetector(
			tpn.RoundHandler,
			tpn.BlockBlackListHandler,
			tpn.BlockTracker,
			0,
			0,
			tpn.EnableEpochsHandler,
			tpn.EnableRoundsHandler,
			tpn.DataPool.Proofs(),
			tpn.ChainParametersHandler,
			tpn.ProcessConfigsHandler,
		)
		argumentsBase.ForkDetector = tpn.ForkDetector
		argumentsBase.TxCoordinator = &mock.TransactionCoordinatorMock{}
		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:          argumentsBase,
			SCToProtocol:              &mock.SCToProtocolStub{},
			PendingMiniBlocksHandler:  &mock.PendingMiniBlocksHandlerStub{},
			EpochStartDataCreator:     &mock.EpochStartDataCreatorStub{},
			EpochEconomics:            &mock.EpochEconomicsStub{},
			EpochRewardsCreator:       &testscommon.RewardsCreatorStub{},
			EpochValidatorInfoCreator: &testscommon.EpochValidatorInfoCreatorStub{},
			ValidatorStatisticsProcessor: &testscommon.ValidatorStatisticsProcessorStub{
				UpdatePeerStateCalled: func(header data.MetaHeaderHandler) ([]byte, error) {
					return []byte("validator stats root hash"), nil
				},
			},
			EpochSystemSCProcessor: &testscommon.EpochStartSystemSCStub{},
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)
	} else {
		tpn.ForkDetector, _ = sync.NewShardForkDetector(
			tpn.RoundHandler,
			tpn.BlockBlackListHandler,
			tpn.BlockTracker,
			0,
			0,
			tpn.EnableEpochsHandler,
			tpn.EnableRoundsHandler,
			tpn.DataPool.Proofs(),
			tpn.ChainParametersHandler,
			tpn.ProcessConfigsHandler,
		)
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
		BlocksQueue:                  tpn.BlocksQueue,
		Hasher:                       TestHasher,
		Marshalizer:                  TestMarshalizer,
		ForkDetector:                 tpn.ForkDetector,
		RequestHandler:               tpn.RequestHandler,
		ShardCoordinator:             tpn.ShardCoordinator,
		Accounts:                     tpn.AccntState,
		BlackListHandler:             tpn.BlockBlackListHandler,
		NetworkWatcher:               tpn.MainMessenger,
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
		ProcessWaitTimeSupernova:     tpn.RoundHandler.TimeDuration(),
		RepopulateTokensSupplies:     false,
		EnableEpochsHandler:          tpn.EnableEpochsHandler,
		EnableRoundsHandler:          tpn.EnableRoundsHandler,
		ProcessConfigsHandler:        tpn.ProcessConfigsHandler,
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
		BlocksQueue:                  tpn.BlocksQueue,
		Hasher:                       TestHasher,
		Marshalizer:                  TestMarshalizer,
		ForkDetector:                 tpn.ForkDetector,
		RequestHandler:               tpn.RequestHandler,
		ShardCoordinator:             tpn.ShardCoordinator,
		Accounts:                     tpn.AccntState,
		BlackListHandler:             tpn.BlockBlackListHandler,
		NetworkWatcher:               tpn.MainMessenger,
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
		ProcessWaitTimeSupernova:     tpn.RoundHandler.TimeDuration(),
		RepopulateTokensSupplies:     false,
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EnableRoundsHandler:          tpn.EnableRoundsHandler,
		ProcessConfigsHandler:        tpn.ProcessConfigsHandler,
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
