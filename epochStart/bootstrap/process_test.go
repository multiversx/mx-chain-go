package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	disabledStatistics "github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/types"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	epochStartMocks "github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks/epochStart"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/scheduledDataSyncer"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageMocks "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/syncer"
	validatorInfoCacherStub "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	"github.com/multiversx/mx-chain-go/trie/factory"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/versioning"
	"github.com/multiversx/mx-chain-core-go/data"
	dataBatch "github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
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
			EnableEpochsHandlerField: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
					if flag == common.StakingV4Step2Flag {
						return 99999
					}
					return 0
				},
			},
		},
		&mock.CryptoComponentsMock{
			PubKey:          &cryptoMocks.PublicKeyStub{},
			PrivKey:         &cryptoMocks.PrivateKeyStub{},
			BlockSig:        &cryptoMocks.SignerStub{},
			TxSig:           &cryptoMocks.SignerStub{},
			BlKeyGen:        &cryptoMocks.KeyGenStub{},
			TxKeyGen:        &cryptoMocks.KeyGenStub{},
			PeerSignHandler: &cryptoMocks.PeerSignatureHandlerStub{},
			ManagedPeers:    &testscommon.ManagedPeersHolderStub{},
		}
}

func createMockEpochStartBootstrapArgs(
	coreMock *mock.CoreComponentsMock,
	cryptoMock *mock.CryptoComponentsMock,
) ArgsEpochStartBootstrap {
	generalCfg := testscommon.GetGeneralConfig()
	return ArgsEpochStartBootstrap{
		ScheduledSCRsStorer:    genericMocks.NewStorerMock(),
		CoreComponentsHolder:   coreMock,
		CryptoComponentsHolder: cryptoMock,
		MainMessenger: &p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer0", "peer1", "peer2", "peer3", "peer4", "peer5"}
			}},
		NodesCoordinatorRegistryFactory: &shardingMocks.NodesCoordinatorRegistryFactoryMock{},
		FullArchiveMessenger:            &p2pmocks.MessengerStub{},
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
				SnapshotsEnabled:            true,
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
			Requesters:      generalCfg.Requesters,
			SovereignConfig: generalCfg.SovereignConfig,
		},
		EconomicsData: &economicsmocks.EconomicsHandlerStub{
			MinGasPriceCalled: func() uint64 {
				return 1
			},
		},
		GenesisNodesConfig:         &genesisMocks.NodesSetupStub{},
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
		TrieSyncStatisticsProvider: &testscommon.SizeSyncStatisticsHandlerStub{},
		StateStatsHandler:          disabledStatistics.NewStateStatistics(),
		RunTypeComponents:          mock.NewRunTypeComponentsStub(),
	}
}

