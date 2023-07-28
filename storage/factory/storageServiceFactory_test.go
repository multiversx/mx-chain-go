package factory

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgument(t *testing.T) StorageServiceFactoryArgs {
	pathMan, err := CreatePathManagerFromSinglePathString(t.TempDir())
	require.Nil(t, err)

	return StorageServiceFactoryArgs{
		Config: config.Config{
			StateTriesConfig: config.StateTriesConfig{},
			StoragePruning: config.StoragePruningConfig{
				Enabled:                    true,
				NumActivePersisters:        3,
				NumEpochsToKeep:            4,
				ObserverCleanOldEpochsData: true,
			},
			ShardHdrNonceHashStorage:           createMockStorageConfig("ShardHdrNonceHashStorage"),
			TxStorage:                          createMockStorageConfig("TxStorage"),
			UnsignedTransactionStorage:         createMockStorageConfig("UnsignedTransactionStorage"),
			RewardTxStorage:                    createMockStorageConfig("RewardTxStorage"),
			ReceiptsStorage:                    createMockStorageConfig("ReceiptsStorage"),
			ScheduledSCRsStorage:               createMockStorageConfig("ScheduledSCRsStorage"),
			BootstrapStorage:                   createMockStorageConfig("BootstrapStorage"),
			MiniBlocksStorage:                  createMockStorageConfig("MiniBlocksStorage"),
			BootstrapStaticStorage:             createMockStorageConfig("BootstrapStaticStorage"),
			MetaBlockStorage:                   createMockStorageConfig("MetaBlockStorage"),
			MetaHdrNonceHashStorage:            createMockStorageConfig("MetaHdrNonceHashStorage"),
			BlockHeaderStorage:                 createMockStorageConfig("BlockHeaderStorage"),
			AccountsTrieStorage:                createMockStorageConfig("AccountsTrieStorage"),
			AccountsTrieCheckpointsStorage:     createMockStorageConfig("AccountsTrieCheckpointsStorage"),
			PeerAccountsTrieStorage:            createMockStorageConfig("PeerAccountsTrieStorage"),
			PeerAccountsTrieCheckpointsStorage: createMockStorageConfig("PeerAccountsTrieCheckpointsStorage"),
			StatusMetricsStorage:               createMockStorageConfig("StatusMetricsStorage"),
			PeerBlockBodyStorage:               createMockStorageConfig("PeerBlockBodyStorage"),
			TrieEpochRootHashStorage:           createMockStorageConfig("TrieEpochRootHashStorage"),
			DbLookupExtensions: config.DbLookupExtensionsConfig{
				Enabled:                            true,
				DbLookupMaxActivePersisters:        10,
				MiniblocksMetadataStorageConfig:    createMockStorageConfig("MiniblocksMetadataStorage"),
				MiniblockHashByTxHashStorageConfig: createMockStorageConfig("MiniblockHashByTxHashStorage"),
				EpochByHashStorageConfig:           createMockStorageConfig("EpochByHashStorage"),
				ResultsHashesByTxHashStorageConfig: createMockStorageConfig("ResultsHashesByTxHashStorage"),
				ESDTSuppliesStorageConfig:          createMockStorageConfig("ESDTSuppliesStorage"),
				RoundHashStorageConfig:             createMockStorageConfig("RoundHashStorage"),
			},
			LogsAndEvents: config.LogsAndEventsConfig{
				SaveInStorageEnabled: true,
				TxLogsStorage:        createMockStorageConfig("TxLogsStorage"),
			},
		},
		PrefsConfig: config.PreferencesConfig{},
		ShardCoordinator: &mock.ShardCoordinatorMock{
			NumShards: 3,
		},
		PathManager:        pathMan,
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		NodeTypeProvider: &nodeTypeProviderMock.NodeTypeProviderStub{
			GetTypeCalled: func() core.NodeType {
				return core.NodeTypeObserver
			},
		},
		StorageType:                   ProcessStorageService,
		CurrentEpoch:                  0,
		CreateTrieEpochRootHashStorer: true,
		ManagedPeersHolder:            &testscommon.ManagedPeersHolderStub{},
	}
}

