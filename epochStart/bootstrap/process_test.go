package bootstrap

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/stretchr/testify/assert"
)

func createPkBytes(numShards uint32) map[uint32][]byte {
	pksbytes := make(map[uint32][]byte, numShards+1)
	for i := uint32(0); i < numShards; i++ {
		pksbytes[i] = make([]byte, 128)
		pksbytes[i] = []byte("afafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafaf")
		pksbytes[i][0] = byte(i)
	}

	pksbytes[core.MetachainShardId] = make([]byte, 128)
	pksbytes[core.MetachainShardId] = []byte("afafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafafaf")
	pksbytes[core.MetachainShardId][0] = byte(numShards)

	return pksbytes
}

func createMockEpochStartBootstrapArgs() ArgsEpochStartBootstrap {
	return ArgsEpochStartBootstrap{
		PublicKey: &mock.PublicKeyStub{
			ToByteArrayStub: func() ([]byte, error) {
				return []byte("pubKey"), nil
			},
		},
		Marshalizer:       &mock.MarshalizerMock{},
		TxSignMarshalizer: &mock.MarshalizerMock{},
		Hasher:            &mock.HasherMock{},
		Messenger:         &mock.MessengerStub{},
		GeneralConfig: config.Config{
			WhiteListPool: config.CacheConfig{
				Type:     "LRU",
				Capacity: 10,
				Shards:   10,
			},
			EpochStartConfig: config.EpochStartConfig{
				MinNumConnectedPeersToStart:       2,
				MinNumOfPeersToConsiderBlockValid: 2,
			},
			StateTriesConfig: config.StateTriesConfig{
				CheckpointRoundsModulus:     5,
				AccountsStatePruningEnabled: true,
				PeerStatePruningEnabled:     true,
				MaxStateTrieLevelInMemory:   5,
				MaxPeerTrieLevelInMemory:    5,
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
			TrieStorageManagerConfig: config.TrieStorageManagerConfig{
				PruningBufferLen:   1000,
				SnapshotsBufferLen: 10,
				MaxSnapshots:       2,
			},
			TrieSync: config.TrieSyncConfig{
				NumConcurrentTrieSyncers:  50,
				MaxHardCapForMissingNodes: 500,
				TrieSyncerVersion:         2,
			},
		},
		EconomicsData:              &economicsmocks.EconomicsHandlerStub{},
		SingleSigner:               &mock.SignerStub{},
		BlockSingleSigner:          &mock.SignerStub{},
		KeyGen:                     &mock.KeyGenMock{},
		BlockKeyGen:                &mock.KeyGenMock{},
		GenesisNodesConfig:         &mock.NodesSetupStub{},
		GenesisShardCoordinator:    mock.NewMultipleShardsCoordinatorMock(),
		PathManager:                &mock.PathManagerStub{},
		WorkingDir:                 "test_directory",
		DefaultDBPath:              "test_db",
		DefaultEpochString:         "test_epoch",
		DefaultShardString:         "test_shard",
		Rater:                      &mock.RaterStub{},
		DestinationShardAsObserver: 0,
		Uint64Converter:            &mock.Uint64ByteSliceConverterMock{},
		NodeShuffler:               &mock.NodeShufflerMock{},
		Rounder:                    &mock.RounderStub{},
		AddressPubkeyConverter:     &mock.PubkeyConverterMock{},
		LatestStorageDataProvider:  &mock.LatestStorageDataProviderStub{},
		StorageUnitOpener:          &mock.UnitOpenerStub{},
		ArgumentsParser:            &mock.ArgumentParserMock{},
		StatusHandler:              &mock.AppStatusHandlerStub{},
		HeaderIntegrityVerifier:    &mock.HeaderIntegrityVerifierStub{},
		TxSignHasher:               &mock.HasherMock{},
		EpochNotifier:              &mock.EpochNotifierStub{},
	}
}

func TestNewEpochStartBootstrap(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartBootstrapArgs()

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(epochStartProvider))
}

func TestNewEpochStartBootstrap_NilTxSignHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartBootstrapArgs()
	args.TxSignHasher = nil

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, epochStartProvider)
	assert.True(t, errors.Is(err, epochStart.ErrNilHasher))
}

func TestNewEpochStartBootstrap_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartBootstrapArgs()
	args.EpochNotifier = nil

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, epochStartProvider)
	assert.True(t, errors.Is(err, epochStart.ErrNilEpochNotifier))
}