func TestNewEpochStartBootstrap_NilArgsChecks(t *testing.T) {
	t.Parallel()

	t.Run("nil shardCoordinator", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GenesisShardCoordinator = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilShardCoordinator))
	})
	t.Run("nil main messenger", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.MainMessenger = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilMessenger))
	})
	t.Run("nil full archive messenger", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.FullArchiveMessenger = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilMessenger))
	})
	t.Run("nil economicsData", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.EconomicsData = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilEconomicsData))
	})
	t.Run("nil coreComponentsHolder", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.CoreComponentsHolder = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilCoreComponentsHolder))
	})
	t.Run("nil cryptoComponentsHolder", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.CryptoComponentsHolder = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilCryptoComponentsHolder))
	})
	t.Run("nil pubKey", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		cryptoComp.PubKey = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilPubKey))
	})
	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.Hash = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilHasher))
	})
	t.Run("nil marshalizer", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.IntMarsh = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilMarshalizer))
	})
	t.Run("nil blockKeyGen", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		cryptoComp.BlKeyGen = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilBlockKeyGen))
	})
	t.Run("nil keyGen", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		cryptoComp.TxKeyGen = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilKeyGen))
	})
	t.Run("nil singleSigner", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		cryptoComp.TxSig = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilSingleSigner))
	})
	t.Run("nil blockSingleSigner", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		cryptoComp.BlockSig = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilBlockSingleSigner))
	})
	t.Run("nil txSignMarshalizer", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.Marsh = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilTxSignMarshalizer))
	})
	t.Run("nil pathManager", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.PathHdl = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilPathManager))
	})
	t.Run("nil genesisNodesConfig", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GenesisNodesConfig = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilGenesisNodesConfig))
	})
	t.Run("nil rater", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.Rater = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilRater))
	})
	t.Run("nil pubkeyConverter", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.AddrPubKeyConv = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilPubkeyConverter))
	})
	t.Run("nil trieSyncStatistics", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.TrieSyncStatisticsProvider = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilTrieSyncStatistics))
	})
	t.Run("nil roundHandler", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.RoundHandler = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilRoundHandler))
	})
	t.Run("nil storageUnitOpener", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.StorageUnitOpener = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilStorageUnitOpener))
	})
	t.Run("nil latestStorageDataProvider", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.LatestStorageDataProvider = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilLatestStorageDataProvider))
	})
	t.Run("nil uint64Converter", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.UInt64ByteSliceConv = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilUint64Converter))
	})
	t.Run("nil shuffler", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.NodeShuffler = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilShuffler))
	})
	t.Run("not enough num of peers to consider block valid from config", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GeneralConfig.EpochStartConfig.MinNumOfPeersToConsiderBlockValid = minNumPeersToConsiderMetaBlockValid - 1

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNotEnoughNumOfPeersToConsiderBlockValid))
	})
	t.Run("not enough num connected peers", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GeneralConfig.EpochStartConfig.MinNumConnectedPeersToStart = minNumConnectedPeers - 1

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNotEnoughNumConnectedPeers))
	})
	t.Run("nil argumentsParser", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.ArgumentsParser = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilArgumentsParser))
	})
	t.Run("nil statusHandler", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.StatusHandler = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilStatusHandler))
	})
	t.Run("nil headerIntegrityVerifier", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.HeaderIntegrityVerifier = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilHeaderIntegrityVerifier))
	})
	t.Run("nil scheduledDataSyncerFactory", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.DataSyncerCreator = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilScheduledDataSyncerFactory))
	})
	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.TxSignHasherField = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilHasher))
	})
	t.Run("nil epochNotifier", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.EpochNotifierField = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilEpochNotifier))
	})
	t.Run("invalid max hardcap for missing nodes", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GeneralConfig.TrieSync.MaxHardCapForMissingNodes = 0

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrInvalidMaxHardCapForMissingNodes))
	})
	t.Run("invalid num concurrent trie syncers", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GeneralConfig.TrieSync.NumConcurrentTrieSyncers = 0

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.True(t, errors.Is(err, epochStart.ErrInvalidNumConcurrentTrieSyncers))
		require.Nil(t, epochStartProvider)
	})
	t.Run("fail to create whiteList cache", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GeneralConfig.WhiteListPool = config.CacheConfig{}

		epochStartProvider, err := NewEpochStartBootstrap(args)
		assert.Equal(t, storage.ErrNotSupportedCacheType, err)
		assert.Nil(t, epochStartProvider)
	})
	t.Run("nil managed peers holder", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		cryptoComp.ManagedPeers = nil
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, epochStart.ErrNilManagedPeersHolder))
	})
	t.Run("nil state statistics handler", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.StateStatsHandler = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, statistics.ErrNilStateStatsHandler))
	})
	t.Run("nil RunTypeComponents should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.RunTypeComponents = nil

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilRunTypeComponents))
		require.Nil(t, epochStartProvider)
	})
	t.Run("nil ShardCoordinatorFactory should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.RunTypeComponents = &mock.RunTypeComponentsStub{
			ShardCoordinatorFactory: nil,
		}

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, errorsMx.ErrNilShardCoordinatorFactory))
	})
	t.Run("nil AdditionalStorageServiceCreator should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.RunTypeComponents = &mock.RunTypeComponentsStub{
			ShardCoordinatorFactory:         &testscommon.MultiShardCoordinatorFactoryMock{},
			AdditionalStorageServiceFactory: nil,
		}

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, errorsMx.ErrNilAdditionalStorageServiceCreator))
	})
	t.Run("nil NodesCoordinatorWithRaterFactory should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.RunTypeComponents = &mock.RunTypeComponentsStub{
			ShardCoordinatorFactory:          &testscommon.MultiShardCoordinatorFactoryMock{},
			AdditionalStorageServiceFactory:  &testscommon.AdditionalStorageServiceFactoryMock{},
			NodesCoordinatorWithRaterFactory: nil,
		}

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, errorsMx.ErrNilNodesCoordinatorFactory))
	})
	t.Run("nil NodesCoordinatorWithRaterFactory should error", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.RunTypeComponents = &mock.RunTypeComponentsStub{
			ShardCoordinatorFactory:          &testscommon.MultiShardCoordinatorFactoryMock{},
			AdditionalStorageServiceFactory:  &testscommon.AdditionalStorageServiceFactoryMock{},
			NodesCoordinatorWithRaterFactory: &testscommon.NodesCoordinatorFactoryMock{},
			RequestHandlerFactory:            nil,
		}

		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, epochStartProvider)
		require.True(t, errors.Is(err, errorsMx.ErrNilRequestHandlerCreator))
	})
}

func TestNewEpochStartBootstrap(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()

	t.Run("hardfork disabled", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		epochStartProvider, err := NewEpochStartBootstrap(args)
		assert.Nil(t, err)
		assert.NotNil(t, epochStartProvider)
	})

	t.Run("hardfork enabled", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.GeneralConfig.Hardfork.AfterHardFork = true

		epochStartProvider, err := NewEpochStartBootstrap(args)
		assert.Nil(t, err)
		assert.NotNil(t, epochStartProvider)
	})
}

