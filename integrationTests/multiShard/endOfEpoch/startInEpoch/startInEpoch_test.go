package startInEpoch

import (
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/types"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/integrationTests/multiShard/endOfEpoch"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/pendingMb"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/scheduledDataSyncer"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestStartInEpochForAShardNodeInMultiShardedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testNodeStartsInEpoch(t, 0, 18)
}

func TestStartInEpochForAMetaNodeInMultiShardedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testNodeStartsInEpoch(t, core.MetachainShardId, 20)
}

func testNodeStartsInEpoch(t *testing.T, shardID uint32, expectedHighestRound uint64) {
	numOfShards := 2
	numNodesPerShard := 3
	numMetachainNodes := 3

	enableEpochsConfig := config.EnableEpochs{
		StakingV2EnableEpoch:                 integrationTests.UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:       integrationTests.UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch: integrationTests.UnreachableEpoch,
		RefactorPeersMiniBlocksEnableEpoch:   integrationTests.UnreachableEpoch,
		StakingV4Step1EnableEpoch:            integrationTests.UnreachableEpoch,
		StakingV4Step2EnableEpoch:            integrationTests.UnreachableEpoch,
		StakingV4Step3EnableEpoch:            integrationTests.UnreachableEpoch,
	}

	nodes := integrationTests.CreateNodesWithEnableEpochs(
		numOfShards,
		numNodesPerShard,
		numMetachainNodes,
		enableEpochsConfig,
	)

	roundsPerEpoch := uint64(10)
	for _, node := range nodes {
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * numNodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * numNodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(1000000000)
	sendValue := big.NewInt(5)
	integrationTests.MintAllNodes(nodes, initialVal)
	receiverAddress := []byte("12345678901234567890123456789012")

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)

	// ----- wait for epoch end period
	epoch := uint32(2)
	nrRoundsToPropagateMultiShard := uint64(5)
	for i := uint64(0); i <= (uint64(epoch)*roundsPerEpoch)+nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++

		for _, node := range nodes {
			integrationTests.CreateAndSendTransaction(node, nodes, sendValue, receiverAddress, "", integrationTests.AdditionalGasLimit)
		}

		time.Sleep(integrationTests.StepDelay)
	}

	time.Sleep(time.Second)

	endOfEpoch.VerifyThatNodesHaveCorrectEpoch(t, epoch, nodes)
	endOfEpoch.VerifyIfAddedShardHeadersAreWithNewEpoch(t, nodes)

	epochHandler := &mock.EpochStartTriggerStub{
		MetaEpochCalled: func() uint32 {
			return epoch
		},
	}
	for _, node := range nodes {
		_ = dataRetriever.SetEpochHandlerToHdrResolver(node.ResolversContainer, epochHandler)
		_ = dataRetriever.SetEpochHandlerToHdrRequester(node.RequestersContainer, epochHandler)
	}

	generalConfig := getGeneralConfig()
	roundDurationMillis := 4000
	epochDurationMillis := generalConfig.EpochStartConfig.RoundsPerEpoch * int64(roundDurationMillis)
	prefsConfig := config.PreferencesConfig{
		FullArchive: false,
	}

	pksBytes := integrationTests.CreatePkBytes(uint32(numOfShards))
	address := []byte("afafafafafafafafafafafafafafafaf")

	nodesConfig := &testscommon.NodesSetupStub{
		InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			oneMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
			for i := uint32(0); i < uint32(numOfShards); i++ {
				oneMap[i] = append(oneMap[i], mock.NewNodeInfo(address, pksBytes[i], i, integrationTests.InitialRating))
			}
			oneMap[core.MetachainShardId] = append(oneMap[core.MetachainShardId],
				mock.NewNodeInfo(address, pksBytes[core.MetachainShardId], core.MetachainShardId, integrationTests.InitialRating))
			return oneMap, nil
		},
		GetStartTimeCalled: func() int64 {
			return time.Now().Add(-time.Duration(epochDurationMillis) * time.Millisecond).Unix()
		},
		GetRoundDurationCalled: func() uint64 {
			return 4000
		},
		GetShardConsensusGroupSizeCalled: func() uint32 {
			return 1
		},
		GetMetaConsensusGroupSizeCalled: func() uint32 {
			return 1
		},
		NumberOfShardsCalled: func() uint32 {
			return uint32(numOfShards)
		},
	}
	defer func() {
		errRemoveDir := os.RemoveAll("Epoch_0")
		assert.NoError(t, errRemoveDir)

		errRemoveDir = os.RemoveAll("Static")
		assert.NoError(t, errRemoveDir)
	}()

	genesisShardCoordinator, _ := sharding.NewMultiShardCoordinator(nodesConfig.NumberOfShards(), 0)

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	nodeToJoinLate := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            uint32(numOfShards),
		NodeShardId:          shardID,
		TxSignPrivKeyShardId: shardID,
	})
	messenger := integrationTests.CreateMessengerWithNoDiscovery()
	time.Sleep(integrationTests.P2pBootstrapDelay)
	nodeToJoinLate.MainMessenger = messenger

	nodeToJoinLate.FullArchiveMessenger = &p2pmocks.MessengerStub{}

	for _, n := range nodes {
		_ = n.ConnectOnMain(nodeToJoinLate)
	}

	roundHandler := &mock.RoundHandlerMock{IndexField: int64(round)}
	cryptoComponents := integrationTests.GetDefaultCryptoComponents()
	cryptoComponents.PubKey = nodeToJoinLate.NodeKeys.MainKey.Pk
	cryptoComponents.BlockSig = &mock.SignerMock{}
	cryptoComponents.TxSig = &mock.SignerMock{}
	cryptoComponents.BlKeyGen = &mock.KeyGenMock{}
	cryptoComponents.TxKeyGen = &mock.KeyGenMock{}

	coreComponents := integrationTests.GetDefaultCoreComponents(integrationTests.CreateEnableEpochsConfig())
	coreComponents.InternalMarshalizerField = integrationTests.TestMarshalizer
	coreComponents.TxMarshalizerField = integrationTests.TestMarshalizer
	coreComponents.HasherField = integrationTests.TestHasher
	coreComponents.AddressPubKeyConverterField = integrationTests.TestAddressPubkeyConverter
	coreComponents.Uint64ByteSliceConverterField = uint64Converter
	coreComponents.PathHandlerField = &testscommon.PathManagerStub{}
	coreComponents.ChainIdCalled = func() string {
		return string(integrationTests.ChainID)
	}
	coreComponents.NodeTypeProviderField = &nodeTypeProviderMock.NodeTypeProviderStub{}
	coreComponents.ChanStopNodeProcessField = endProcess.GetDummyEndProcessChannel()
	coreComponents.HardforkTriggerPubKeyField = []byte("provided hardfork pub key")

	nodesCoordinatorRegistryFactory, _ := nodesCoordinator.NewNodesCoordinatorRegistryFactory(
		&marshallerMock.MarshalizerMock{},
		444,
	)
	additionalStorageServiceFactory := &testscommon.AdditionalStorageServiceFactoryMock{}

	argsBootstrapHandler := bootstrap.ArgsEpochStartBootstrap{
		NodesCoordinatorRegistryFactory: nodesCoordinatorRegistryFactory,
		CryptoComponentsHolder:          cryptoComponents,
		CoreComponentsHolder:            coreComponents,
		MainMessenger:                   nodeToJoinLate.MainMessenger,
		FullArchiveMessenger:            nodeToJoinLate.FullArchiveMessenger,
		GeneralConfig:                   generalConfig,
		PrefsConfig: config.PreferencesConfig{
			FullArchive: false,
		},
		GenesisShardCoordinator:    genesisShardCoordinator,
		EconomicsData:              nodeToJoinLate.EconomicsData,
		LatestStorageDataProvider:  &mock.LatestStorageDataProviderStub{},
		StorageUnitOpener:          &mock.UnitOpenerStub{},
		GenesisNodesConfig:         nodesConfig,
		Rater:                      &mock.RaterMock{},
		DestinationShardAsObserver: shardID,
		NodeShuffler:               &shardingMocks.NodeShufflerMock{},
		RoundHandler:               roundHandler,
		ArgumentsParser:            smartContract.NewArgumentParser(),
		StatusHandler:              &statusHandlerMock.AppStatusHandlerStub{},
		HeaderIntegrityVerifier:    integrationTests.CreateHeaderIntegrityVerifier(),
		DataSyncerCreator: &scheduledDataSyncer.ScheduledSyncerFactoryStub{
			CreateCalled: func(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
				return &scheduledDataSyncer.ScheduledSyncerStub{
					UpdateSyncDataIfNeededCalled: func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, map[string]*block.MiniBlock, error) {
						return notarizedShardHeader, nil, nil, nil
					},
					GetRootHashToSyncCalled: func(notarizedShardHeader data.ShardHeaderHandler) []byte {
						return notarizedShardHeader.GetRootHash()
					},
				}, nil
			},
		},
		ScheduledSCRsStorer: genericMocks.NewStorerMock(),
		FlagsConfig: config.ContextFlagsConfig{
			ForceStartFromNetwork: false,
		},
		TrieSyncStatisticsProvider:       &testscommon.SizeSyncStatisticsHandlerStub{},
		StateStatsHandler:                disabled.NewStateStatistics(),
		NodesCoordinatorWithRaterFactory: nodesCoordinator.NewIndexHashedNodesCoordinatorWithRaterFactory(),
		ShardCoordinatorFactory:          sharding.NewMultiShardCoordinatorFactory(),
		AdditionalStorageServiceCreator:  additionalStorageServiceFactory,
	}

	epochStartBootstrap, err := bootstrap.NewEpochStartBootstrap(argsBootstrapHandler)
	assert.Nil(t, err)

	bootstrapParams, err := epochStartBootstrap.Bootstrap()
	assert.NoError(t, err)
	assert.Equal(t, bootstrapParams.SelfShardId, shardID)
	assert.Equal(t, bootstrapParams.Epoch, epoch)

	shardC, _ := sharding.NewMultiShardCoordinator(2, shardID)

	storageFactory, err := factory.NewStorageServiceFactory(
		factory.StorageServiceFactoryArgs{
			Config:                          generalConfig,
			PrefsConfig:                     prefsConfig,
			ShardCoordinator:                shardC,
			PathManager:                     &testscommon.PathManagerStub{},
			EpochStartNotifier:              notifier.NewEpochStartSubscriptionHandler(),
			NodeTypeProvider:                &nodeTypeProviderMock.NodeTypeProviderStub{},
			CurrentEpoch:                    0,
			StorageType:                     factory.ProcessStorageService,
			CreateTrieEpochRootHashStorer:   false,
			NodeProcessingMode:              common.Normal,
			ManagedPeersHolder:              &testscommon.ManagedPeersHolderStub{},
			StateStatsHandler:               disabled.NewStateStatistics(),
			AdditionalStorageServiceCreator: additionalStorageServiceFactory,
		},
	)
	assert.NoError(t, err)
	storageServiceShard, err := storageFactory.CreateForMeta()
	assert.NoError(t, err)
	assert.NotNil(t, storageServiceShard)

	bootstrapUnit, _ := storageServiceShard.GetStorer(dataRetriever.BootstrapUnit)
	assert.NotNil(t, bootstrapUnit)

	bootstrapStorer, err := bootstrapStorage.NewBootstrapStorer(integrationTests.TestMarshalizer, bootstrapUnit)
	assert.NoError(t, err)
	assert.NotNil(t, bootstrapStorer)

	argsBaseBootstrapper := storageBootstrap.ArgsBaseStorageBootstrapper{
		BootStorer:     bootstrapStorer,
		ForkDetector:   &mock.ForkDetectorStub{},
		BlockProcessor: &mock.BlockProcessorMock{},
		ChainHandler: &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				if shardID != core.MetachainShardId {
					return &block.Header{}
				} else {
					return &block.MetaBlock{}
				}
			},
		},
		Marshalizer:         integrationTests.TestMarshalizer,
		Store:               storageServiceShard,
		Uint64Converter:     uint64Converter,
		BootstrapRoundIndex: round,
		ShardCoordinator:    shardC,
		NodesCoordinator:    &shardingMocks.NodesCoordinatorMock{},
		EpochStartTrigger:   &mock.EpochStartTriggerStub{},
		BlockTracker: &mock.BlockTrackerStub{
			RestoreToGenesisCalled: func() {},
		},
		ChainID:                      string(integrationTests.ChainID),
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
		EpochNotifier:                &epochNotifierMock.EpochNotifierStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		AppStatusHandler:             &statusHandlerMock.AppStatusHandlerMock{},
	}

	bootstrapper, err := getBootstrapper(shardID, argsBaseBootstrapper)
	assert.NoError(t, err)
	assert.NotNil(t, bootstrapper)

	err = bootstrapper.LoadFromStorage()
	assert.NoError(t, err)

	highestNonce := bootstrapper.GetHighestBlockNonce()
	assert.True(t, highestNonce > expectedHighestRound)
}

