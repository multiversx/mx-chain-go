package staking

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/nodetype"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/blockchain"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	mock3 "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	mock2 "github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	blproc "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	economicsHandler "github.com/ElrondNetwork/elrond-go/process/economics"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	metaProcess "github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
)

// TestMetaProcessor -
type TestMetaProcessor struct {
	MetaBlockProcessor process.BlockProcessor
	SystemSCProcessor  process.EpochStartSystemSCProcessor
	NodesCoordinator   nodesCoordinator.NodesCoordinator
}

// NewTestMetaProcessor -
func NewTestMetaProcessor(
	numOfMetaNodes int,
	numOfShards int,
	numOfNodesPerShard int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
) *TestMetaProcessor {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders(uint32(numOfShards))
	nc := createNodesCoordinator(numOfMetaNodes, numOfShards, numOfNodesPerShard, shardConsensusGroupSize, metaConsensusGroupSize)
	scp := createSystemSCProcessor(nc)
	return &TestMetaProcessor{
		MetaBlockProcessor: createMetaBlockProcessor(nc, scp, coreComponents, dataComponents, bootstrapComponents, statusComponents),
	}
}

// shuffler constants
const (
	shuffleBetweenShards    = false
	adaptivity              = false
	hysteresis              = float32(0.2)
	maxTrieLevelInMemory    = uint(5)
	delegationManagementKey = "delegationManagement"
	delegationContractsList = "delegationContracts"
)

func createSystemSCProcessor(nc nodesCoordinator.NodesCoordinator) process.EpochStartSystemSCProcessor {
	args, _ := createFullArgumentsForSystemSCProcessing(nc, 1000, integrationTests.CreateMemUnit())
	s, _ := metachain.NewSystemSCProcessor(args)
	return s
}

func createNodesCoordinator(
	numOfMetaNodes int,
	numOfShards int,
	numOfNodesPerShard int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
) nodesCoordinator.NodesCoordinator {
	validatorsMap := generateGenesisNodeInfoMap(numOfMetaNodes, numOfShards, numOfNodesPerShard)
	validatorsMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)

	waitingMap := generateGenesisNodeInfoMap(numOfMetaNodes, numOfShards, numOfNodesPerShard)
	waitingMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(waitingMap)

	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:                     uint32(numOfNodesPerShard),
		NodesMeta:                      uint32(numOfMetaNodes),
		Hysteresis:                     hysteresis,
		Adaptivity:                     adaptivity,
		ShuffleBetweenShards:           shuffleBetweenShards,
		MaxNodesEnableConfig:           nil,
		WaitingListFixEnableEpoch:      0,
		BalanceWaitingListsEnableEpoch: 0,
	}
	nodeShuffler, _ := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	epochStartSubscriber := notifier.NewEpochStartSubscriptionHandler()
	bootStorer := integrationTests.CreateMemUnit()

	cache, _ := lrucache.NewCache(10000)
	ncrf, _ := nodesCoordinator.NewNodesCoordinatorRegistryFactory(integrationTests.TestMarshalizer, forking.NewGenericEpochNotifier(), 4444)
	argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ShardConsensusGroupSize:         shardConsensusGroupSize,
		MetaConsensusGroupSize:          metaConsensusGroupSize,
		Marshalizer:                     integrationTests.TestMarshalizer,
		Hasher:                          integrationTests.TestHasher,
		ShardIDAsObserver:               core.MetachainShardId,
		NbShards:                        uint32(numOfShards),
		EligibleNodes:                   validatorsMapForNodesCoordinator,
		WaitingNodes:                    waitingMapForNodesCoordinator,
		SelfPublicKey:                   validatorsMap[core.MetachainShardId][0].PubKeyBytes(),
		ConsensusGroupCache:             cache,
		ShuffledOutHandler:              &mock2.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch:      0,
		ChanStopNode:                    endProcess.GetDummyEndProcessChannel(),
		IsFullArchive:                   false,
		Shuffler:                        nodeShuffler,
		BootStorer:                      bootStorer,
		EpochStartNotifier:              epochStartSubscriber,
		StakingV4EnableEpoch:            444,
		NodesCoordinatorRegistryFactory: ncrf,
		NodeTypeProvider:                nodetype.NewNodeTypeProvider(core.NodeTypeValidator),
	}

	nodesCoordinator, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
	if err != nil {
		fmt.Println("error creating node coordinator")
	}

	return nodesCoordinator
}