func TestEpochStartBootstrap_Boostrap(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()

	t.Run("failed to set shard coordinator, wrong number of shards", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.GeneralConfig.GeneralSettings.StartInEpochEnabled = true
		args.GenesisShardCoordinator = testscommon.NewMultiShardsCoordinatorMock(0)

		epochStartProvider, _ := NewEpochStartBootstrap(args)

		params, err := epochStartProvider.Bootstrap()
		assert.Equal(t, nodesCoordinator.ErrInvalidNumberOfShards, err)
		assert.Equal(t, Parameters{}, params)
	})
	t.Run("boostrap from local storage, fail to get boostrap data", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.GeneralConfig = testscommon.GetGeneralConfig()
		args.GeneralConfig.GeneralSettings.StartInEpochEnabled = false
		args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
			GetCalled: func() (storage.LatestDataFromStorage, error) {
				return storage.LatestDataFromStorage{
					Epoch:     2,
					ShardID:   0,
					LastRound: 10,
				}, nil
			},
		}

		epochStartProvider, _ := NewEpochStartBootstrap(args)

		expectedErr := errors.New("expected err")
		epochStartProvider.storageOpenerHandler = &storageMocks.UnitOpenerStub{
			GetMostRecentStorageUnitCalled: func(config config.DBConfig) (storage.Storer, error) {
				return &storageMocks.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						return nil, expectedErr
					},
				}, nil
			},
		}

		params, err := epochStartProvider.Bootstrap()
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, Parameters{}, params)
	})

	t.Run("bootstrap from local storage with StartInEpoch not enabled, should work", func(t *testing.T) {
		t.Parallel()

		testBoostrapByStartInEpochFlag(t, false)
	})

	t.Run("bootstrap from saved epoch, should work", func(t *testing.T) {
		t.Parallel()

		testBoostrapByStartInEpochFlag(t, true)
	})
}

func testBoostrapByStartInEpochFlag(t *testing.T, startInEpochEnabled bool) {
	coreComp, cryptoComp := createComponentsForEpochStart()

	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GeneralConfig = testscommon.GetGeneralConfig()
	args.GeneralConfig.GeneralSettings.StartInEpochEnabled = startInEpochEnabled

	epoch := uint32(1)
	shardId := uint32(0)
	args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetCalled: func() (storage.LatestDataFromStorage, error) {
			return storage.LatestDataFromStorage{
				Epoch:   epoch,
				ShardID: shardId,
			}, nil
		},
	}

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	pksBytes := createPkBytes(args.GenesisNodesConfig.NumberOfShards())

	nodesCoord := &nodesCoordinator.NodesCoordinatorRegistry{
		EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
			strconv.Itoa(int(epoch)): {
				EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{
					"0": {
						&nodesCoordinator.SerializableValidator{
							PubKey:  pksBytes[0],
							Chances: 1,
						},
					},
					"4294967295": {
						&nodesCoordinator.SerializableValidator{
							PubKey:  pksBytes[core.MetachainShardId],
							Chances: 1,
						},
					},
				},
				WaitingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
				LeavingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
			},
		},
	}
	nodesCoordBytes, _ := json.Marshal(nodesCoord)

	epochStartProvider.storageOpenerHandler = &storageMocks.UnitOpenerStub{
		GetMostRecentStorageUnitCalled: func(config config.DBConfig) (storage.Storer, error) {
			return &storageMocks.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nodesCoordBytes, nil
				},
				SearchFirstCalled: func(key []byte) ([]byte, error) {
					return nodesCoordBytes, nil
				},
			}, nil
		},
	}

	expectedParams := Parameters{
		Epoch:       epoch,
		SelfShardId: shardId,
		NumOfShards: uint32(len(nodesCoord.EpochsConfig[strconv.Itoa(int(epoch))].EligibleValidators)),
		NodesConfig: nodesCoord,
	}

	params, err := epochStartProvider.Bootstrap()
	assert.Nil(t, err)
	assert.Equal(t, expectedParams, params)
}

