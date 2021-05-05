package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// GetGeneralConfig returns the common configuration used for testing
func GetGeneralConfig() config.Config {
	return config.Config{
		GeneralSettings: config.GeneralSettingsConfig{
			StartInEpochEnabled:      true,
			GenesisMaxNumberOfShards: 100,
		},
		EpochStartConfig: config.EpochStartConfig{
			MinRoundsBetweenEpochs:            5,
			RoundsPerEpoch:                    10,
			MinNumConnectedPeersToStart:       2,
			MinNumOfPeersToConsiderBlockValid: 2,
		},
		WhiteListPool:          getLRUCacheConfig(),
		WhiteListerVerifiedTxs: getLRUCacheConfig(),
		StoragePruning: config.StoragePruningConfig{
			Enabled:             false,
			CleanOldEpochsData:  false,
			NumEpochsToKeep:     3,
			NumActivePersisters: 3,
		},
		EvictionWaitingList: config.EvictionWaitingListConfig{
			Size: 100,
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("EvictionWaitingList"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		TrieSnapshotDB: config.DBConfig{
			FilePath:          AddTimestampSuffix("TrieSnapshot"),
			Type:              string(storageUnit.MemoryDB),
			BatchDelaySeconds: 30,
			MaxBatchSize:      6,
			MaxOpenFiles:      10,
		},
		AccountsTrieStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("AccountsTrie/MainDB"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerAccountsTrieStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("PeerAccountsTrie/MainDB"),
				Type:              string(storageUnit.MemoryDB),
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
		TxBlockBodyDataPool:   getLRUCacheConfig(),
		PeerBlockBodyDataPool: getLRUCacheConfig(),
		TrieSyncStorage: config.TrieSyncStorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("TrieSync"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 2,
				MaxBatchSize:      1000,
				MaxOpenFiles:      10,
			},
			UseTmpAsFilePath: true,
		},
		SmartContractDataPool: getLRUCacheConfig(),
		TxStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("Transactions"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MiniBlocksStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("MiniBlocks"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		ShardHdrNonceHashStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("ShardHdrHashNonce"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MetaBlockStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("MetaBlock"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MetaHdrNonceHashStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("MetaHdrHashNonce"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		UnsignedTransactionStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("UnsignedTransactions"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		RewardTxStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("RewardTransactions"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		BlockHeaderStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("BlockHeaders"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		Heartbeat: config.HeartbeatConfig{
			HeartbeatStorage: config.StorageConfig{
				Cache: getLRUCacheConfig(),
				DB: config.DBConfig{
					FilePath:          AddTimestampSuffix("HeartbeatStorage"),
					Type:              string(storageUnit.MemoryDB),
					BatchDelaySeconds: 30,
					MaxBatchSize:      6,
					MaxOpenFiles:      10,
				},
			},
		},
		StatusMetricsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("StatusMetricsStorageDB"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		SmartContractsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("SmartContractsStorage"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerBlockBodyStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("PeerBlocks"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		BootstrapStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("BootstrapData"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 1,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		TxLogsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("Logs"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 2,
				MaxBatchSize:      100,
				MaxOpenFiles:      10,
			},
		},
		ReceiptsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("Receipts"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		Versions: config.VersionsConfig{
			DefaultVersion: "default",
			VersionsByEpochs: []config.VersionByEpochs{
				{
					StartEpoch: 0,
					Version:    "*",
				},
			},
		},
		TrieSync: config.TrieSyncConfig{
			NumConcurrentTrieSyncers:  50,
			MaxHardCapForMissingNodes: 500,
			TrieSyncerVersion:         2,
		},
	}
}

func getLRUCacheConfig() config.CacheConfig {
	return config.CacheConfig{
		Type:     "LRU",
		Capacity: 1000,
		Shards:   1,
	}
}