func generateGenesisNodeInfoMap(
	numOfMetaNodes int,
	numOfShards int,
	numOfNodesPerShard int,
) map[uint32][]nodesCoordinator.GenesisNodeInfoHandler {
	validatorsMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
	for shardId := 0; shardId < numOfShards; shardId++ {
		for n := 0; n < numOfNodesPerShard; n++ {
			addr := []byte("addr" + strconv.Itoa(n) + "-shard" + strconv.Itoa(shardId))
			validator := mock2.NewNodeInfo(addr, addr, uint32(shardId), 5)
			validatorsMap[uint32(shardId)] = append(validatorsMap[uint32(shardId)], validator)
		}
	}

	for n := 0; n < numOfMetaNodes; n++ {
		addr := []byte("addr" + strconv.Itoa(n) + "-shard" + strconv.Itoa(int(core.MetachainShardId)))
		validator := mock2.NewNodeInfo(addr, addr, uint32(core.MetachainShardId), 5)
		validatorsMap[core.MetachainShardId] = append(validatorsMap[core.MetachainShardId], validator)
	}

	return validatorsMap
}

func createMetaBlockProcessor(
	nc nodesCoordinator.NodesCoordinator,
	systemSCProcessor process.EpochStartSystemSCProcessor,
	coreComponents *mock.CoreComponentsMock,
	dataComponents *mock.DataComponentsMock,
	bootstrapComponents *mock.BootstrapComponentsMock,
	statusComponents *mock.StatusComponentsMock,
) process.BlockProcessor {
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents, nc, systemSCProcessor)

	metaProc, _ := blproc.NewMetaProcessor(arguments)
	return metaProc
}

func createMockComponentHolders(numOfShards uint32) (
	*mock.CoreComponentsMock,
	*mock.DataComponentsMock,
	*mock.BootstrapComponentsMock,
	*mock.StatusComponentsMock,
) {
	coreComponents := &mock.CoreComponentsMock{
		IntMarsh:            &mock.MarshalizerMock{},
		Hash:                &mock.HasherStub{},
		UInt64ByteSliceConv: &mock.Uint64ByteSliceConverterMock{},
		StatusField:         &statusHandlerMock.AppStatusHandlerStub{},
		RoundField:          &mock.RoundHandlerMock{RoundTimeDuration: time.Second},
	}

	dataComponents := &mock.DataComponentsMock{
		Storage:  &mock.ChainStorerMock{},
		DataPool: dataRetrieverMock.NewPoolsHolderMock(),
		BlockChain: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{Nonce: 0}
			},
		},
	}
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(numOfShards, core.MetachainShardId)
	boostrapComponents := &mock.BootstrapComponentsMock{
		Coordinator:          shardCoordinator,
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		VersionedHdrFactory: &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32) data.HeaderHandler {
				return &block.MetaBlock{}
			},
		},
	}

	statusComponents := &mock.StatusComponentsMock{
		Outport: &testscommon.OutportStub{},
	}

	return coreComponents, dataComponents, boostrapComponents, statusComponents
}

