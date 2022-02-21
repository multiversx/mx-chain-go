package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/versioning"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/types"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	epochStartMocks "github.com/ElrondNetwork/elrond-go/testscommon/bootstrapMocks/epochStart"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/ElrondNetwork/elrond-go/testscommon/scheduledDataSyncer"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	storageMocks "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon/syncer"
	"github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func createComponentsForEpochStart() (*mock.CoreComponentsMock, *mock.CryptoComponentsMock) {
	return &mock.CoreComponentsMock{
			IntMarsh:              &mock.MarshalizerMock{},
			Marsh:                 &mock.MarshalizerMock{},
			Hash:                  &hashingMocks.HasherMock{},
			TxSignHasherField:     &hashingMocks.HasherMock{},
			UInt64ByteSliceConv:   &mock.Uint64ByteSliceConverterMock{},
			AddrPubKeyConv:        &mock.PubkeyConverterMock{},
			PathHdl:               &testscommon.PathManagerStub{},
			EpochNotifierField:    &epochNotifier.EpochNotifierStub{},
			TxVersionCheckField:   versioning.NewTxVersionChecker(1),
			NodeTypeProviderField: &nodeTypeProviderMock.NodeTypeProviderStub{},
		}, &mock.CryptoComponentsMock{
			PubKey:   &cryptoMocks.PublicKeyStub{},
			BlockSig: &cryptoMocks.SignerStub{},
			TxSig:    &cryptoMocks.SignerStub{},
			BlKeyGen: &cryptoMocks.KeyGenStub{},
			TxKeyGen: &cryptoMocks.KeyGenStub{},
		}
}