func TestNewEpochStartBootstrap_InvalidMaxHardCapForMissingNodesShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartBootstrapArgs()
	args.GeneralConfig.TrieSync.MaxHardCapForMissingNodes = 0

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, epochStartProvider)
	assert.True(t, errors.Is(err, epochStart.ErrInvalidMaxHardCapForMissingNodes))
}

func TestNewEpochStartBootstrap_InvalidNumConcurrentTrieSyncersShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartBootstrapArgs()
	args.GeneralConfig.TrieSync.NumConcurrentTrieSyncers = 0

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, epochStartProvider)
	assert.True(t, errors.Is(err, epochStart.ErrInvalidNumConcurrentTrieSyncers))
}

func TestIsStartInEpochZero(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartBootstrapArgs()
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		GetStartTimeCalled: func() int64 {
			return 1000
		},
	}

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	result := epochStartProvider.isStartInEpochZero()
	assert.False(t, result)
}

func TestEpochStartBootstrap_BootstrapStartInEpochNotEnabled(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()

	err := errors.New("localErr")
	args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetCalled: func() (storage.LatestDataFromStorage, error) {
			return storage.LatestDataFromStorage{}, err
		},
	}
	epochStartProvider, _ := NewEpochStartBootstrap(args)

	params, err := epochStartProvider.Bootstrap()
	assert.Nil(t, err)
	assert.NotNil(t, params)
}

func TestEpochStartBootstrap_Bootstrap(t *testing.T) {
	roundsPerEpoch := int64(100)
	roundDuration := uint64(60000)
	args := createMockEpochStartBootstrapArgs()
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		GetRoundDurationCalled: func() uint64 {
			return roundDuration
		},
	}
	args.GeneralConfig = testscommon.GetGeneralConfig()
	args.GeneralConfig.EpochStartConfig.RoundsPerEpoch = roundsPerEpoch
	epochStartProvider, _ := NewEpochStartBootstrap(args)

	done := make(chan bool, 1)

	go func() {
		_, _ = epochStartProvider.Bootstrap()
		<-done
	}()

	for {
		select {
		case <-done:
			assert.Fail(t, "should not be reach")
		case <-time.After(time.Second):
			assert.True(t, true, "pass with timeout")
			return
		}
	}
}

func TestPrepareForEpochZero(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	params, err := epochStartProvider.prepareEpochZero()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), params.Epoch)
}

func TestPrepareForEpochZero_NodeInGenesisShouldNotAlterShardID(t *testing.T) {
	shardIDAsValidator := uint32(1)

	args := createMockEpochStartBootstrapArgs()
	args.GenesisShardCoordinator = &mock.ShardCoordinatorStub{
		SelfIdCalled: func() uint32 {
			return shardIDAsValidator
		},
	}
	args.DestinationShardAsObserver = uint32(7)
	args.PublicKey = &mock.PublicKeyStub{
		ToByteArrayStub: func() ([]byte, error) {
			return []byte("pubKey11"), nil
		},
	}
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
			eligibleMap := map[uint32][]sharding.GenesisNodeInfoHandler{
				1: {mock.NewNodeInfo([]byte("addr"), []byte("pubKey11"), 1, initRating)},
			}
			return eligibleMap, nil
		},
	}

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	params, err := epochStartProvider.prepareEpochZero()
	assert.NoError(t, err)
	assert.Equal(t, shardIDAsValidator, params.SelfShardId)
}

func TestPrepareForEpochZero_NodeNotInGenesisShouldAlterShardID(t *testing.T) {
	desiredShardAsObserver := uint32(7)

	args := createMockEpochStartBootstrapArgs()
	args.GenesisShardCoordinator = &mock.ShardCoordinatorStub{
		SelfIdCalled: func() uint32 {
			return uint32(1)
		},
	}
	args.DestinationShardAsObserver = desiredShardAsObserver
	args.PublicKey = &mock.PublicKeyStub{
		ToByteArrayStub: func() ([]byte, error) {
			return []byte("pubKeyNotInGenesis"), nil
		},
	}
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
			eligibleMap := map[uint32][]sharding.GenesisNodeInfoHandler{
				1: {mock.NewNodeInfo([]byte("addr"), []byte("pubKey11"), 1, initRating)},
			}
			return eligibleMap, nil
		},
	}

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	params, err := epochStartProvider.prepareEpochZero()
	assert.NoError(t, err)
	assert.Equal(t, desiredShardAsObserver, params.SelfShardId)
}