func createMockMetaArguments(
	coreComponents *mock.CoreComponentsMock,
	dataComponents *mock.DataComponentsMock,
	bootstrapComponents *mock.BootstrapComponentsMock,
	statusComponents *mock.StatusComponentsMock,
	nodesCoord nodesCoordinator.NodesCoordinator,
	systemSCProcessor process.EpochStartSystemSCProcessor,
) blproc.ArgMetaProcessor {

	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:      &mock.HasherStub{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)

	startHeaders := createGenesisBlocks(bootstrapComponents.ShardCoordinator())
	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = &stateMock.AccountsStub{
		CommitCalled: func() ([]byte, error) {
			return nil, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return nil, nil
		},
	}
	accountsDb[state.PeerAccountsState] = &stateMock.AccountsStub{
		CommitCalled: func() ([]byte, error) {
			return nil, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return nil, nil
		},
	}

	arguments := blproc.ArgMetaProcessor{
		ArgBaseProcessor: blproc.ArgBaseProcessor{
			CoreComponents:      coreComponents,
			DataComponents:      dataComponents,
			BootstrapComponents: bootstrapComponents,
			StatusComponents:    statusComponents,
			AccountsDB:          accountsDb,
			ForkDetector:        &mock.ForkDetectorMock{},
			NodesCoordinator:    nodesCoord,
			FeeHandler:          &mock.FeeAccumulatorStub{},
			RequestHandler:      &testscommon.RequestHandlerStub{},
			BlockChainHook:      &testscommon.BlockChainHookStub{},
			TxCoordinator:       &mock.TransactionCoordinatorMock{},
			EpochStartTrigger:   &mock.EpochStartTriggerStub{},
			HeaderValidator:     headerValidator,
			GasHandler:          &mock.GasHandlerMock{},
			BootStorer: &mock.BoostrapStorerMock{
				PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
					return nil
				},
			},
			BlockTracker:                   mock.NewBlockTrackerMock(bootstrapComponents.ShardCoordinator(), startHeaders),
			BlockSizeThrottler:             &mock.BlockSizeThrottlerStub{},
			HistoryRepository:              &dblookupext.HistoryRepositoryStub{},
			EpochNotifier:                  &epochNotifier.EpochNotifierStub{},
			RoundNotifier:                  &mock.RoundNotifierStub{},
			ScheduledTxsExecutionHandler:   &testscommon.ScheduledTxsExecutionStub{},
			ScheduledMiniBlocksEnableEpoch: 2,
		},
		SCToProtocol:                 &mock.SCToProtocolStub{},
		PendingMiniBlocksHandler:     &mock.PendingMiniBlocksHandlerStub{},
		EpochStartDataCreator:        &mock.EpochStartDataCreatorStub{},
		EpochEconomics:               &mock.EpochEconomicsStub{},
		EpochRewardsCreator:          &testscommon.RewardsCreatorStub{},
		EpochValidatorInfoCreator:    &testscommon.EpochValidatorInfoCreatorStub{},
		ValidatorStatisticsProcessor: &testscommon.ValidatorStatisticsProcessorStub{},
		EpochSystemSCProcessor:       systemSCProcessor,
	}
	return arguments
}

func createGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for ShardID := uint32(0); ShardID < shardCoordinator.NumberOfShards(); ShardID++ {
		genesisBlocks[ShardID] = createGenesisBlock(ShardID)
	}

	genesisBlocks[core.MetachainShardId] = createGenesisMetaBlock()

	return genesisBlocks
}

func createGenesisBlock(ShardID uint32) *block.Header {
	rootHash := []byte("roothash")
	return &block.Header{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		ShardID:       ShardID,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
	}
}

func createGenesisMetaBlock() *block.MetaBlock {
	rootHash := []byte("roothash")
	return &block.MetaBlock{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
	}
}

