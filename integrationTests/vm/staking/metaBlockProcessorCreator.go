package staking

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/executionTrack"
	"github.com/multiversx/mx-chain-go/process/estimator"
	"github.com/multiversx/mx-chain-go/process/missingData"

	"github.com/multiversx/mx-chain-go/process/block/headerForBlock"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/containers"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/factory"
	integrationMocks "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/postprocess"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/scToProtocol"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/disabled"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	testsFactory "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func createMetaBlockProcessor(
	nc nodesCoordinator.NodesCoordinator,
	systemSCProcessor process.EpochStartSystemSCProcessor,
	coreComponents factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	bootstrapComponents factory.BootstrapComponentsHolder,
	statusComponents factory.StatusComponentsHolder,
	stateComponents factory.StateComponentsHandler,
	validatorsInfoCreator process.ValidatorStatisticsProcessor,
	blockChainHook process.BlockChainHookHandler,
	metaVMFactory process.VirtualMachinesContainerFactory,
	epochStartHandler process.EpochStartTriggerHandler,
	vmContainer process.VirtualMachinesContainer,
	txCoordinator process.TransactionCoordinator,
) process.BlockProcessor {
	blockTracker := createBlockTracker(
		dataComponents.Blockchain().GetGenesisHeader(),
		bootstrapComponents.ShardCoordinator(),
	)
	epochStartDataCreator := createEpochStartDataCreator(
		coreComponents,
		dataComponents,
		bootstrapComponents.ShardCoordinator(),
		epochStartHandler,
		blockTracker,
	)

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = stateComponents.AccountsAdapter()
	accountsDb[state.PeerAccountsState] = stateComponents.PeerAccounts()

	bootStrapStorer, _ := dataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit)
	bootStorer, _ := bootstrapStorage.NewBootstrapStorer(
		coreComponents.InternalMarshalizer(),
		bootStrapStorer,
	)

	headerValidator := createHeaderValidator(coreComponents)
	valInfoCreator := createValidatorInfoCreator(coreComponents, dataComponents, bootstrapComponents.ShardCoordinator())
	stakingToPeer := createSCToProtocol(coreComponents, stateComponents, dataComponents.Datapool().CurrentBlockTxs())

	headersForBlock, _ := headerForBlock.NewHeadersForBlock(headerForBlock.ArgHeadersForBlock{
		DataPool:            dataComponents.Datapool(),
		RequestHandler:      &testscommon.RequestHandlerStub{},
		EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
		ShardCoordinator:    bootstrapComponents.ShardCoordinator(),
		BlockTracker:        blockTracker,
		TxCoordinator:       txCoordinator,
		RoundHandler:        coreComponents.RoundHandler(),
		ExtraDelayForRequestBlockInfoInMilliseconds: 100,
		GenesisNonce: 0,
	})

	preprocessors := containers.NewPreProcessorsContainer()
	blockDataRequesterArgs := coordinator.BlockDataRequestArgs{
		RequestHandler:      &testscommon.RequestHandlerStub{},
		MiniBlockPool:       dataComponents.Datapool().MiniBlocks(),
		PreProcessors:       preprocessors,
		ShardCoordinator:    bootstrapComponents.ShardCoordinator(),
		EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
	}
	// second instance for proposal missing data fetching to avoid interferences
	proposalBlockDataRequester, _ := coordinator.NewBlockDataRequester(blockDataRequesterArgs)

	mbSelectionSession, _ := blproc.NewMiniBlocksSelectionSession(
		bootstrapComponents.ShardCoordinator().SelfId(),
		coreComponents.InternalMarshalizer(),
		coreComponents.Hasher(),
	)

	executionResultsTracker := executionTrack.NewExecutionResultsTracker()
	execResultsVerifier, _ := blproc.NewExecutionResultsVerifier(dataComponents.Blockchain(), executionResultsTracker)
	inclusionEstimator := estimator.NewExecutionResultInclusionEstimator(
		config.ExecutionResultInclusionEstimatorConfig{
			SafetyMargin:       110,
			MaxResultsPerBlock: 20,
		},
		coreComponents.RoundHandler(),
	)

	missingDataArgs := missingData.ResolverArgs{
		HeadersPool:        dataComponents.Datapool().Headers(),
		ProofsPool:         dataComponents.Datapool().Proofs(),
		RequestHandler:     &testscommon.RequestHandlerStub{},
		BlockDataRequester: proposalBlockDataRequester,
	}
	missingDataResolver, _ := missingData.NewMissingDataResolver(missingDataArgs)

	shardInfoCreator, _ := blproc.NewShardInfoCreateData(
		coreComponents.EnableEpochsHandler(),
		dataComponents.Datapool().Headers(),
		dataComponents.Datapool().Proofs(),
		&mock.PendingMiniBlocksHandlerStub{},
		blockTracker,
	)

	args := blproc.ArgMetaProcessor{
		ArgBaseProcessor: blproc.ArgBaseProcessor{
			CoreComponents:      coreComponents,
			DataComponents:      dataComponents,
			BootstrapComponents: bootstrapComponents,
			StatusComponents:    statusComponents,
			StatusCoreComponents: &testsFactory.StatusCoreComponentsStub{
				AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
			},
			AccountsDB:                         accountsDb,
			AccountsProposal:                   stateComponents.AccountsAdapterProposal(),
			ForkDetector:                       &integrationMocks.ForkDetectorStub{},
			NodesCoordinator:                   nc,
			FeeHandler:                         postprocess.NewFeeAccumulator(),
			RequestHandler:                     &testscommon.RequestHandlerStub{},
			BlockChainHook:                     blockChainHook,
			TxCoordinator:                      txCoordinator,
			EpochStartTrigger:                  epochStartHandler,
			HeaderValidator:                    headerValidator,
			BootStorer:                         bootStorer,
			BlockTracker:                       blockTracker,
			BlockSizeThrottler:                 &mock.BlockSizeThrottlerStub{},
			HistoryRepository:                  &dblookupext.HistoryRepositoryStub{},
			VMContainersFactory:                metaVMFactory,
			VmContainer:                        vmContainer,
			GasHandler:                         &mock.GasHandlerMock{},
			ScheduledTxsExecutionHandler:       &testscommon.ScheduledTxsExecutionStub{},
			ScheduledMiniBlocksEnableEpoch:     10000,
			ProcessedMiniBlocksTracker:         processedMb.NewProcessedMiniBlocksTracker(),
			OutportDataProvider:                &outport.OutportDataProviderStub{},
			ReceiptsRepository:                 &testscommon.ReceiptsRepositoryStub{},
			ManagedPeersHolder:                 &testscommon.ManagedPeersHolderStub{},
			BlockProcessingCutoffHandler:       &testscommon.BlockProcessingCutoffStub{},
			SentSignaturesTracker:              &testscommon.SentSignatureTrackerStub{},
			StateAccessesCollector:             disabled.NewDisabledStateAccessesCollector(),
			HeadersForBlock:                    headersForBlock,
			MiniBlocksSelectionSession:         mbSelectionSession,
			ExecutionResultsVerifier:           execResultsVerifier,
			MissingDataResolver:                missingDataResolver,
			ExecutionResultsInclusionEstimator: inclusionEstimator,
			ExecutionResultsTracker:            executionResultsTracker,
			GasComputation:                     &testscommon.GasComputationMock{},
		},
		SCToProtocol:             stakingToPeer,
		PendingMiniBlocksHandler: &mock.PendingMiniBlocksHandlerStub{},
		EpochStartDataCreator:    epochStartDataCreator,
		EpochEconomics:           &mock.EpochEconomicsStub{},
		EpochRewardsCreator: &testscommon.RewardsCreatorStub{
			GetLocalTxCacheCalled: func() epochStart.TransactionCacher {
				return dataComponents.Datapool().CurrentBlockTxs()
			},
		},
		EpochValidatorInfoCreator:    valInfoCreator,
		ValidatorStatisticsProcessor: validatorsInfoCreator,
		EpochSystemSCProcessor:       systemSCProcessor,
		ShardInfoCreator:             shardInfoCreator,
	}

	metaProc, _ := blproc.NewMetaProcessor(args)
	return metaProc
}