func createMockStorageConfig(dbName string) config.StorageConfig {
	return config.StorageConfig{
		Cache: config.CacheConfig{
			Type:     "LRU",
			Capacity: 1000,
		},
		DB: config.DBConfig{
			FilePath:          dbName,
			Type:              "LvlDBSerial",
			BatchDelaySeconds: 5,
			MaxBatchSize:      100,
			MaxOpenFiles:      10,
		},
	}
}

func TestNewStorageServiceFactory(t *testing.T) {
	t.Parallel()

	t.Run("invalid StoragePruning.NumActivePersisters should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.StoragePruning.NumActivePersisters = 0
		storageServiceFactory, err := NewStorageServiceFactory(args)
		assert.Equal(t, storage.ErrInvalidNumberOfActivePersisters, err)
		assert.Nil(t, storageServiceFactory)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.ShardCoordinator = nil
		storageServiceFactory, err := NewStorageServiceFactory(args)
		assert.Equal(t, storage.ErrNilShardCoordinator, err)
		assert.Nil(t, storageServiceFactory)
	})
	t.Run("nil path manager should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.PathManager = nil
		storageServiceFactory, err := NewStorageServiceFactory(args)
		assert.Equal(t, storage.ErrNilPathManager, err)
		assert.Nil(t, storageServiceFactory)
	})
	t.Run("nil epoch start notifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.EpochStartNotifier = nil
		storageServiceFactory, err := NewStorageServiceFactory(args)
		assert.Equal(t, storage.ErrNilEpochStartNotifier, err)
		assert.Nil(t, storageServiceFactory)
	})
	t.Run("invalid number of epochs to save should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.StoragePruning.NumEpochsToKeep = 1
		storageServiceFactory, err := NewStorageServiceFactory(args)
		assert.Equal(t, storage.ErrInvalidNumberOfEpochsToSave, err)
		assert.Nil(t, storageServiceFactory)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		storageServiceFactory, err := NewStorageServiceFactory(args)
		assert.Nil(t, err)
		assert.NotNil(t, storageServiceFactory)
	})
}