func createMockEpochStartBootstrapArgs(
	coreMock *mock.CoreComponentsMock,
	cryptoMock *mock.CryptoComponentsMock,
) ArgsEpochStartBootstrap {
	generalCfg := testscommon.GetGeneralConfig()
	return ArgsEpochStartBootstrap{
		ScheduledSCRsStorer:    genericMocks.NewStorerMock("path", 0),
		CoreComponentsHolder:   coreMock,
		CryptoComponentsHolder: cryptoMock,
		Messenger:              &mock.MessengerStub{},
		GeneralConfig: config.Config{
			MiniBlocksStorage:                  generalCfg.MiniBlocksStorage,
			PeerBlockBodyStorage:               generalCfg.PeerBlockBodyStorage,
			BlockHeaderStorage:                 generalCfg.BlockHeaderStorage,
			TxStorage:                          generalCfg.TxStorage,
			UnsignedTransactionStorage:         generalCfg.UnsignedTransactionStorage,
			RewardTxStorage:                    generalCfg.RewardTxStorage,
			ShardHdrNonceHashStorage:           generalCfg.ShardHdrNonceHashStorage,
			MetaHdrNonceHashStorage:            generalCfg.MetaHdrNonceHashStorage,
			StatusMetricsStorage:               generalCfg.StatusMetricsStorage,
			ReceiptsStorage:                    generalCfg.ReceiptsStorage,
			SmartContractsStorage:              generalCfg.SmartContractsStorage,
			SmartContractsStorageForSCQuery:    generalCfg.SmartContractsStorageForSCQuery,
			TrieEpochRootHashStorage:           generalCfg.TrieEpochRootHashStorage,
			BootstrapStorage:                   generalCfg.BootstrapStorage,
			MetaBlockStorage:                   generalCfg.MetaBlockStorage,
			AccountsTrieStorageOld:             generalCfg.AccountsTrieStorageOld,
			PeerAccountsTrieStorageOld:         generalCfg.PeerAccountsTrieStorageOld,
			AccountsTrieStorage:                generalCfg.AccountsTrieStorage,
			PeerAccountsTrieStorage:            generalCfg.PeerAccountsTrieStorage,
			AccountsTrieCheckpointsStorage:     generalCfg.AccountsTrieCheckpointsStorage,
			PeerAccountsTrieCheckpointsStorage: generalCfg.PeerAccountsTrieCheckpointsStorage,
			Heartbeat:                          generalCfg.Heartbeat,
			TrieSnapshotDB: config.DBConfig{
				FilePath:          "TrieSnapshot",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
			EvictionWaitingList: config.EvictionWaitingListConfig{
				HashesSize:     100,
				RootHashesSize: 100,
				DB: config.DBConfig{
					FilePath:          "EvictionWaitingList",
					Type:              "MemoryDB",
					BatchDelaySeconds: 30,
					MaxBatchSize:      6,
					MaxOpenFiles:      10,
				},
			},
			StateTriesConfig: config.StateTriesConfig{
				CheckpointRoundsModulus:     5,
				AccountsStatePruningEnabled: true,
				PeerStatePruningEnabled:     true,
				MaxStateTrieLevelInMemory:   5,
				MaxPeerTrieLevelInMemory:    5,
			},
			TrieStorageManagerConfig: config.TrieStorageManagerConfig{
				PruningBufferLen:      1000,
				SnapshotsBufferLen:    10,
				MaxSnapshots:          2,
				SnapshotsGoroutineNum: 1,
			},
			WhiteListPool: config.CacheConfig{
				Type:     "LRU",
				Capacity: 10,
				Shards:   10,
			},
			EpochStartConfig: config.EpochStartConfig{
				MinNumConnectedPeersToStart:       2,
				MinNumOfPeersToConsiderBlockValid: 2,
			},
			StoragePruning: config.StoragePruningConfig{
				Enabled:                     true,
				ValidatorCleanOldEpochsData: true,
				ObserverCleanOldEpochsData:  true,
				NumEpochsToKeep:             2,
				NumActivePersisters:         2,
			},
			TrieSync: config.TrieSyncConfig{
				NumConcurrentTrieSyncers:  50,
				MaxHardCapForMissingNodes: 500,
				TrieSyncerVersion:         2,
			},
			ScheduledSCRsStorage: config.StorageConfig{
				Cache: config.CacheConfig{
					Type:     "LRU",
					Capacity: 10,
					Shards:   10,
				},
				DB: config.DBConfig{
					FilePath:          "scheduledSCRs",
					Type:              "MemoryDB",
					BatchDelaySeconds: 30,
					MaxBatchSize:      6,
					MaxOpenFiles:      10,
				},
			},
			TxDataPool: config.CacheConfig{
				Type:     "LRU",
				Capacity: 10,
				Shards:   10,
			},
		},
		EconomicsData: &economicsmocks.EconomicsHandlerStub{
			MinGasPriceCalled: func() uint64 {
				return 1
			},
		},
		GenesisNodesConfig:         &mock.NodesSetupStub{},
		GenesisShardCoordinator:    mock.NewMultipleShardsCoordinatorMock(),
		Rater:                      &mock.RaterStub{},
		DestinationShardAsObserver: 0,
		NodeShuffler:               &shardingMocks.NodeShufflerMock{},
		RoundHandler:               &mock.RoundHandlerStub{},
		LatestStorageDataProvider:  &mock.LatestStorageDataProviderStub{},
		StorageUnitOpener:          &storageMocks.UnitOpenerStub{},
		ArgumentsParser:            &mock.ArgumentParserMock{},
		StatusHandler:              &statusHandlerMock.AppStatusHandlerStub{},
		HeaderIntegrityVerifier:    &mock.HeaderIntegrityVerifierStub{},
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
	}
}

func TestNewEpochStartBootstrap(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(epochStartProvider))
}

func TestNewEpochStartBootstrap_NilTxSignHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	coreComp.TxSignHasherField = nil
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, epochStartProvider)
	assert.True(t, errors.Is(err, epochStart.ErrNilHasher))
}

func TestNewEpochStartBootstrap_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	coreComp.EpochNotifierField = nil
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, epochStartProvider)
	assert.True(t, errors.Is(err, epochStart.ErrNilEpochNotifier))
}

func TestNewEpochStartBootstrap_InvalidMaxHardCapForMissingNodesShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GeneralConfig.TrieSync.MaxHardCapForMissingNodes = 0

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, epochStartProvider)
	assert.True(t, errors.Is(err, epochStart.ErrInvalidMaxHardCapForMissingNodes))
}

func TestNewEpochStartBootstrap_InvalidNumConcurrentTrieSyncersShouldErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GeneralConfig.TrieSync.NumConcurrentTrieSyncers = 0

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, epochStartProvider)
	assert.True(t, errors.Is(err, epochStart.ErrInvalidNumConcurrentTrieSyncers))
}