func TestIsStartInEpochZero(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GenesisNodesConfig = &genesisMocks.NodesSetupStub{
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
	args.GenesisNodesConfig = &genesisMocks.NodesSetupStub{
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
	args.GenesisNodesConfig = &genesisMocks.NodesSetupStub{
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
	args.GenesisNodesConfig = &genesisMocks.NodesSetupStub{
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
			return &testscommon.HeadersCacherStub{}
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
		PeerAuthenticationsCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
		HeartbeatsCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
	}
	epochStartProvider.whiteListHandler = &testscommon.WhiteListHandlerStub{}
	epochStartProvider.whiteListerVerifiedTxs = &testscommon.WhiteListHandlerStub{}
	epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
	epochStartProvider.storageService = &storageMocks.ChainStorerStub{}

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
		disabled.NewChainStorer(),
		disabledStatistics.NewStateStatistics(),
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
		disabled.NewChainStorer(),
		disabledStatistics.NewStateStatistics(),
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
		disabled.NewChainStorer(),
		disabledStatistics.NewStateStatistics(),
	)
	assert.Nil(t, err)
	epochStartProvider.trieContainer = triesContainer
	epochStartProvider.trieStorageManagers = trieStorageManagers

	rootHash := []byte("rootHash")
	err = epochStartProvider.syncUserAccountsState(rootHash)
	assert.Equal(t, state.ErrNilRequestHandler, err)
}

func TestRequestAndProcessForShard_ShouldFail(t *testing.T) {
	notarizedShardHeaderHash := []byte("notarizedShardHeaderHash")
	prevShardHeaderHash := []byte("prevShardHeaderHash")
	notarizedMetaHeaderHash := []byte("notarizedMetaHeaderHash")
	prevMetaHeaderHash := []byte("prevMetaHeaderHash")

	metaBlock := &block.MetaBlock{
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{HeaderHash: notarizedShardHeaderHash, ShardID: 0},
			},
		},
	}

	emptyMiniBlocksSlice := make([]*block.MiniBlock, 0)
	t.Run("find self shard epoch start data not found", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.epochStartMeta = &block.MetaBlock{}

		err := epochStartProvider.requestAndProcessForShard(emptyMiniBlocksSlice)
		assert.Equal(t, epochStart.ErrEpochStartDataForShardNotFound, err)
	})
	t.Run("fail to sync pending miniblocks", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.epochStartMeta = metaBlock

		expectedErr := errors.New("sync pending miniblocks error")
		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{
			SyncPendingMiniBlocksCalled: func(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error {
				return expectedErr
			},
		}

		err := epochStartProvider.requestAndProcessForShard(emptyMiniBlocksSlice)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("fail to get pending miniblocks", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.epochStartMeta = metaBlock

		expectedErr := errors.New("get pending miniblocks error")
		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{
			GetMiniBlocksCalled: func() (map[string]*block.MiniBlock, error) {
				return nil, expectedErr
			},
		}

		err := epochStartProvider.requestAndProcessForShard(emptyMiniBlocksSlice)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("fail to sync missing headers", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.epochStartMeta = metaBlock

		expectedErr := errors.New("sync miniBlocksSyncer headers by hash error")
		epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			SyncMissingHeadersByHashCalled: func(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error {
				return expectedErr
			},
		}

		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}

		err := epochStartProvider.requestAndProcessForShard(emptyMiniBlocksSlice)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("fail to get needed headers", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.epochStartMeta = metaBlock

		expectedErr := errors.New("get pending miniblocks error")
		epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
				return nil, expectedErr
			},
		}

		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}

		err := epochStartProvider.requestAndProcessForShard(emptyMiniBlocksSlice)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("fail to get data to sync", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		prevShardHeader := &block.Header{}
		notarizedShardHeader := &block.Header{
			PrevHash: prevShardHeaderHash,
		}

		expectedErr := fmt.Errorf("expected error")
		args.DataSyncerCreator = &scheduledDataSyncer.ScheduledSyncerFactoryStub{
			CreateCalled: func(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
				return &scheduledDataSyncer.ScheduledSyncerStub{
					UpdateSyncDataIfNeededCalled: func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, map[string]*block.MiniBlock, error) {
						return nil, nil, nil, expectedErr
					},
				}, nil
			},
		}

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.syncedHeaders = make(map[string]data.HeaderHandler)
		epochStartProvider.epochStartMeta = metaBlock
		epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
				return map[string]data.HeaderHandler{
					string(notarizedShardHeaderHash): notarizedShardHeader,
					string(prevShardHeaderHash):      prevShardHeader,
				}, nil
			},
		}

		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}

		err := epochStartProvider.requestAndProcessForShard(emptyMiniBlocksSlice)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("fail to create user accounts syncer", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		prevShardHeader := &block.Header{}
		notarizedShardHeader := &block.Header{
			PrevHash: prevShardHeaderHash,
		}

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.syncedHeaders = make(map[string]data.HeaderHandler)
		epochStartProvider.epochStartMeta = metaBlock
		epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
				return map[string]data.HeaderHandler{
					string(notarizedShardHeaderHash): notarizedShardHeader,
					string(prevShardHeaderHash):      prevShardHeader,
				}, nil
			},
		}
		epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
		epochStartProvider.dataPool = &dataRetrieverMock.PoolsHolderStub{
			TrieNodesCalled: func() storage.Cacher {
				return nil
			},
		}

		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}

		err := epochStartProvider.requestAndProcessForShard(emptyMiniBlocksSlice)
		assert.Equal(t, state.ErrNilCacher, err)
	})
	t.Run("fail to save data to storage", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.IntMarsh = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}

		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		prevShardHeader := &block.Header{}
		notarizedShardHeader := &block.Header{
			PrevHash: prevShardHeaderHash,
		}
		notarizedMetaHeader := &block.MetaBlock{
			PrevHash: prevMetaHeaderHash,
		}
		metaBlockInstance := &block.MetaBlock{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{
						HeaderHash:            notarizedShardHeaderHash,
						ShardID:               0,
						FirstPendingMetaBlock: notarizedMetaHeaderHash,
					},
				},
			},
		}
		prevMetaBlock := &block.MetaBlock{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{
						HeaderHash: notarizedShardHeaderHash,
						ShardID:    0,
					},
				},
			},
		}

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.syncedHeaders = make(map[string]data.HeaderHandler)
		epochStartProvider.epochStartMeta = metaBlockInstance
		epochStartProvider.prevEpochStartMeta = prevMetaBlock
		epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
				return map[string]data.HeaderHandler{
					string(notarizedShardHeaderHash): notarizedShardHeader,
					string(notarizedMetaHeaderHash):  notarizedMetaHeader,
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

		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}
		epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
		epochStartProvider.nodesConfig = &nodesCoordinator.NodesCoordinatorRegistry{}

		err := epochStartProvider.requestAndProcessForShard(emptyMiniBlocksSlice)
		assert.Equal(t, expectedErr, err)
	})
}

