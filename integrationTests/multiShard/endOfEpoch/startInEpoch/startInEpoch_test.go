package startInEpoch

import (
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/types"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/pendingMb"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/sync/storageBootstrap"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageunit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	epochNotifierMock "github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/ElrondNetwork/elrond-go/testscommon/scheduledDataSyncer"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
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
	}

	generalConfig := getGeneralConfig()
	roundDurationMillis := 4000
	epochDurationMillis := generalConfig.EpochStartConfig.RoundsPerEpoch * int64(roundDurationMillis)
	prefsConfig := config.PreferencesConfig{
		FullArchive: false,
	}

	pksBytes := integrationTests.CreatePkBytes(uint32(numOfShards))
	address := []byte("afafafafafafafafafafafafafafafaf")

	nodesConfig := &mock.NodesSetupStub{
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
		GetChainIdCalled: func() string {
			return string(integrationTests.ChainID)
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
		GetMinTransactionVersionCalled: func() uint32 {
			return integrationTests.MinTransactionVersion
		},
	}

	defer func() {
		errRemoveDir := os.RemoveAll("Epoch_0")
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
	nodeToJoinLate.Messenger = messenger

	for _, n := range nodes {
		_ = n.ConnectTo(nodeToJoinLate)
	}

	roundHandler := &mock.RoundHandlerMock{IndexField: int64(round)}
	cryptoComponents := integrationTests.GetDefaultCryptoComponents()
	cryptoComponents.PubKey = nodeToJoinLate.NodeKeys.Pk
	cryptoComponents.BlockSig = &mock.SignerMock{}
	cryptoComponents.TxSig = &mock.SignerMock{}
	cryptoComponents.BlKeyGen = &mock.KeyGenMock{}
	cryptoComponents.TxKeyGen = &mock.KeyGenMock{}

	coreComponents := integrationTests.GetDefaultCoreComponents()
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

	argsBootstrapHandler := bootstrap.ArgsEpochStartBootstrap{
		CryptoComponentsHolder: cryptoComponents,
		CoreComponentsHolder:   coreComponents,
		Messenger:              nodeToJoinLate.Messenger,
		GeneralConfig:          generalConfig,
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
					UpdateSyncDataIfNeededCalled: func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error) {
						return notarizedShardHeader, nil, nil
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
		TrieSyncStatisticsProvider: &testscommon.SizeSyncStatisticsHandlerStub{},
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
			Config:                        generalConfig,
			PrefsConfig:                   prefsConfig,
			ShardCoordinator:              shardC,
			PathManager:                   &testscommon.PathManagerStub{},
			EpochStartNotifier:            notifier.NewEpochStartSubscriptionHandler(),
			NodeTypeProvider:              &nodeTypeProviderMock.NodeTypeProviderStub{},
			CurrentEpoch:                  0,
			StorageType:                   factory.ProcessStorageService,
			CreateTrieEpochRootHashStorer: false,
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