func TestIsStartInEpochZero(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
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
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

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

func TestEpochStartBootstrap_BootstrapShouldStartBootstrapProcess(t *testing.T) {
	roundsPerEpoch := int64(100)
	roundDuration := uint64(60000)
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		GetRoundDurationCalled: func() uint64 {
			return roundDuration
		},
	}
	args.GeneralConfig = testscommon.GetGeneralConfig()
	args.GeneralConfig.EpochStartConfig.RoundsPerEpoch = roundsPerEpoch
	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)

	done := make(chan bool, 1)

	go func() {
		_, err = epochStartProvider.Bootstrap()
		require.Nil(t, err)
		<-done
	}()

	for {
		select {
		case <-done:
			assert.Fail(t, "should not be reach")
		case <-time.After(time.Second):
			return
		}
	}
}

func TestPrepareForEpochZero(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	params, err := epochStartProvider.prepareEpochZero()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), params.Epoch)
}

func TestPrepareForEpochZero_NodeInGenesisShouldNotAlterShardID(t *testing.T) {
	shardIDAsValidator := uint32(1)

	coreComp, cryptoComp := createComponentsForEpochStart()
	cryptoComp.PubKey = &cryptoMocks.PublicKeyStub{
		ToByteArrayStub: func() ([]byte, error) {
			return []byte("pubKey11"), nil
		},
	}
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GenesisShardCoordinator = &mock.ShardCoordinatorStub{
		SelfIdCalled: func() uint32 {
			return shardIDAsValidator
		},
		NumberOfShardsCalled: func() uint32 {
			return 2
		},
	}

	args.DestinationShardAsObserver = uint32(7)
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			eligibleMap := map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
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

	coreComp, cryptoComp := createComponentsForEpochStart()
	cryptoComp.PubKey = &cryptoMocks.PublicKeyStub{
		ToByteArrayStub: func() ([]byte, error) {
			return []byte("pubKeyNotInGenesis"), nil
		},
	}
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GenesisShardCoordinator = &mock.ShardCoordinatorStub{
		SelfIdCalled: func() uint32 {
			return uint32(1)
		},
		NumberOfShardsCalled: func() uint32 {
			return 2
		},
	}
	args.DestinationShardAsObserver = desiredShardAsObserver
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			eligibleMap := map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
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
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.shardCoordinator = mock.NewMultipleShardsCoordinatorMock()
	epochStartProvider.dataPool = &dataRetrieverMock.PoolsHolderStub{
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
	epochStartProvider.whiteListHandler = &testscommon.WhiteListHandlerStub{}
	epochStartProvider.whiteListerVerifiedTxs = &testscommon.WhiteListHandlerStub{}
	epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}

	err := epochStartProvider.createSyncers()
	assert.Nil(t, err)
}

func TestSyncHeadersFrom_MockHeadersSyncerShouldSyncHeaders(t *testing.T) {
	hdrHash1 := []byte("hdrHash1")
	hdrHash2 := []byte("hdrHash2")
	header1 := &block.Header{}
	header2 := &block.MetaBlock{}

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
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

func TestSyncValidatorAccountsState_NilRequestHandlerErr(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.dataPool = &dataRetrieverMock.PoolsHolderStub{
		TrieNodesCalled: func() storage.Cacher {
			return &testscommon.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, true
				},
			}
		},
	}
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		args.GeneralConfig,
		coreComp,
		args.GenesisShardCoordinator.SelfId(),
		disabled.NewChainStorer(),
		0,
		coreComp.EpochNotifier(),
	)
	assert.Nil(t, err)
	epochStartProvider.trieContainer = triesContainer
	epochStartProvider.trieStorageManagers = trieStorageManagers

	rootHash := []byte("rootHash")
	err = epochStartProvider.syncValidatorAccountsState(rootHash)
	assert.Equal(t, state.ErrNilRequestHandler, err)
}

func TestCreateTriesForNewShardID(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GeneralConfig = testscommon.GetGeneralConfig()

	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		args.GeneralConfig,
		coreComp,
		1,
		disabled.NewChainStorer(),
		0,
		coreComp.EpochNotifier(),
	)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(triesContainer.GetAll()))
	assert.Equal(t, 2, len(trieStorageManagers))
}

func TestSyncUserAccountsState(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.shardCoordinator = mock.NewMultipleShardsCoordinatorMock()
	epochStartProvider.dataPool = &dataRetrieverMock.PoolsHolderStub{
		TrieNodesCalled: func() storage.Cacher {
			return &testscommon.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, true
				},
			}
		},
	}

	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		args.GeneralConfig,
		coreComp,
		args.GenesisShardCoordinator.SelfId(),
		disabled.NewChainStorer(),
		0,
		coreComp.EpochNotifier(),
	)
	assert.Nil(t, err)
	epochStartProvider.trieContainer = triesContainer
	epochStartProvider.trieStorageManagers = trieStorageManagers

	rootHash := []byte("rootHash")
	err = epochStartProvider.syncUserAccountsState(rootHash)
	assert.Equal(t, state.ErrNilRequestHandler, err)
}