func TestRequestAndProcessForMeta_ShouldFail(t *testing.T) {
	notarizedShardHeaderHash := []byte("notarizedShardHeaderHash")
	prevShardHeaderHash := []byte("prevShardHeaderHash")

	emptyMiniBlocksSlice := make([]*block.MiniBlock, 0)
	t.Run("fail to create storage handler component", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		metaBlock := &block.MetaBlock{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: notarizedShardHeaderHash, ShardID: 0},
				},
			},
		}

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.epochStartMeta = metaBlock

		epochStartProvider.shardCoordinator = nil

		err := epochStartProvider.requestAndProcessForMeta(emptyMiniBlocksSlice)
		assert.Equal(t, storage.ErrNilShardCoordinator, err)
	})
	t.Run("fail to create validators accounts syncer", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		prevShardHeader := &block.Header{}
		notarizedShardHeader := &block.Header{
			PrevHash: prevShardHeaderHash,
		}
		metaBlock := &block.MetaBlock{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: notarizedShardHeaderHash, ShardID: 0},
				},
			},
		}

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.epochStartMeta = metaBlock
		epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
				return map[string]data.HeaderHandler{
					string(notarizedShardHeaderHash): notarizedShardHeader,
					string(prevShardHeaderHash):      prevShardHeader,
				}, nil
			},
		}
		epochStartProvider.dataPool = dataRetrieverMock.NewPoolsHolderMock()

		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}

		err := epochStartProvider.requestAndProcessForMeta(emptyMiniBlocksSlice)
		assert.Equal(t, state.ErrNilRequestHandler, err)
	})
	t.Run("fail to sync user accounts state", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.IntMarsh = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		prevShardHeader := &block.Header{}
		notarizedShardHeader := &block.Header{
			PrevHash: prevShardHeaderHash,
		}
		metaBlock := &block.MetaBlock{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: notarizedShardHeaderHash, ShardID: 0},
				},
			},
		}

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.syncedHeaders = make(map[string]data.HeaderHandler)
		epochStartProvider.epochStartMeta = metaBlock
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

		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}
		epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}

		err := epochStartProvider.requestAndProcessForMeta(emptyMiniBlocksSlice)
		assert.Equal(t, expectedErr, err)
	})
}

func TestPrepareComponentsToSyncFromNetwork(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 0

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.syncedHeaders = make(map[string]data.HeaderHandler)
	epochStartProvider.dataPool = dataRetrieverMock.NewPoolsHolderMock()

	epochStartProvider.shardCoordinator = shardCoordinator
	epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}
	epochStartProvider.nodesConfig = &nodesCoordinator.NodesCoordinatorRegistry{}

	assert.Nil(t, epochStartProvider.requestHandler)
	assert.Nil(t, epochStartProvider.epochStartMetaBlockSyncer)

	err := epochStartProvider.prepareComponentsToSyncFromNetwork()
	assert.Nil(t, err)

	assert.NotNil(t, epochStartProvider.requestHandler)
	assert.NotNil(t, epochStartProvider.epochStartMetaBlockSyncer)
}