func createFullArgumentsForSystemSCProcessing(nc nodesCoordinator.NodesCoordinator, stakingV2EnableEpoch uint32, trieStorer storage.Storer) (metachain.ArgsNewEpochStartSystemSCProcessing, vm.SystemSCContainer) {
	hasher := sha256.NewSha256()
	marshalizer := &marshal.GogoProtoMarshalizer{}
	trieFactoryManager, _ := trie.NewTrieStorageManagerWithoutPruning(trieStorer)
	userAccountsDB := createAccountsDB(hasher, marshalizer, factory.NewAccountCreator(), trieFactoryManager)
	peerAccountsDB := createAccountsDB(hasher, marshalizer, factory.NewPeerAccountCreator(), trieFactoryManager)
	en := forking.NewGenericEpochNotifier()

	argsValidatorsProcessor := peer.ArgValidatorStatisticsProcessor{
		Marshalizer:                          marshalizer,
		NodesCoordinator:                     nc,
		ShardCoordinator:                     &mock.ShardCoordinatorStub{},
		DataPool:                             &dataRetrieverMock.PoolsHolderStub{},
		StorageService:                       &mock3.ChainStorerStub{},
		PubkeyConv:                           &mock.PubkeyConverterMock{},
		PeerAdapter:                          peerAccountsDB,
		Rater:                                &mock3.RaterStub{},
		RewardsHandler:                       &mock3.RewardsHandlerStub{},
		NodesSetup:                           &mock.NodesSetupStub{},
		MaxComputableRounds:                  1,
		MaxConsecutiveRoundsOfRatingDecrease: 2000,
		EpochNotifier:                        en,
		StakingV2EnableEpoch:                 stakingV2EnableEpoch,
		StakingV4EnableEpoch:                 444,
	}
	vCreator, _ := peer.NewValidatorStatisticsProcessor(argsValidatorsProcessor)

	blockChain, _ := blockchain.NewMetaChain(&statusHandlerMock.AppStatusHandlerStub{})
	gasSchedule := arwenConfig.MakeGasMapForTests()
	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasSchedule)
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:     gasScheduleNotifier,
		MapDNSAddresses: make(map[string]struct{}),
		Marshalizer:     marshalizer,
		Accounts:        userAccountsDB,
		ShardCoordinator: &mock.ShardCoordinatorStub{SelfIdCalled: func() uint32 {
			return core.MetachainShardId
		}},
		EpochNotifier: &epochNotifier.EpochNotifierStub{},
	}
	builtInFuncs, _, _ := builtInFunctions.CreateBuiltInFuncContainerAndNFTStorageHandler(argsBuiltIn)

	testDataPool := dataRetrieverMock.NewPoolsHolderMock()
	argsHook := hooks.ArgBlockChainHook{
		Accounts:           userAccountsDB,
		PubkeyConv:         &mock.PubkeyConverterMock{},
		StorageService:     &mock3.ChainStorerStub{},
		BlockChain:         blockChain,
		ShardCoordinator:   &mock.ShardCoordinatorStub{},
		Marshalizer:        marshalizer,
		Uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		NFTStorageHandler:  &testscommon.SimpleNFTStorageHandlerStub{},
		BuiltInFunctions:   builtInFuncs,
		DataPool:           testDataPool,
		CompiledSCPool:     testDataPool.SmartContracts(),
		EpochNotifier:      &epochNotifier.EpochNotifierStub{},
		NilCompiledSCStore: true,
	}

	defaults.FillGasMapInternal(gasSchedule, 1)
	signVerifer, _ := disabled.NewMessageSignVerifier(&cryptoMocks.KeyGenStub{})

	nodesSetup := &mock.NodesSetupStub{}

	blockChainHookImpl, _ := hooks.NewBlockChainHookImpl(argsHook)
	argsNewVMContainerFactory := metaProcess.ArgsNewVMContainerFactory{
		BlockChainHook:      blockChainHookImpl,
		PubkeyConv:          argsHook.PubkeyConv,
		Economics:           createEconomicsData(),
		MessageSignVerifier: signVerifer,
		GasSchedule:         gasScheduleNotifier,
		NodesConfigProvider: nodesSetup,
		Hasher:              hasher,
		Marshalizer:         marshalizer,
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost:  "1000",
				OwnerAddress:     "aaaaaa",
				DelegationTicker: "DEL",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        "50",
					MinPassThreshold: "50",
					MinVetoThreshold: "50",
				},
				FirstWhitelistedAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             5,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
				StakeLimitPercentage:                 100.0,
				NodeLimitPercentage:                  100.0,
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "100",
				MinStakeAmount:      "100",
				ConfigChangeAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 100,
			},
		},
		ValidatorAccountsDB: peerAccountsDB,
		ChanceComputer:      &mock3.ChanceComputerStub{},
		EpochNotifier:       en,
		EpochConfig: &config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				StakingV2EnableEpoch:               stakingV2EnableEpoch,
				StakeEnableEpoch:                   0,
				DelegationManagerEnableEpoch:       0,
				DelegationSmartContractEnableEpoch: 0,
				StakeLimitsEnableEpoch:             10,
				StakingV4InitEnableEpoch:           444,
				StakingV4EnableEpoch:               445,
			},
		},
		ShardCoordinator: &mock.ShardCoordinatorStub{},
		NodesCoordinator: nc,
	}
	metaVmFactory, _ := metaProcess.NewVMContainerFactory(argsNewVMContainerFactory)

	vmContainer, _ := metaVmFactory.Create()
	systemVM, _ := vmContainer.Get(vmFactory.SystemVirtualMachine)

	stakingSCprovider, _ := metachain.NewStakingDataProvider(systemVM, "1000")
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, core.MetachainShardId)

	args := metachain.ArgsNewEpochStartSystemSCProcessing{
		SystemVM:                systemVM,
		UserAccountsDB:          userAccountsDB,
		PeerAccountsDB:          peerAccountsDB,
		Marshalizer:             marshalizer,
		StartRating:             5,
		ValidatorInfoCreator:    vCreator,
		EndOfEpochCallerAddress: vm.EndOfEpochAddress,
		StakingSCAddress:        vm.StakingSCAddress,
		ChanceComputer:          &mock3.ChanceComputerStub{},
		EpochNotifier:           en,
		GenesisNodesConfig:      nodesSetup,
		StakingDataProvider:     stakingSCprovider,
		NodesConfigProvider: &shardingMocks.NodesCoordinatorStub{
			ConsensusGroupSizeCalled: func(shardID uint32) int {
				if shardID == core.MetachainShardId {
					return 400
				}
				return 63
			},
		},
		ShardCoordinator:      shardCoordinator,
		ESDTOwnerAddressBytes: bytes.Repeat([]byte{1}, 32),
		EpochConfig: config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				StakingV2EnableEpoch:     1000000,
				ESDTEnableEpoch:          1000000,
				StakingV4InitEnableEpoch: 444,
				StakingV4EnableEpoch:     445,
			},
		},
	}
	return args, metaVmFactory.SystemSmartContractContainer()
}