func TestRequestAndProcessForShard(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

	notarizedShardHeaderHash := []byte("notarizedShardHeaderHash")
	prevShardHeaderHash := []byte("prevShardHeaderHash")
	prevShardHeader := &block.Header{}
	notarizedShardHeader := &block.Header{
		PrevHash: prevShardHeaderHash,
	}
	metaBlock := &block.MetaBlock{
		Epoch: 2,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{HeaderHash: notarizedShardHeaderHash, ShardID: 0},
			},
		},
	}

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 0

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.syncedHeaders = make(map[string]data.HeaderHandler)
	epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}
	epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
		GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
			return map[string]data.HeaderHandler{
				string(notarizedShardHeaderHash): notarizedShardHeader,
				string(prevShardHeaderHash):      prevShardHeader,
			}, nil
		},
	}
	epochStartProvider.dataPool = &dataRetrieverMock.PoolsHolderStub{
		TrieNodesCalled: func() storage.Cacher {
			return &testscommon.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, true
				},
			}
		},
	}

	epochStartProvider.txSyncerForScheduled = &syncer.TransactionsSyncHandlerMock{
		SyncTransactionsForCalled: func(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error {
			return nil
		},
		GetTransactionsCalled: func() (map[string]data.TransactionHandler, error) {
			return nil, nil
		},
	}

	epochStartProvider.shardCoordinator = shardCoordinator
	epochStartProvider.epochStartMeta = metaBlock
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		args.GeneralConfig,
		coreComp,
		shardCoordinator.SelfId(),
		disabled.NewChainStorer(),
		0,
		coreComp.EpochNotifier(),
	)
	assert.Nil(t, err)
	epochStartProvider.trieContainer = triesContainer
	epochStartProvider.trieStorageManagers = trieStorageManagers

	err = epochStartProvider.requestAndProcessForShard()
	assert.Equal(t, state.ErrNilRequestHandler, err)
}

func getNodesConfigMock(numOfShards uint32) sharding.GenesisNodesSetupHandler {
	pksBytes := createPkBytes(numOfShards)
	address := []byte("afafafafafafafafafafafafafafafaf")

	roundDurationMillis := 4000
	epochDurationMillis := 50 * int64(roundDurationMillis)

	nodesConfig := &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			oneMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
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
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GeneralConfig.StoragePruning.ObserverCleanOldEpochsData = true
	args.GeneralConfig.StoragePruning.ValidatorCleanOldEpochsData = true
	args.GeneralConfig.StoragePruning.NumActivePersisters = 0
	args.GenesisNodesConfig = getNodesConfigMock(1)

	prevPrevEpochStartMetaHeaderHash := []byte("prevPrevEpochStartMetaHeaderHash")
	prevEpochStartMetaHeaderHash := []byte("prevEpochStartMetaHeaderHash")
	prevEpochNotarizedShardHeaderHash := []byte("prevEpochNotarizedShardHeaderHash")
	notarizedShardHeaderHash := []byte("notarizedShardHeaderHash")
	epochStartMetaBlockHash := []byte("epochStartMetaBlockHash")
	prevNotarizedShardHeaderHash := []byte("prevNotarizedShardHeaderHash")
	notarizedShardHeader := &block.Header{
		PrevHash: prevNotarizedShardHeaderHash,
	}
	prevNotarizedShardHeader := &block.Header{}

	epochStartMetaBlock := &block.MetaBlock{
		Epoch: 0,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{HeaderHash: notarizedShardHeaderHash, ShardID: 0},
			},
			Economics: block.Economics{
				PrevEpochStartHash: prevEpochStartMetaHeaderHash,
			},
		},
	}
	prevEpochStartMetaBlock := &block.MetaBlock{
		Epoch: 0,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{HeaderHash: prevEpochNotarizedShardHeaderHash, ShardID: 0},
			},
			Economics: block.Economics{
				PrevEpochStartHash: prevPrevEpochStartMetaHeaderHash,
			},
		},
	}

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.epochStartMeta = epochStartMetaBlock
	epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
		GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
			return map[string]data.HeaderHandler{
				string(notarizedShardHeaderHash):     notarizedShardHeader,
				string(prevEpochStartMetaHeaderHash): prevEpochStartMetaBlock,
				string(epochStartMetaBlockHash):      epochStartMetaBlock,
				string(prevNotarizedShardHeaderHash): prevNotarizedShardHeader,
			}, nil
		},
	}
	epochStartProvider.dataPool = &dataRetrieverMock.PoolsHolderStub{
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
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
	}
	epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
	epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}
	epochStartProvider.txSyncerForScheduled = &syncer.TransactionsSyncHandlerMock{}

	params, err := epochStartProvider.requestAndProcessing()
	assert.Equal(t, Parameters{}, params)
	assert.Equal(t, storage.ErrInvalidNumberOfActivePersisters, err)
}