func getNodesConfigMock(numOfShards uint32) sharding.GenesisNodesSetupHandler {
	pksBytes := createPkBytes(numOfShards)
	address := []byte("afafafafafafafafafafafafafafafaf")

	roundDurationMillis := 4000
	epochDurationMillis := 50 * int64(roundDurationMillis)

	nodesConfig := &genesisMocks.NodesSetupStub{
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
	prevPrevEpochStartMetaHeaderHash := []byte("prevPrevEpochStartMetaHeaderHash")
	prevEpochStartMetaHeaderHash := []byte("prevEpochStartMetaHeaderHash")
	prevEpochNotarizedShardHeaderHash := []byte("prevEpochNotarizedShardHeaderHash")
	notarizedShardHeaderHash := []byte("notarizedShardHeaderHash")
	epochStartMetaBlockHash := []byte("epochStartMetaBlockHash")
	prevNotarizedShardHeaderHash := []byte("prevNotarizedShardHeaderHash")

	t.Run("fail to sync headers from epoch start", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		shardId := uint32(0)
		epochStartMetaBlock := &block.MetaBlock{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: notarizedShardHeaderHash, ShardID: shardId},
				},
				Economics: block.Economics{
					PrevEpochStartHash: prevEpochStartMetaHeaderHash,
				},
			},
		}

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.epochStartMeta = epochStartMetaBlock
		expectedErr := errors.New("sync miniBlocksSyncer headers by hash error")
		epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			SyncMissingHeadersByHashCalled: func(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error {
				assert.Equal(t, [][]byte{notarizedShardHeaderHash}, headersHashes)
				assert.Equal(t, []uint32{shardId}, shardIDs)
				return expectedErr
			},
		}

		params, err := epochStartProvider.requestAndProcessing()
		assert.Equal(t, Parameters{}, params)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("fail with wrong type assertion on epoch start meta", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())

		notarizedShardHeader := &block.Header{
			PrevHash: prevNotarizedShardHeaderHash,
		}
		prevNotarizedShardHeader := &block.Header{}

		epochStartMetaBlock := &block.MetaBlock{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: notarizedShardHeaderHash, ShardID: 0},
				},
				Economics: block.Economics{
					PrevEpochStartHash: prevEpochStartMetaHeaderHash,
				},
			},
		}

		epochStartProvider, _ := NewEpochStartBootstrap(args)
		epochStartProvider.epochStartMeta = epochStartMetaBlock
		epochStartProvider.headersSyncer = &epochStartMocks.HeadersByHashSyncerStub{
			GetHeadersCalled: func() (m map[string]data.HeaderHandler, err error) {
				return map[string]data.HeaderHandler{
					string(notarizedShardHeaderHash):     notarizedShardHeader,
					string(epochStartMetaBlockHash):      epochStartMetaBlock,
					string(prevNotarizedShardHeaderHash): prevNotarizedShardHeader,
				}, nil
			},
		}

		params, err := epochStartProvider.requestAndProcessing()
		assert.Equal(t, Parameters{}, params)
		assert.Equal(t, epochStart.ErrWrongTypeAssertion, err)
	})
	t.Run("fail to get public key bytes", func(t *testing.T) {
		t.Parallel()

		coreComp, cryptoComp := createComponentsForEpochStart()
		expectedErr := errors.New("expected err")
		cryptoComp.PubKey = &cryptoMocks.PublicKeyStub{
			ToByteArrayStub: func() ([]byte, error) {
				return nil, expectedErr
			},
		}

		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

		notarizedShardHeader := &block.Header{
			PrevHash: prevNotarizedShardHeaderHash,
		}
		prevNotarizedShardHeader := &block.Header{}

		epochStartMetaBlock := &block.MetaBlock{
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

		params, err := epochStartProvider.requestAndProcessing()
		assert.Equal(t, Parameters{}, params)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("failed to set shard coordinator, wrong number of shards", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GenesisNodesConfig = getNodesConfigMock(1)

		notarizedShardHeader := &block.Header{
			PrevHash: prevNotarizedShardHeaderHash,
		}
		prevNotarizedShardHeader := &block.Header{}

		epochStartMetaBlock := &block.MetaBlock{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{},
				Economics: block.Economics{
					PrevEpochStartHash: prevEpochStartMetaHeaderHash,
				},
			},
		}
		prevEpochStartMetaBlock := &block.MetaBlock{
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
		epochStartProvider.dataPool = dataRetrieverMock.NewPoolsHolderMock()
		epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
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

		params, err := epochStartProvider.requestAndProcessing()
		assert.Equal(t, Parameters{}, params)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), nodesCoordinator.ErrInvalidNumberOfShards.Error()))
	})
	t.Run("failed to create main messenger topic", func(t *testing.T) {
		t.Parallel()

		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GenesisNodesConfig = getNodesConfigMock(1)

		expectedErr := errors.New("expected error")
		args.MainMessenger = &p2pmocks.MessengerStub{
			CreateTopicCalled: func(topic string, identifier bool) error {
				return expectedErr
			},
		}

		notarizedShardHeader := &block.Header{
			PrevHash: prevNotarizedShardHeaderHash,
		}
		prevNotarizedShardHeader := &block.Header{}

		epochStartMetaBlock := &block.MetaBlock{
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
		epochStartProvider.dataPool = dataRetrieverMock.NewPoolsHolderMock()
		epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
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

		params, err := epochStartProvider.requestAndProcessing()
		assert.Equal(t, Parameters{}, params)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("request and process for shard fail, invalid num active persisters", func(t *testing.T) {
		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GeneralConfig.StoragePruning.NumActivePersisters = 0
		args.GenesisNodesConfig = getNodesConfigMock(1)

		notarizedShardHeader := &block.Header{
			PrevHash: prevNotarizedShardHeaderHash,
		}
		prevNotarizedShardHeader := &block.Header{}

		epochStartMetaBlock := &block.MetaBlock{
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
				return &testscommon.HeadersCacherStub{}
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &validatorInfoCacherStub.ValidatorInfoCacherStub{}
			},
		}
		epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}
		epochStartProvider.txSyncerForScheduled = &syncer.TransactionsSyncHandlerMock{}

		params, err := epochStartProvider.requestAndProcessing()
		assert.Equal(t, Parameters{}, params)
		assert.Equal(t, storage.ErrInvalidNumberOfActivePersisters, err)
	})
	t.Run("request and process for meta fail, invalid num active persisters", func(t *testing.T) {
		args := createMockEpochStartBootstrapArgs(createComponentsForEpochStart())
		args.GeneralConfig.StoragePruning.NumActivePersisters = 0
		args.GenesisNodesConfig = getNodesConfigMock(1)
		args.DestinationShardAsObserver = core.MetachainShardId

		notarizedShardHeader := &block.Header{
			PrevHash: prevNotarizedShardHeaderHash,
		}
		prevNotarizedShardHeader := &block.Header{}

		epochStartMetaBlock := &block.MetaBlock{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: notarizedShardHeaderHash, ShardID: core.MetachainShardId},
				},
				Economics: block.Economics{
					PrevEpochStartHash: prevEpochStartMetaHeaderHash,
				},
			},
		}
		prevEpochStartMetaBlock := &block.MetaBlock{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{HeaderHash: prevEpochNotarizedShardHeaderHash, ShardID: core.MetachainShardId},
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
				return &testscommon.HeadersCacherStub{}
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &validatorInfoCacherStub.ValidatorInfoCacherStub{}
			},
		}
		epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
		epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}
		epochStartProvider.txSyncerForScheduled = &syncer.TransactionsSyncHandlerMock{}

		params, err := epochStartProvider.requestAndProcessing()
		assert.Equal(t, Parameters{}, params)
		assert.Equal(t, storage.ErrInvalidNumberOfActivePersisters, err)
	})

	t.Run("should work for shard", func(t *testing.T) {
		t.Parallel()

		testRequestAndProcessingByShardId(t, uint32(0))
	})

	t.Run("should work for meta", func(t *testing.T) {
		t.Parallel()

		testRequestAndProcessingByShardId(t, core.MetachainShardId)
	})
}

