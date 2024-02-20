package bootstrap

import (
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/versioning"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/types"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/scheduledDataSyncer"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageMocks "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestEpochStartBootstrapperFactory_NewEpochStartBootstrapperFactory(t *testing.T) {
	t.Parallel()

	esbf := NewEpochStartBootstrapperFactory()

	require.NotNil(t, esbf)
}

func TestEpochStartBootstrapperFactory_CreateEpochStartBootstrapper(t *testing.T) {
	t.Parallel()

	esbf := NewEpochStartBootstrapperFactory()
	esb, err := esbf.CreateEpochStartBootstrapper(getDefaultArgs())

	require.Nil(t, err)
	require.NotNil(t, esb)
}

func TestEpochStartBootstrapperFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	esbf := NewEpochStartBootstrapperFactory()

	require.False(t, esbf.IsInterfaceNil())

	esbf = (*epochStartBootstrapperFactory)(nil)
	require.True(t, esbf.IsInterfaceNil())
}

func getDefaultArgs() ArgsEpochStartBootstrap {
	generalCfg := testscommon.GetGeneralConfig()

	coreMock := &mock.CoreComponentsMock{
		IntMarsh:                     &mock.MarshalizerMock{},
		Marsh:                        &mock.MarshalizerMock{},
		Hash:                         &hashingMocks.HasherMock{},
		TxSignHasherField:            &hashingMocks.HasherMock{},
		UInt64ByteSliceConv:          &mock.Uint64ByteSliceConverterMock{},
		AddrPubKeyConv:               &testscommon.PubkeyConverterMock{},
		PathHdl:                      &testscommon.PathManagerStub{},
		EpochNotifierField:           &epochNotifier.EpochNotifierStub{},
		TxVersionCheckField:          versioning.NewTxVersionChecker(1),
		NodeTypeProviderField:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		ProcessStatusHandlerInstance: &testscommon.ProcessStatusHandlerStub{},
		HardforkTriggerPubKeyField:   []byte("provided hardfork pub key"),
		EnableEpochsHandlerField:     &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	cryptoMock := &mock.CryptoComponentsMock{
		PubKey:          &cryptoMocks.PublicKeyStub{},
		PrivKey:         &cryptoMocks.PrivateKeyStub{},
		BlockSig:        &cryptoMocks.SignerStub{},
		TxSig:           &cryptoMocks.SignerStub{},
		BlKeyGen:        &cryptoMocks.KeyGenStub{},
		TxKeyGen:        &cryptoMocks.KeyGenStub{},
		PeerSignHandler: &cryptoMocks.PeerSignatureHandlerStub{},
		ManagedPeers:    &testscommon.ManagedPeersHolderStub{},
	}

	messenger := &p2pmocks.MessengerStub{
		ConnectedPeersCalled: func() []core.PeerID {
			return []core.PeerID{"peer0", "peer1", "peer2", "peer3", "peer4", "peer5"}
		},
	}

	return ArgsEpochStartBootstrap{
		ScheduledSCRsStorer:    genericMocks.NewStorerMock(),
		CoreComponentsHolder:   coreMock,
		CryptoComponentsHolder: cryptoMock,
		MainMessenger:          messenger,
		FullArchiveMessenger:   messenger,
		GeneralConfig: config.Config{
			MiniBlocksStorage:               generalCfg.MiniBlocksStorage,
			PeerBlockBodyStorage:            generalCfg.PeerBlockBodyStorage,
			BlockHeaderStorage:              generalCfg.BlockHeaderStorage,
			TxStorage:                       generalCfg.TxStorage,
			UnsignedTransactionStorage:      generalCfg.UnsignedTransactionStorage,
			RewardTxStorage:                 generalCfg.RewardTxStorage,
			ShardHdrNonceHashStorage:        generalCfg.ShardHdrNonceHashStorage,
			MetaHdrNonceHashStorage:         generalCfg.MetaHdrNonceHashStorage,
			StatusMetricsStorage:            generalCfg.StatusMetricsStorage,
			ReceiptsStorage:                 generalCfg.ReceiptsStorage,
			SmartContractsStorage:           generalCfg.SmartContractsStorage,
			SmartContractsStorageForSCQuery: generalCfg.SmartContractsStorageForSCQuery,
			TrieEpochRootHashStorage:        generalCfg.TrieEpochRootHashStorage,
			BootstrapStorage:                generalCfg.BootstrapStorage,
			MetaBlockStorage:                generalCfg.MetaBlockStorage,
			AccountsTrieStorage:             generalCfg.AccountsTrieStorage,
			PeerAccountsTrieStorage:         generalCfg.PeerAccountsTrieStorage,
			HeartbeatV2:                     generalCfg.HeartbeatV2,
			Hardfork:                        generalCfg.Hardfork,
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
				AccountsStatePruningEnabled: true,
				PeerStatePruningEnabled:     true,
				MaxStateTrieLevelInMemory:   5,
				MaxPeerTrieLevelInMemory:    5,
			},
			TrieStorageManagerConfig: config.TrieStorageManagerConfig{
				PruningBufferLen:      1000,
				SnapshotsBufferLen:    10,
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
				CheckNodesOnDisk:          false,
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
			Requesters: generalCfg.Requesters,
		},
		EconomicsData: &economicsmocks.EconomicsHandlerStub{
			MinGasPriceCalled: func() uint64 {
				return 1
			},
		},
		GenesisNodesConfig:         &testscommon.NodesSetupStub{},
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
					UpdateSyncDataIfNeededCalled: func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, map[string]*block.MiniBlock, error) {
						return notarizedShardHeader, nil, nil, nil
					},
					GetRootHashToSyncCalled: func(notarizedShardHeader data.ShardHeaderHandler) []byte {
						return notarizedShardHeader.GetRootHash()
					},
				}, nil
			},
		},
		FlagsConfig: config.ContextFlagsConfig{
			ForceStartFromNetwork: false,
		},
		TrieSyncStatisticsProvider:       &testscommon.SizeSyncStatisticsHandlerStub{},
		AdditionalStorageServiceCreator:  &testscommon.AdditionalStorageServiceFactoryMock{},
		NodesCoordinatorWithRaterFactory: nodesCoordinator.NewIndexHashedNodesCoordinatorWithRaterFactory(),
		ShardCoordinatorFactory:          sharding.NewMultiShardCoordinatorFactory(),
		StateStatsHandler:                disabled.NewStateStatistics(),
	}
}