func TestCreateSyncers(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.shardCoordinator = mock.NewMultipleShardsCoordinatorMock()
	epochStartProvider.dataPool = &testscommon.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		MiniBlocksCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
		TrieNodesCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
	}
	epochStartProvider.whiteListHandler = &mock.WhiteListHandlerStub{}
	epochStartProvider.whiteListerVerifiedTxs = &mock.WhiteListHandlerStub{}
	epochStartProvider.requestHandler = &mock.RequestHandlerStub{}

	err := epochStartProvider.createSyncers()
	assert.Nil(t, err)
}

func TestSyncHeadersFrom_MockHeadersSyncerShouldSyncHeaders(t *testing.T) {
	hdrHash1 := []byte("hdrHash1")
	hdrHash2 := []byte("hdrHash2")
	header1 := &block.Header{}
	header2 := &block.MetaBlock{}

	args := createMockEpochStartBootstrapArgs()
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.headersSyncer = &mock.HeadersByHashSyncerStub{
		SyncMissingHeadersByHashCalled: func(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error {
			return nil
		},
		GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
			return map[string]data.HeaderHandler{
				string(hdrHash1): header1,
				string(hdrHash2): header2,
			}, nil
		},
	}

	metaBlock := &block.MetaBlock{
		Epoch: 2,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{HeaderHash: hdrHash1, ShardID: 0},
			},
			Economics: block.Economics{
				PrevEpochStartHash: hdrHash2,
			},
		},
	}

	headers, err := epochStartProvider.syncHeadersFrom(metaBlock)
	assert.Nil(t, err)
	assert.Equal(t, header1, headers[string(hdrHash1)])
	assert.Equal(t, header2, headers[string(hdrHash2)])
}

func TestSyncPeerAccountsState_NilRequestHandlerErr(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.dataPool = &testscommon.PoolsHolderStub{
		TrieNodesCalled: func() storage.Cacher {
			return &testscommon.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, true
				},
			}
		},
	}
	_ = epochStartProvider.createTriesComponentsForShardId(args.GenesisShardCoordinator.SelfId())
	rootHash := []byte("rootHash")
	err := epochStartProvider.syncPeerAccountsState(rootHash)
	assert.Equal(t, state.ErrNilRequestHandler, err)
}

func TestCreateTriesForNewShardID(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()
	args.GeneralConfig = testscommon.GetGeneralConfig()
	epochStartProvider, _ := NewEpochStartBootstrap(args)

	err := epochStartProvider.createTriesComponentsForShardId(1)
	assert.Nil(t, err)
}

func TestSyncUserAccountsState(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.shardCoordinator = mock.NewMultipleShardsCoordinatorMock()
	epochStartProvider.dataPool = &testscommon.PoolsHolderStub{
		TrieNodesCalled: func() storage.Cacher {
			return &testscommon.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, true
				},
			}
		},
	}
	_ = epochStartProvider.createTriesComponentsForShardId(args.GenesisShardCoordinator.SelfId())
	rootHash := []byte("rootHash")
	err := epochStartProvider.syncUserAccountsState(rootHash)
	assert.Equal(t, state.ErrNilRequestHandler, err)
}

func TestRequestAndProcessForShard(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()

	hdrHash1 := []byte("hdrHash1")
	header1 := &block.Header{}
	metaBlock := &block.MetaBlock{
		Epoch: 2,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{HeaderHash: hdrHash1, ShardID: 0},
			},
		},
	}

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 0

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.syncedHeaders = make(map[string]data.HeaderHandler)
	epochStartProvider.miniBlocksSyncer = &mock.PendingMiniBlockSyncHandlerStub{}
	epochStartProvider.headersSyncer = &mock.HeadersByHashSyncerStub{
		GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
			return map[string]data.HeaderHandler{
				string(hdrHash1): header1,
			}, nil
		},
	}
	epochStartProvider.dataPool = &testscommon.PoolsHolderStub{
		TrieNodesCalled: func() storage.Cacher {
			return &testscommon.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, true
				},
			}
		},
	}

	epochStartProvider.shardCoordinator = shardCoordinator
	epochStartProvider.epochStartMeta = metaBlock
	_ = epochStartProvider.createTriesComponentsForShardId(shardCoordinator.SelfId())
	err := epochStartProvider.requestAndProcessForShard()
	assert.Equal(t, state.ErrNilRequestHandler, err)
}