func testRequestAndProcessingByShardId(t *testing.T, shardId uint32) {
	coreComp, cryptoComp := createComponentsForEpochStart()

	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GenesisNodesConfig = getNodesConfigMock(1)
	args.DestinationShardAsObserver = shardId

	prevPrevEpochStartMetaHeaderHash := []byte("prevPrevEpochStartMetaHeaderHash")
	prevEpochStartMetaHeaderHash := []byte("prevEpochStartMetaHeaderHash")
	notarizedShardHeaderHash := []byte("notarizedShardHeaderHash")
	epochStartMetaBlockHash := []byte("epochStartMetaBlockHash")
	prevNotarizedShardHeaderHash := []byte("prevNotarizedShardHeaderHash")
	notarizedShardHeader := &block.Header{
		PrevHash: prevNotarizedShardHeaderHash,
	}
	prevNotarizedShardHeader := &block.Header{}
	notarizedMetaHeaderHash := []byte("notarizedMetaHeaderHash")
	prevMetaHeaderHash := []byte("prevMetaHeaderHash")
	notarizedMetaHeader := &block.MetaBlock{
		PrevHash: prevMetaHeaderHash,
	}

	epochStartMetaBlock := &block.MetaBlock{
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					HeaderHash:            notarizedShardHeaderHash,
					ShardID:               shardId,
					FirstPendingMetaBlock: notarizedMetaHeaderHash,
				},
			},
			Economics: block.Economics{
				PrevEpochStartHash: prevEpochStartMetaHeaderHash,
			},
		},
	}
	prevEpochStartMetaBlock := &block.MetaBlock{
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					HeaderHash: notarizedShardHeaderHash,
					ShardID:    shardId,
				},
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
				string(notarizedMetaHeaderHash):      notarizedMetaHeader,
			}, nil
		},
	}
	epochStartProvider.dataPool = dataRetrieverMock.NewPoolsHolderMock()
	epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
	epochStartProvider.miniBlocksSyncer = &epochStartMocks.PendingMiniBlockSyncHandlerStub{}

	pksBytes := createPkBytes(args.GenesisNodesConfig.NumberOfShards())

	requiredParameters := Parameters{
		SelfShardId: shardId,
		NumOfShards: args.GenesisNodesConfig.NumberOfShards(),
		NodesConfig: &nodesCoordinator.NodesCoordinatorRegistry{
			EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
				"0": {
					EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{
						"0": {
							&nodesCoordinator.SerializableValidator{
								PubKey:  pksBytes[0],
								Chances: 1,
							},
						},
						"4294967295": {
							&nodesCoordinator.SerializableValidator{
								PubKey:  pksBytes[core.MetachainShardId],
								Chances: 1,
							},
						},
					},
					WaitingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
					LeavingValidators: map[string][]*nodesCoordinator.SerializableValidator{},
				},
			},
		},
	}

	params, err := epochStartProvider.requestAndProcessing()
	assert.Equal(t, requiredParameters, params)
	assert.Nil(t, err)
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
			return &testscommon.HeadersCacherStub{}
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
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &validatorInfoCacherStub.ValidatorInfoCacherStub{}
		},
	}
	epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
	epochStartProvider.epochStartMeta = &block.MetaBlock{Epoch: 0}
	epochStartProvider.prevEpochStartMeta = &block.MetaBlock{}
	peerMiniBlocks, err := epochStartProvider.processNodesConfig([]byte("something"))
	assert.Nil(t, err)
	assert.Empty(t, peerMiniBlocks)
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
				UpdateSyncDataIfNeededCalled: func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, map[string]*block.MiniBlock, error) {
					return nil, nil, nil, expectedErr
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
				UpdateSyncDataIfNeededCalled: func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, map[string]*block.MiniBlock, error) {
					return expectedSyncData.ownShardHdr, expectedSyncData.additionalHeaders, nil, nil
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
				UpdateSyncDataIfNeededCalled: func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, map[string]*block.MiniBlock, error) {
					return expectedSyncData.ownShardHdr, expectedSyncData.additionalHeaders, nil, nil
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

func TestEpochStartBootstrap_ComputeAllPendingMiniblocks(t *testing.T) {
	t.Parallel()

	pendingMiniblocksHashes := [][]byte{
		[]byte("pending miniblock hash 1"),
		[]byte("pending miniblock hash 2"),
		[]byte("pending miniblock hash 3"),
		[]byte("pending miniblock hash 4"),
	}

	metablock := &block.MetaBlock{
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					HeaderHash: []byte("header hash 1"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: pendingMiniblocksHashes[0],
						},
						{
							Hash: pendingMiniblocksHashes[1],
						},
					},
				},
				{
					HeaderHash: []byte("header hash 2"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: pendingMiniblocksHashes[2],
						},
						{
							Hash: pendingMiniblocksHashes[3],
						},
					},
				},
			},
		},
	}

	e := &epochStartBootstrap{
		epochStartMeta: metablock,
	}

	allPendingMiniblocksHeaders := e.computeAllPendingMiniblocksHeaders()
	require.Equal(t, len(pendingMiniblocksHashes), len(allPendingMiniblocksHeaders))
	for i := 0; i < len(pendingMiniblocksHashes); i++ {
		assert.Equal(t, pendingMiniblocksHashes[i], allPendingMiniblocksHeaders[i].GetHash())
	}
}

