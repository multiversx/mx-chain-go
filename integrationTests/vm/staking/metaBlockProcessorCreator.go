package staking

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/factory"
	integrationMocks "github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	blproc "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/postprocess"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/scToProtocol"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
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

	bootStorer, _ := bootstrapStorage.NewBootstrapStorer(
		coreComponents.InternalMarshalizer(),
		dataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
	)

	headerValidator := createHeaderValidator(coreComponents)
	valInfoCreator := createValidatorInfoCreator(coreComponents, dataComponents, bootstrapComponents.ShardCoordinator())
	stakingToPeer := createSCToProtocol(coreComponents, stateComponents, dataComponents.Datapool().CurrentBlockTxs())

	args := blproc.ArgMetaProcessor{
		ArgBaseProcessor: blproc.ArgBaseProcessor{
			CoreComponents:                 coreComponents,
			DataComponents:                 dataComponents,
			BootstrapComponents:            bootstrapComponents,
			StatusComponents:               statusComponents,
			AccountsDB:                     accountsDb,
			ForkDetector:                   &integrationMocks.ForkDetectorStub{},
			NodesCoordinator:               nc,
			FeeHandler:                     postprocess.NewFeeAccumulator(),
			RequestHandler:                 &testscommon.RequestHandlerStub{},
			BlockChainHook:                 blockChainHook,
			TxCoordinator:                  txCoordinator,
			EpochStartTrigger:              epochStartHandler,
			HeaderValidator:                headerValidator,
			GasHandler:                     &mock.GasHandlerMock{},
			BootStorer:                     bootStorer,
			BlockTracker:                   blockTracker,
			BlockSizeThrottler:             &mock.BlockSizeThrottlerStub{},
			HistoryRepository:              &dblookupext.HistoryRepositoryStub{},
			EpochNotifier:                  coreComponents.EpochNotifier(),
			RoundNotifier:                  &mock.RoundNotifierStub{},
			ScheduledTxsExecutionHandler:   &testscommon.ScheduledTxsExecutionStub{},
			ScheduledMiniBlocksEnableEpoch: 10000,
			VMContainersFactory:            metaVMFactory,
			VmContainer:                    vmContainer,
			ProcessedMiniBlocksTracker:     processedMb.NewProcessedMiniBlocksTracker(),
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
	}

	metaProc, _ := blproc.NewMetaProcessor(args)
	return metaProc
}

func createValidatorInfoCreator(
	coreComponents factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	shardCoordinator sharding.Coordinator,
) process.EpochStartValidatorInfoCreator {
	args := metachain.ArgsNewValidatorInfoCreator{
		ShardCoordinator: shardCoordinator,
		MiniBlockStorage: dataComponents.StorageService().GetStorer(dataRetriever.MiniBlockUnit),
		Hasher:           coreComponents.Hasher(),
		Marshalizer:      coreComponents.InternalMarshalizer(),
		DataPool:         dataComponents.Datapool(),
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
		Marshalizer:       coreComponents.InternalMarshalizer(),
		Hasher:            coreComponents.Hasher(),
		Store:             dataComponents.StorageService(),
		DataPool:          dataComponents.Datapool(),
		BlockTracker:      blockTracker,
		ShardCoordinator:  shardCoordinator,
		EpochStartTrigger: epochStartTrigger,
		RequestHandler:    &testscommon.RequestHandlerStub{},
		GenesisEpoch:      0,
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
		Hasher:      coreComponents.Hasher(),
		Marshalizer: coreComponents.InternalMarshalizer(),
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
		PubkeyConv:         coreComponents.AddressPubKeyConverter(),
		Hasher:             coreComponents.Hasher(),
		Marshalizer:        coreComponents.InternalMarshalizer(),
		PeerState:          stateComponents.PeerAccounts(),
		BaseState:          stateComponents.AccountsAdapter(),
		ArgParser:          smartContract.NewArgumentParser(),
		CurrTxs:            txCacher,
		RatingsData:        &mock.RatingsInfoMock{},
		EpochNotifier:      coreComponents.EpochNotifier(),
		StakingV4InitEpoch: stakingV4InitEpoch,
	}
	stakingToPeer, _ := scToProtocol.NewStakingToPeer(args)
	return stakingToPeer
}