func TestStorageServiceFactory_CreateForShard(t *testing.T) {
	t.Parallel()

	expectedErrForCacheString := "not supported cache type"

	t.Run("wrong config for ShardHdrNonceHashStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.ShardHdrNonceHashStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for ShardHdrNonceHashStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for TxStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.TxStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for TxStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for UnsignedTransactionStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.UnsignedTransactionStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for UnsignedTransactionStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for RewardTxStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.RewardTxStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for RewardTxStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for ReceiptsStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.ReceiptsStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for ReceiptsStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for ScheduledSCRsStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.ScheduledSCRsStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for ScheduledSCRsStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for BootstrapStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.BootstrapStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for BootstrapStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for MiniBlocksStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.MiniBlocksStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for MiniBlocksStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for MetaBlockStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.MetaBlockStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for MetaBlockStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for MetaHdrNonceHashStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.MetaHdrNonceHashStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for MetaHdrNonceHashStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for BlockHeaderStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.BlockHeaderStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for BlockHeaderStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for AccountsTrieStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.AccountsTrieStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for AccountsTrieStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for AccountsTrieCheckpointsStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.AccountsTrieCheckpointsStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for AccountsTrieCheckpointsStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for PeerAccountsTrieStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.PeerAccountsTrieStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for PeerAccountsTrieStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for PeerAccountsTrieCheckpointsStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.PeerAccountsTrieCheckpointsStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for PeerAccountsTrieCheckpointsStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for StatusMetricsStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.StatusMetricsStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for StatusMetricsStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for PeerBlockBodyStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.PeerBlockBodyStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for PeerBlockBodyStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for TrieEpochRootHashStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.TrieEpochRootHashStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for TrieEpochRootHashStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for DbLookupExtensions.MiniblocksMetadataStorageConfig should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.DbLookupExtensions.MiniblocksMetadataStorageConfig.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for DbLookupExtensions.MiniblocksMetadataStorageConfig", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for DbLookupExtensions.MiniblockHashByTxHashStorageConfig should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.DbLookupExtensions.MiniblockHashByTxHashStorageConfig.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for DbLookupExtensions.MiniblockHashByTxHashStorageConfig", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for DbLookupExtensions.EpochByHashStorageConfig should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.DbLookupExtensions.EpochByHashStorageConfig.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for DbLookupExtensions.EpochByHashStorageConfig", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for DbLookupExtensions.ResultsHashesByTxHashStorageConfig should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.DbLookupExtensions.ResultsHashesByTxHashStorageConfig.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for DbLookupExtensions.ResultsHashesByTxHashStorageConfig", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for DbLookupExtensions.ESDTSuppliesStorageConfig should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.DbLookupExtensions.ESDTSuppliesStorageConfig.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for DbLookupExtensions.ESDTSuppliesStorageConfig", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for DbLookupExtensions.RoundHashStorageConfig should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.DbLookupExtensions.RoundHashStorageConfig.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for DbLookupExtensions.RoundHashStorageConfig", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for LogsAndEvents.TxLogsStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.LogsAndEvents.TxLogsStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for LogsAndEvents.TxLogsStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for bootstrap static storage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.BootstrapStaticStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Equal(t, expectedErrForCacheString+" for BootstrapStaticStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(storageService))
		allStorers := storageService.GetAllStorers()
		expectedStorers := 26
		assert.Equal(t, expectedStorers, len(allStorers))
		_ = storageService.CloseAll()
	})
	t.Run("should work without DbLookupExtensions", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.DbLookupExtensions.Enabled = false
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(storageService))
		allStorers := storageService.GetAllStorers()
		numDBLookupExtensionUnits := 6
		expectedStorers := 26 - numDBLookupExtensionUnits
		assert.Equal(t, expectedStorers, len(allStorers))
		_ = storageService.CloseAll()
	})
	t.Run("should work without TrieEpochRootHashStorage", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.CreateTrieEpochRootHashStorer = false
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForShard()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(storageService))
		allStorers := storageService.GetAllStorers()
		expectedStorers := 25 // we still have a storer for trie epoch root hash
		assert.Equal(t, expectedStorers, len(allStorers))
		_ = storageService.CloseAll()
	})
}

func TestStorageServiceFactory_CreateForMeta(t *testing.T) {
	t.Parallel()

	expectedErrForCacheString := "not supported cache type"

	t.Run("wrong config for ShardHdrNonceHashStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.ShardHdrNonceHashStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForMeta()
		assert.Equal(t, expectedErrForCacheString+" for ShardHdrNonceHashStorage on shard 0", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for AccountsTrieStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.AccountsTrieStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForMeta()
		assert.Equal(t, expectedErrForCacheString+" for AccountsTrieStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for DbLookupExtensions.RoundHashStorageConfig should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.DbLookupExtensions.RoundHashStorageConfig.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForMeta()
		assert.Equal(t, expectedErrForCacheString+" for DbLookupExtensions.RoundHashStorageConfig", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for LogsAndEvents.TxLogsStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.LogsAndEvents.TxLogsStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForMeta()
		assert.Equal(t, expectedErrForCacheString+" for LogsAndEvents.TxLogsStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("wrong config for Bootstrap static storage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		args.Config.BootstrapStaticStorage.Cache.Type = ""
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForMeta()
		assert.Equal(t, expectedErrForCacheString+" for BootstrapStaticStorage", err.Error())
		assert.True(t, check.IfNil(storageService))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgument(t)
		storageServiceFactory, _ := NewStorageServiceFactory(args)
		storageService, err := storageServiceFactory.CreateForMeta()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(storageService))
		allStorers := storageService.GetAllStorers()
		missingStorers := 2 // PeerChangesUnit and ShardHdrNonceHashDataUnit
		numShardHdrStorage := 3
		expectedStorers := 26 - missingStorers + numShardHdrStorage
		assert.Equal(t, expectedStorers, len(allStorers))
		_ = storageService.CloseAll()
	})
}