func TestEpochStartBootstrap_WithDisabledShardIDAsObserver(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.DestinationShardAsObserver = common.DisabledShardIDAsObserver
	args.GenesisNodesConfig = getNodesConfigMock(2)

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(epochStartProvider))

	epochStartProvider.dataPool = &dataRetrieverMock.PoolsHolderStub{
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
	epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
	epochStartProvider.epochStartMeta = &block.MetaBlock{Epoch: 0}
	epochStartProvider.prevEpochStartMeta = &block.MetaBlock{}
	err = epochStartProvider.processNodesConfig([]byte("something"))
	assert.Nil(t, err)
}

func TestEpochStartBootstrap_updateDataForScheduledNoScheduledRootHash_UpdateSyncDataIfNeededWithError(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.DestinationShardAsObserver = common.DisabledShardIDAsObserver
	args.GenesisNodesConfig = getNodesConfigMock(2)
	expectedErr := fmt.Errorf("expected error")
	args.DataSyncerCreator = &scheduledDataSyncer.ScheduledSyncerFactoryStub{
		CreateCalled: func(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
			return &scheduledDataSyncer.ScheduledSyncerStub{
				UpdateSyncDataIfNeededCalled: func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error) {
					return nil, nil, expectedErr
				},
				GetRootHashToSyncCalled: func(notarizedShardHeader data.ShardHeaderHandler) []byte {
					return notarizedShardHeader.GetRootHash()
				},
			}, nil
		},
	}

	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)

	notarizedShardHdr := &block.HeaderV2{
		Header:            nil,
		ScheduledRootHash: nil,
	}

	syncData, err := epochStartProvider.updateDataForScheduled(notarizedShardHdr)
	require.Equal(t, expectedErr, err)
	require.Nil(t, syncData)
}

func TestEpochStartBootstrap_updateDataForScheduled_ScheduledTxExecutionCreationWithErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.DestinationShardAsObserver = common.DisabledShardIDAsObserver
	args.GenesisNodesConfig = getNodesConfigMock(2)

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	notarizedShardHdr := &block.HeaderV2{
		Header:            nil,
		ScheduledRootHash: nil,
	}
	epochStartProvider.storerScheduledSCRs = nil

	syncData, err := epochStartProvider.updateDataForScheduled(notarizedShardHdr)
	require.Nil(t, syncData)
	require.Equal(t, process.ErrNilStorage, err)
}

func TestEpochStartBootstrap_updateDataForScheduled_ScheduledSyncerCreateWithError(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.DestinationShardAsObserver = common.DisabledShardIDAsObserver
	args.GenesisNodesConfig = getNodesConfigMock(2)

	expectedError := fmt.Errorf("expected error")
	args.DataSyncerCreator = &scheduledDataSyncer.ScheduledSyncerFactoryStub{
		CreateCalled: func(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
			return nil, expectedError
		},
	}

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	notarizedShardHdr := &block.HeaderV2{
		Header:            nil,
		ScheduledRootHash: nil,
	}

	syncData, err := epochStartProvider.updateDataForScheduled(notarizedShardHdr)
	require.Nil(t, syncData)
	require.Equal(t, expectedError, err)
}