func createValidatorInfoCreator(
	coreComponents factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	shardCoordinator sharding.Coordinator,
) process.EpochStartValidatorInfoCreator {
	mbStorer, _ := dataComponents.StorageService().GetStorer(dataRetriever.MiniBlockUnit)

	args := metachain.ArgsNewValidatorInfoCreator{
		ShardCoordinator:     shardCoordinator,
		MiniBlockStorage:     mbStorer,
		Hasher:               coreComponents.Hasher(),
		Marshalizer:          coreComponents.InternalMarshalizer(),
		DataPool:             dataComponents.Datapool(),
		EnableEpochsHandler:  coreComponents.EnableEpochsHandler(),
		ValidatorInfoStorage: integrationtests.CreateMemUnit(),
	}

	valInfoCreator, _ := metachain.NewValidatorInfoCreator(args)
	return valInfoCreator
}

func createEpochStartDataCreator(
	coreComponents factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	shardCoordinator sharding.Coordinator,
	epochStartTrigger process.EpochStartTriggerHandler,
	blockTracker process.BlockTracker,
) process.EpochStartDataCreator {
	argsEpochStartDataCreator := metachain.ArgsNewEpochStartData{
		Marshalizer:         coreComponents.InternalMarshalizer(),
		Hasher:              coreComponents.Hasher(),
		Store:               dataComponents.StorageService(),
		DataPool:            dataComponents.Datapool(),
		BlockTracker:        blockTracker,
		ShardCoordinator:    shardCoordinator,
		EpochStartTrigger:   epochStartTrigger,
		RequestHandler:      &testscommon.RequestHandlerStub{},
		GenesisEpoch:        0,
		EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
	}
	epochStartDataCreator, _ := metachain.NewEpochStartData(argsEpochStartDataCreator)
	return epochStartDataCreator
}