func getBootstrapper(shardID uint32, baseArgs storageBootstrap.ArgsBaseStorageBootstrapper) (process.BootstrapperFromStorage, error) {
	if shardID == core.MetachainShardId {
		pendingMiniBlocksHandler, _ := pendingMb.NewPendingMiniBlocks()
		bootstrapperArgs := storageBootstrap.ArgsMetaStorageBootstrapper{
			ArgsBaseStorageBootstrapper: baseArgs,
			PendingMiniBlocksHandler:    pendingMiniBlocksHandler,
		}

		return storageBootstrap.NewMetaStorageBootstrapper(bootstrapperArgs)
	}

	bootstrapperArgs := storageBootstrap.ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: baseArgs}
	return storageBootstrap.NewShardStorageBootstrapper(bootstrapperArgs)
}

func getGeneralConfig() config.Config {
	generalConfig := testscommon.GetGeneralConfig()
	generalConfig.MiniBlocksStorage.DB.Type = string(storageunit.LvlDBSerial)
	generalConfig.ShardHdrNonceHashStorage.DB.Type = string(storageunit.LvlDBSerial)
	generalConfig.MetaBlockStorage.DB.Type = string(storageunit.LvlDBSerial)
	generalConfig.MetaHdrNonceHashStorage.DB.Type = string(storageunit.LvlDBSerial)
	generalConfig.BlockHeaderStorage.DB.Type = string(storageunit.LvlDBSerial)
	generalConfig.BootstrapStorage.DB.Type = string(storageunit.LvlDBSerial)
	generalConfig.ReceiptsStorage.DB.Type = string(storageunit.LvlDBSerial)
	generalConfig.ScheduledSCRsStorage.DB.Type = string(storageunit.LvlDBSerial)

	return generalConfig
}