func TestEpochStartBootstrap_updateDataForScheduled(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.DestinationShardAsObserver = common.DisabledShardIDAsObserver
	args.GenesisNodesConfig = getNodesConfigMock(2)
	expectedSyncData := &dataToSync{
		ownShardHdr: &block.HeaderV2{
			ScheduledRootHash: []byte("rootHash1"),
		},
		rootHashToSync:    []byte("rootHash2"),
		withScheduled:     false,
		additionalHeaders: map[string]data.HeaderHandler{"key1": &block.HeaderV2{}},
	}

	args.DataSyncerCreator = &scheduledDataSyncer.ScheduledSyncerFactoryStub{
		CreateCalled: func(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
			return &scheduledDataSyncer.ScheduledSyncerStub{
				UpdateSyncDataIfNeededCalled: func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error) {
					return expectedSyncData.ownShardHdr, expectedSyncData.additionalHeaders, nil
				},
				GetRootHashToSyncCalled: func(notarizedShardHeader data.ShardHeaderHandler) []byte {
					return expectedSyncData.rootHashToSync
				},
			}, nil
		},
	}

	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)

	notarizedShardHdr := &block.HeaderV2{
		Header:            nil,
		ScheduledRootHash: nil,
	}

	syncData, err := epochStartProvider.updateDataForScheduled(notarizedShardHdr)
	require.Nil(t, err)
	require.Equal(t, expectedSyncData, syncData)
}

func TestEpochStartBootstrap_getDataToSyncErrorOpeningDB(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.DestinationShardAsObserver = common.DisabledShardIDAsObserver
	args.GenesisNodesConfig = getNodesConfigMock(2)

	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)

	expectedErr := fmt.Errorf("expected error")
	epochStartProvider.storageOpenerHandler = &storageMocks.UnitOpenerStub{
		OpenDBCalled: func(dbConfig config.DBConfig, shardID uint32, epoch uint32) (storage.Storer, error) {
			return nil, expectedErr
		},
	}

	shardNotarizedHeader := &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("scheduled root hash"),
	}
	epochStartData := &epochStartMocks.EpochStartShardDataStub{}

	syncData, err := epochStartProvider.getDataToSync(epochStartData, shardNotarizedHeader)
	require.Nil(t, syncData)
	require.Equal(t, expectedErr, err)
}

func TestEpochStartBootstrap_getDataToSyncErrorUpdatingDataForScheduled(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.DestinationShardAsObserver = common.DisabledShardIDAsObserver
	args.GenesisNodesConfig = getNodesConfigMock(2)

	expectedErr := fmt.Errorf("expected error")

	// Simulate an error in getDataToSync through the factory
	args.DataSyncerCreator = &scheduledDataSyncer.ScheduledSyncerFactoryStub{
		CreateCalled: func(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
			return nil, expectedErr
		},
	}

	shardNotarizedHeader := &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("scheduled root hash"),
	}
	epochStartData := &epochStartMocks.EpochStartShardDataStub{}

	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)

	syncData, err := epochStartProvider.getDataToSync(epochStartData, shardNotarizedHeader)
	require.Nil(t, syncData)
	require.Equal(t, expectedErr, err)
}

func TestEpochStartBootstrap_getDataToSyncWithSCRStorageCloseErr(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.DestinationShardAsObserver = common.DisabledShardIDAsObserver
	args.GenesisNodesConfig = getNodesConfigMock(2)

	shardNotarizedHeader := &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("scheduled root hash"),
	}

	expectedSyncData := &dataToSync{
		ownShardHdr:       shardNotarizedHeader,
		rootHashToSync:    []byte("rootHash2"),
		withScheduled:     false,
		additionalHeaders: map[string]data.HeaderHandler{"key1": &block.HeaderV2{}},
	}

	args.DataSyncerCreator = &scheduledDataSyncer.ScheduledSyncerFactoryStub{
		CreateCalled: func(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
			return &scheduledDataSyncer.ScheduledSyncerStub{
				UpdateSyncDataIfNeededCalled: func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error) {
					return expectedSyncData.ownShardHdr, expectedSyncData.additionalHeaders, nil
				},
				GetRootHashToSyncCalled: func(notarizedShardHeader data.ShardHeaderHandler) []byte {
					return expectedSyncData.rootHashToSync
				},
			}, nil
		},
	}
	epochStartData := &epochStartMocks.EpochStartShardDataStub{}

	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)

	expectedErr := fmt.Errorf("expected error")
	epochStartProvider.storerScheduledSCRs = &storageMocks.StorerStub{
		CloseCalled: func() error {
			return expectedErr
		},
	}

	syncData, err := epochStartProvider.getDataToSync(epochStartData, shardNotarizedHeader)
	require.Nil(t, err)
	require.Equal(t, expectedSyncData, syncData)
}
