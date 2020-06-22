package startInEpoch

import (
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/pendingMb"
	"github.com/ElrondNetwork/elrond-go/process/sync/storageBootstrap"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
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

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		numNodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
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
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000)
	sendValue := big.NewInt(5)
	integrationTests.MintAllNodes(nodes, initialVal)
	receiverAddress := []byte("12345678901234567890123456789012")

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)

	/////////----- wait for epoch end period
	epoch := uint32(2)
	nrRoundsToPropagateMultiShard := uint64(5)
	for i := uint64(0); i <= (uint64(epoch)*roundsPerEpoch)+nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++

		for _, node := range nodes {
			integrationTests.CreateAndSendTransaction(node, sendValue, receiverAddress, "")
		}

		time.Sleep(time.Second)
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

	pksBytes := integrationTests.CreatePkBytes(uint32(numOfShards))
	address := make([]byte, 32)
	address = []byte("afafafafafafafafafafafafafafafaf")

	nodesConfig := &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (m map[uint32][]sharding.GenesisNodeInfoHandler, m2 map[uint32][]sharding.GenesisNodeInfoHandler) {
			oneMap := make(map[uint32][]sharding.GenesisNodeInfoHandler)
			for i := uint32(0); i < uint32(numOfShards); i++ {
				oneMap[i] = append(oneMap[i], mock.NewNodeInfo(address, pksBytes[i], i))
			}
			oneMap[core.MetachainShardId] = append(oneMap[core.MetachainShardId], mock.NewNodeInfo(address, pksBytes[core.MetachainShardId], core.MetachainShardId))
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
	}

	defer func() {
		errRemoveDir := os.RemoveAll("Epoch_0")
		assert.NoError(t, errRemoveDir)
	}()

	genesisShardCoordinator, _ := sharding.NewMultiShardCoordinator(nodesConfig.NumberOfShards(), 0)

	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	nodeToJoinLate := integrationTests.NewTestProcessorNode(uint32(numOfShards), shardID, shardID, "")
	messenger := integrationTests.CreateMessengerWithKadDht(integrationTests.GetConnectableAddress(advertiser))
	_ = messenger.Bootstrap()
	time.Sleep(integrationTests.P2pBootstrapDelay)
	nodeToJoinLate.Messenger = messenger

	rounder := &mock.RounderMock{IndexField: int64(round)}
	argsBootstrapHandler := bootstrap.ArgsEpochStartBootstrap{
		CryptoComponentsHolder: &mock.CryptoComponentsMock{
			PubKey:   nodeToJoinLate.NodeKeys.Pk,
			BlockSig: &mock.SignerMock{},
			TxSig:    &mock.SignerMock{},
			BlKeyGen: &mock.KeyGenMock{},
			TxKeyGen: &mock.KeyGenMock{},
		},
		CoreComponentsHolder: &mock.CoreComponentsMock{
			IntMarsh:            integrationTests.TestMarshalizer,
			TxMarsh:             integrationTests.TestTxSignMarshalizer,
			Hash:                integrationTests.TestHasher,
			UInt64ByteSliceConv: uint64Converter,
			AddrPubKeyConv:      integrationTests.TestAddressPubkeyConverter,
			PathHdl:             &mock.PathManagerStub{},
			ChainIdCalled: func() string {
				return string(integrationTests.ChainID)
			},
		},
		Messenger:                  nodeToJoinLate.Messenger,
		GeneralConfig:              getGeneralConfig(),
		GenesisShardCoordinator:    genesisShardCoordinator,
		EconomicsData:              integrationTests.CreateEconomicsData(),
		LatestStorageDataProvider:  &mock.LatestStorageDataProviderStub{},
		StorageUnitOpener:          &mock.UnitOpenerStub{},
		GenesisNodesConfig:         nodesConfig,
		Rater:                      &mock.RaterMock{},
		DestinationShardAsObserver: shardID,
		NodeShuffler:               &mock.NodeShufflerMock{},
		Rounder:                    rounder,
		ImportStartHandler:         &mock.ImportStartHandlerStub{},
	}
	epochStartBootstrap, err := bootstrap.NewEpochStartBootstrap(argsBootstrapHandler)
	assert.Nil(t, err)

	bootstrapParams, err := epochStartBootstrap.Bootstrap()
	assert.NoError(t, err)
	assert.Equal(t, bootstrapParams.SelfShardId, shardID)
	assert.Equal(t, bootstrapParams.Epoch, epoch)

	shardC, _ := sharding.NewMultiShardCoordinator(2, shardID)

	storageFactory, err := factory.NewStorageServiceFactory(
		&generalConfig,
		shardC,
		&mock.PathManagerStub{},
		&mock.EpochStartNotifierStub{},
		0)
	assert.NoError(t, err)
	storageServiceShard, err := storageFactory.CreateForMeta()
	assert.NoError(t, err)
	assert.NotNil(t, storageServiceShard)

	bootstrapUnit := storageServiceShard.GetStorer(dataRetriever.BootstrapUnit)
	assert.NotNil(t, bootstrapUnit)

	bootstrapStorer, err := bootstrapStorage.NewBootstrapStorer(integrationTests.TestMarshalizer, bootstrapUnit)
	assert.NoError(t, err)
	assert.NotNil(t, bootstrapStorer)

	argsBaseBootstrapper := storageBootstrap.ArgsBaseStorageBootstrapper{
		BootStorer:          bootstrapStorer,
		ForkDetector:        &mock.ForkDetectorStub{},
		BlockProcessor:      &mock.BlockProcessorMock{},
		ChainHandler:        &mock.BlockChainMock{},
		Marshalizer:         integrationTests.TestMarshalizer,
		Store:               storageServiceShard,
		Uint64Converter:     uint64Converter,
		BootstrapRoundIndex: round,
		ShardCoordinator:    shardC,
		NodesCoordinator:    &mock.NodesCoordinatorMock{},
		EpochStartTrigger:   &mock.EpochStartTriggerStub{},
		BlockTracker: &mock.BlockTrackerStub{
			RestoreToGenesisCalled: func() {},
		},
		ChainID: string(integrationTests.ChainID),
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

// TODO: We should remove this type of configs hidden in tests
func getGeneralConfig() config.Config {
	return config.Config{
		GeneralSettings: config.GeneralSettingsConfig{
			StartInEpochEnabled: true,
		},
		EpochStartConfig: config.EpochStartConfig{
			MinRoundsBetweenEpochs:            5,
			RoundsPerEpoch:                    10,
			MinNumConnectedPeersToStart:       2,
			MinNumOfPeersToConsiderBlockValid: 2,
		},
		WhiteListPool: config.CacheConfig{
			Capacity: 10000,
			Type:     "LRU",
			Shards:   1,
		},
		WhiteListerVerifiedTxs: config.CacheConfig{
			Capacity: 10000,
			Type:     "LRU",
			Shards:   1,
		},
		StoragePruning: config.StoragePruningConfig{
			Enabled:             false,
			FullArchive:         true,
			NumEpochsToKeep:     3,
			NumActivePersisters: 3,
		},
		EvictionWaitingList: config.EvictionWaitingListConfig{
			Size: 100,
			DB: config.DBConfig{
				FilePath:          "EvictionWaitingList",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		TrieSnapshotDB: config.DBConfig{
			FilePath:          "TrieSnapshot",
			Type:              "MemoryDB",
			BatchDelaySeconds: 30,
			MaxBatchSize:      6,
			MaxOpenFiles:      10,
		},
		AccountsTrieStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "AccountsTrie/MainDB",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerAccountsTrieStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "PeerAccountsTrie/MainDB",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		StateTriesConfig: config.StateTriesConfig{
			CheckpointRoundsModulus:     100,
			AccountsStatePruningEnabled: false,
			PeerStatePruningEnabled:     false,
			MaxStateTrieLevelInMemory:   5,
			MaxPeerTrieLevelInMemory:    5,
		},
		TrieStorageManagerConfig: config.TrieStorageManagerConfig{
			PruningBufferLen:   1000,
			SnapshotsBufferLen: 10,
			MaxSnapshots:       2,
		},
		TxDataPool: config.CacheConfig{
			Capacity:             10000,
			SizePerSender:        1000,
			SizeInBytes:          1000000000,
			SizeInBytesPerSender: 10000000,
			Shards:               1,
		},
		UnsignedTransactionDataPool: config.CacheConfig{
			Capacity:    10000,
			SizeInBytes: 1000000000,
			Shards:      1,
		},
		RewardTransactionDataPool: config.CacheConfig{
			Capacity:    10000,
			SizeInBytes: 1000000000,
			Shards:      1,
		},
		HeadersPoolConfig: config.HeadersPoolConfig{
			MaxHeadersPerShard:            100,
			NumElementsToRemoveOnEviction: 1,
		},
		TxBlockBodyDataPool: config.CacheConfig{
			Capacity: 10000,
			Type:     "LRU",
			Shards:   1,
		},
		PeerBlockBodyDataPool: config.CacheConfig{
			Capacity: 10000,
			Type:     "LRU",
			Shards:   1,
		},
		TrieNodesDataPool: config.CacheConfig{
			Capacity: 10000,
			Type:     "LRU",
			Shards:   1,
		},
		TxStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "Transactions",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MiniBlocksStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "MiniBlocks",
				Type:              string(storageUnit.LvlDBSerial),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		ShardHdrNonceHashStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "ShardHdrHashNonce",
				Type:              string(storageUnit.LvlDBSerial),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MetaBlockStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "MetaBlock",
				Type:              string(storageUnit.LvlDBSerial),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MetaHdrNonceHashStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "MetaHdrHashNonce",
				Type:              string(storageUnit.LvlDBSerial),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		UnsignedTransactionStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "UnsignedTransactions",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		RewardTxStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "RewardTransactions",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		BlockHeaderStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "BlockHeaders",
				Type:              string(storageUnit.LvlDBSerial),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		Heartbeat: config.HeartbeatConfig{
			HeartbeatStorage: config.StorageConfig{
				Cache: config.CacheConfig{
					Capacity: 10000,
					Type:     "LRU",
					Shards:   1,
				},
				DB: config.DBConfig{
					FilePath:          "HeartbeatStorage",
					Type:              "MemoryDB",
					BatchDelaySeconds: 30,
					MaxBatchSize:      6,
					MaxOpenFiles:      10,
				},
			},
		},
		StatusMetricsStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "StatusMetricsStorageDB",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerBlockBodyStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "PeerBlocks",
				Type:              string(storageUnit.LvlDBSerial),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		BootstrapStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "BootstrapData",
				Type:              string(storageUnit.LvlDBSerial),
				BatchDelaySeconds: 1,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		TxLogsStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Type:     "LRU",
				Capacity: 1000,
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "Logs",
				Type:              string(storageUnit.LvlDBSerial),
				BatchDelaySeconds: 2,
				MaxBatchSize:      100,
				MaxOpenFiles:      10,
			},
		},
	}
}