func createAccountsDB(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountFactory state.AccountFactory,
	trieStorageManager common.StorageManager,
) *state.AccountsDB {
	tr, _ := trie.NewTrie(trieStorageManager, marshalizer, hasher, 5)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(10, testscommon.NewMemDbMock(), marshalizer)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)
	adb, _ := state.NewAccountsDB(tr, hasher, marshalizer, accountFactory, spm, common.Normal)
	return adb
}

func createEconomicsData() process.EconomicsDataHandler {
	maxGasLimitPerBlock := strconv.FormatUint(1500000000, 10)
	minGasPrice := strconv.FormatUint(10, 10)
	minGasLimit := strconv.FormatUint(10, 10)

	argsNewEconomicsData := economicsHandler.ArgsNewEconomicsData{
		Economics: &config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				GenesisTotalSupply: "2000000000000000000000",
				MinimumInflation:   0,
				YearSettings: []*config.YearSetting{
					{
						Year:             0,
						MaximumInflation: 0.01,
					},
				},
			},
			RewardsSettings: config.RewardsSettings{
				RewardsConfigByEpoch: []config.EpochRewardSettings{
					{
						LeaderPercentage:                 0.1,
						DeveloperPercentage:              0.1,
						ProtocolSustainabilityPercentage: 0.1,
						ProtocolSustainabilityAddress:    "protocol",
						TopUpGradientPoint:               "300000000000000000000",
						TopUpFactor:                      0.25,
					},
				},
			},
			FeeSettings: config.FeeSettings{
				GasLimitSettings: []config.GasLimitSetting{
					{
						MaxGasLimitPerBlock:         maxGasLimitPerBlock,
						MaxGasLimitPerMiniBlock:     maxGasLimitPerBlock,
						MaxGasLimitPerMetaBlock:     maxGasLimitPerBlock,
						MaxGasLimitPerMetaMiniBlock: maxGasLimitPerBlock,
						MaxGasLimitPerTx:            maxGasLimitPerBlock,
						MinGasLimit:                 minGasLimit,
					},
				},
				MinGasPrice:      minGasPrice,
				GasPerDataByte:   "1",
				GasPriceModifier: 1.0,
			},
		},
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  &epochNotifier.EpochNotifierStub{},
		BuiltInFunctionsCostHandler:    &mock.BuiltInCostHandlerStub{},
	}
	economicsData, _ := economicsHandler.NewEconomicsData(argsNewEconomicsData)
	return economicsData
}