func getNodesConfigMock(numOfShards uint32) sharding.GenesisNodesSetupHandler {
	pksBytes := createPkBytes(numOfShards)
	address := make([]byte, 32)
	address = []byte("afafafafafafafafafafafafafafafaf")

	roundDurationMillis := 4000
	epochDurationMillis := 50 * int64(roundDurationMillis)

	nodesConfig := &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (m map[uint32][]sharding.GenesisNodeInfoHandler, m2 map[uint32][]sharding.GenesisNodeInfoHandler) {
			oneMap := make(map[uint32][]sharding.GenesisNodeInfoHandler)
			for i := uint32(0); i < numOfShards; i++ {
				oneMap[i] = append(oneMap[i], mock.NewNodeInfo(address, pksBytes[i], i, initRating))
			}
			oneMap[core.MetachainShardId] = append(oneMap[core.MetachainShardId], mock.NewNodeInfo(address, pksBytes[core.MetachainShardId], core.MetachainShardId, initRating))
			return oneMap, nil
		},
		GetStartTimeCalled: func() int64 {
			return time.Now().Add(-time.Duration(epochDurationMillis) * time.Millisecond).Unix()
		},
		GetRoundDurationCalled: func() uint64 {
			return 4000
		},
		GetChainIdCalled: func() string {
			return "chainId"
		},
		GetShardConsensusGroupSizeCalled: func() uint32 {
			return 1
		},
		GetMetaConsensusGroupSizeCalled: func() uint32 {
			return 1
		},
		NumberOfShardsCalled: func() uint32 {
			return numOfShards
		},
	}

	return nodesConfig
}

func TestRequestAndProcessing(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()
	args.GeneralConfig.StoragePruning.CleanOldEpochsData = true
	args.PublicKey = &mock.PublicKeyMock{}
	args.GenesisNodesConfig = getNodesConfigMock(1)

	hdrHash1 := []byte("hdrHash1")
	hdrHash2 := []byte("hdrHash2")
	header1 := &block.Header{}
	header2 := &block.MetaBlock{
		Epoch: 1,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{HeaderHash: hdrHash1, ShardID: 0},
			},
			Economics: block.Economics{
				PrevEpochStartHash: hdrHash1,
			},
		},
	}
	metaBlock := &block.MetaBlock{
		Epoch: 0,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{HeaderHash: hdrHash1, ShardID: 0},
			},
			Economics: block.Economics{
				PrevEpochStartHash: hdrHash2,
			},
		},
	}

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.epochStartMeta = metaBlock
	epochStartProvider.headersSyncer = &mock.HeadersByHashSyncerStub{
		GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
			return map[string]data.HeaderHandler{
				string(hdrHash1): header1,
				string(hdrHash2): header2,
			}, nil
		},
	}
	epochStartProvider.dataPool = &testscommon.PoolsHolderStub{
		MiniBlocksCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
		TrieNodesCalled: func() storage.Cacher {
			return &testscommon.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, true
				},
			}
		},
	}
	epochStartProvider.requestHandler = &mock.RequestHandlerStub{}
	epochStartProvider.miniBlocksSyncer = &mock.PendingMiniBlockSyncHandlerStub{}

	params, err := epochStartProvider.requestAndProcessing()
	assert.Equal(t, Parameters{}, params)
	assert.Equal(t, storage.ErrInvalidNumberOfEpochsToSave, err)
}

func TestEpochStartBootstrap_WithDisabledShardIDAsOBserver(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartBootstrapArgs()
	args.DestinationShardAsObserver = core.DisabledShardIDAsObserver
	args.GenesisNodesConfig = getNodesConfigMock(2)

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(epochStartProvider))

	epochStartProvider.dataPool = &testscommon.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return testscommon.NewShardedDataStub()
		},
		MiniBlocksCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
		TrieNodesCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
	}
	epochStartProvider.requestHandler = &mock.RequestHandlerStub{}
	epochStartProvider.epochStartMeta = &block.MetaBlock{Epoch: 0}
	err = epochStartProvider.processNodesConfig([]byte("something"))
	assert.Nil(t, err)
}