func TestEpochStartBootstrap_Close(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.dataPool = &dataRetrieverMock.PoolsHolderStub{
		CloseCalled: func() error {
			return expectedErr
		}}

	err := epochStartProvider.Close()
	assert.Equal(t, expectedErr, err)
}

func TestSyncSetGuardianTransaction(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.shardCoordinator = mock.NewMultipleShardsCoordinatorMock()
	transactions := testscommon.NewShardedDataCacheNotifierMock()
	epochStartProvider.dataPool = &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &testscommon.HeadersCacherStub{}
		},
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return transactions
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
		PeerAuthenticationsCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
		HeartbeatsCalled: func() storage.Cacher {
			return testscommon.NewCacherStub()
		},
	}
	epochStartProvider.whiteListHandler = &testscommon.WhiteListHandlerStub{
		IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
			return true
		},
	}
	epochStartProvider.whiteListerVerifiedTxs = &testscommon.WhiteListHandlerStub{}
	epochStartProvider.requestHandler = &testscommon.RequestHandlerStub{}
	epochStartProvider.storageService = &storageMocks.ChainStorerStub{}

	err := epochStartProvider.createSyncers()
	assert.Nil(t, err)

	topicName := "transactions_0"
	interceptor, err := epochStartProvider.mainInterceptorContainer.Get(topicName)
	assert.Nil(t, err)

	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(0),
		GasPrice:  args.EconomicsData.MinGasPrice(),
		GasLimit:  args.EconomicsData.MinGasLimit() * 2,
		Data:      []byte("SetGuardian@aa@bb"),
		ChainID:   []byte(coreComp.ChainID()),
		Signature: bytes.Repeat([]byte("2"), 32),
		Version:   1,
	}
	txBytes, _ := coreComp.IntMarsh.Marshal(tx)

	batch := &dataBatch.Batch{
		Data: [][]byte{txBytes},
	}
	batchBytes, _ := coreComp.IntMarsh.Marshal(batch)

	msg := &p2pmocks.P2PMessageMock{
		FromField:      nil,
		DataField:      batchBytes,
		SeqNoField:     nil,
		TopicField:     "topicName",
		SignatureField: nil,
		KeyField:       nil,
		PeerField:      "",
		PayloadField:   nil,
		TimestampField: 0,
	}

	err = interceptor.ProcessReceivedMessage(msg, "pid", nil)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	txHash := coreComp.Hash.Compute(string(txBytes))
	_, found := transactions.SearchFirstData(txHash)
	assert.True(t, found)
}
