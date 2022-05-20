package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// GetGeneralConfig returns the common configuration used for testing
func GetGeneralConfig() config.Config {
	return config.Config{
		Hardfork: config.HardforkConfig{
			PublicKeyToListenFrom:     "153dae6cb3963260f309959bf285537b77ae16d82e9933147be7827f7394de8dc97d9d9af41e970bc72aecb44b77e819621081658c37f7000d21e2d0e8963df83233407bde9f46369ba4fcd03b57f40b80b06c191a428cfb5c447ec510e79307",
			CloseAfterExportInMinutes: 2,
		},
		PublicKeyPeerId: config.CacheConfig{
			Type:     "LRU",
			Capacity: 5000,
			Shards:   16,
		},
		PublicKeyShardId: config.CacheConfig{
			Type:     "LRU",
			Capacity: 5000,
			Shards:   16,
		},
		PeerIdShardId: config.CacheConfig{
			Type:     "LRU",
			Capacity: 5000,
			Shards:   16,
		},
		PeerHonesty: config.CacheConfig{
			Type:     "LRU",
			Capacity: 5000,
			Shards:   16,
		},
		AddressPubkeyConverter: config.PubkeyConfig{
			Length:          32,
			Type:            "bech32",
			SignatureLength: 0,
		},
		ValidatorPubkeyConverter: config.PubkeyConfig{
			Length:          96,
			Type:            "hex",
			SignatureLength: 48,
		},
		Consensus: config.ConsensusConfig{
			Type: "bls",
		},
		ValidatorStatistics: config.ValidatorStatisticsConfig{
			CacheRefreshIntervalInSec: uint32(100),
		},
		GeneralSettings: config.GeneralSettingsConfig{
			StartInEpochEnabled:                  true,
			GenesisMaxNumberOfShards:             100,
			MaxComputableRounds:                  1000,
			MaxConsecutiveRoundsOfRatingDecrease: 2000,
			SyncProcessTimeInMillis:              6000,
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
			Enabled:                     false,
			ValidatorCleanOldEpochsData: false,
			ObserverCleanOldEpochsData:  false,
			NumEpochsToKeep:             3,
			NumActivePersisters:         3,
		},
		EvictionWaitingList: config.EvictionWaitingListConfig{
			HashesSize:     100,
			RootHashesSize: 100,
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("EvictionWaitingList"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		AccountsTrieStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("AccountsTrie"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		AccountsTrieCheckpointsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("AccountsTrieCheckpoints"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerAccountsTrieStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("PeerAccountsTrie"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerAccountsTrieCheckpointsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("PeerAccountsTrieCheckpoints"),
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
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("TrieSync"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 2,
				MaxBatchSize:      1000,
				MaxOpenFiles:      10,
				UseTmpAsFilePath:  true,
			},
			Capacity:    10,
			SizeInBytes: 10000,
		},
		TrieNodesChunksDataPool: getLRUCacheConfig(),
		SmartContractDataPool:   getLRUCacheConfig(),
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
		HeartbeatV2: config.HeartbeatV2Config{
			PeerAuthenticationTimeBetweenSendsInSec:          1,
			PeerAuthenticationTimeBetweenSendsWhenErrorInSec: 1,
			PeerAuthenticationThresholdBetweenSends:          0.1,
			HeartbeatTimeBetweenSendsInSec:                   1,
			HeartbeatTimeBetweenSendsWhenErrorInSec:          1,
			HeartbeatThresholdBetweenSends:                   0.1,
			MaxNumOfPeerAuthenticationInResponse:             5,
			DelayBetweenConnectionNotificationsInSec:         5,
			HeartbeatExpiryTimespanInSec:                     30,
			MaxDurationPeerUnresponsiveInSec:                 10,
			HideInactiveValidatorIntervalInSec:               60,
			HardforkTimeBetweenSendsInSec:                    5,
			PeerAuthenticationPool: config.PeerAuthenticationPoolConfig{
				DefaultSpanInSec: 30,
				CacheExpiryInSec: 30,
			},
			HeartbeatPool: getLRUCacheConfig(),
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
		SmartContractsStorageSimulate: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("SmartContractsStorageSimulate"),
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
		LogsAndEvents: config.LogsAndEventsConfig{
			SaveInStorageEnabled: false,
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
		ScheduledSCRsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("ScheduledSCRs"),
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		Versions: config.VersionsConfig{
			Cache: config.CacheConfig{
				Type:     "LRU",
				Capacity: 1000,
				Shards:   1,
			},
			DefaultVersion: "default",
			VersionsByEpochs: []config.VersionByEpochs{
				{
					StartEpoch: 0,
					Version:    "*",
				},
			},
		},
		SoftwareVersionConfig: config.SoftwareVersionConfig{
			PollingIntervalInMinutes: 30,
		},
		TrieSync: config.TrieSyncConfig{
			NumConcurrentTrieSyncers:  50,
			MaxHardCapForMissingNodes: 500,
			TrieSyncerVersion:         2,
		},
		Antiflood: config.AntifloodConfig{
			NumConcurrentResolverJobs: 2,
			TxAccumulator: config.TxAccumulatorConfig{
				MaxAllowedTimeInMilliseconds:   10,
				MaxDeviationTimeInMilliseconds: 1,
			},
		},
		Resolvers: config.ResolverConfig{
			NumCrossShardPeers:  2,
			NumIntraShardPeers:  1,
			NumFullHistoryPeers: 3,
		},
		VirtualMachine: config.VirtualMachineServicesConfig{
			Execution: config.VirtualMachineConfig{
				ArwenVersions: []config.ArwenVersionByEpoch{
					{StartEpoch: 0, Version: "*"},
				},
			},
			Querying: config.QueryVirtualMachineConfig{
				NumConcurrentVMs: 1,
				VirtualMachineConfig: config.VirtualMachineConfig{
					ArwenVersions: []config.ArwenVersionByEpoch{
						{StartEpoch: 0, Version: "*"},
					},
				},
			},
		},
		VMOutputCacher: config.CacheConfig{
			Type:     "LRU",
			Capacity: 10000,
			Name:     "VMOutputCacher",
		},
		PeersRatingConfig: config.PeersRatingConfig{
			TopRatedCacheCapacity: 1000,
			BadRatedCacheCapacity: 1000,
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