func createBlockTracker(
	genesisMetaHeader data.HeaderHandler,
	shardCoordinator sharding.Coordinator,
) process.BlockTracker {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for ShardID := uint32(0); ShardID < shardCoordinator.NumberOfShards(); ShardID++ {
		genesisBlocks[ShardID] = createGenesisBlock(ShardID)
	}

	genesisBlocks[core.MetachainShardId] = genesisMetaHeader
	return mock.NewBlockTrackerMock(shardCoordinator, genesisBlocks)
}

func createGenesisBlock(shardID uint32) *block.Header {
	rootHash := []byte("roothash")
	return &block.Header{
		Nonce:           0,
		Round:           0,
		Signature:       rootHash,
		RandSeed:        rootHash,
		PrevRandSeed:    rootHash,
		ShardID:         shardID,
		PubKeysBitmap:   rootHash,
		RootHash:        rootHash,
		PrevHash:        rootHash,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
}

func createGenesisMetaBlock() *block.MetaBlock {
	rootHash := []byte("roothash")
	return &block.MetaBlock{
		Nonce:                  0,
		Round:                  0,
		Signature:              rootHash,
		RandSeed:               rootHash,
		PrevRandSeed:           rootHash,
		PubKeysBitmap:          rootHash,
		RootHash:               rootHash,
		PrevHash:               rootHash,
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
	}
}

func createHeaderValidator(coreComponents factory.CoreComponentsHolder) epochStart.HeaderValidator {
	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:              coreComponents.Hasher(),
		Marshalizer:         coreComponents.InternalMarshalizer(),
		EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)
	return headerValidator
}

func createSCToProtocol(
	coreComponents factory.CoreComponentsHolder,
	stateComponents factory.StateComponentsHandler,
	txCacher dataRetriever.TransactionCacher,
) process.SmartContractToProtocolHandler {
	args := scToProtocol.ArgStakingToPeer{
		PubkeyConv:          coreComponents.AddressPubKeyConverter(),
		Hasher:              coreComponents.Hasher(),
		Marshalizer:         coreComponents.InternalMarshalizer(),
		PeerState:           stateComponents.PeerAccounts(),
		BaseState:           stateComponents.AccountsAdapter(),
		ArgParser:           smartContract.NewArgumentParser(),
		CurrTxs:             txCacher,
		RatingsData:         &mock.RatingsInfoMock{},
		EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
	}
	stakingToPeer, _ := scToProtocol.NewStakingToPeer(args)
	return stakingToPeer
}
