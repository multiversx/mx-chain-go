package testscommon

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
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
			Hrp:             "erd",
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
			SetGuardianEpochsDelay:               20,
			StatusPollingIntervalSec:             10,
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
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		AccountsTrieStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("AccountsTrie"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerAccountsTrieStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("PeerAccountsTrie"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		StateTriesConfig: config.StateTriesConfig{
			SnapshotsEnabled:            true,
			AccountsStatePruningEnabled: false,
			PeerStatePruningEnabled:     false,
			MaxStateTrieLevelInMemory:   5,
			MaxPeerTrieLevelInMemory:    5,
		},
		TrieStorageManagerConfig: config.TrieStorageManagerConfig{
			PruningBufferLen:      1000,
			SnapshotsBufferLen:    10,
			SnapshotsGoroutineNum: 2,
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
		ValidatorInfoPool: config.CacheConfig{
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
				Type:              string(storageunit.MemoryDB),
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
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MiniBlocksStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("MiniBlocks"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		ShardHdrNonceHashStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("ShardHdrHashNonce"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MetaBlockStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("MetaBlock"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MetaHdrNonceHashStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("MetaHdrHashNonce"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		UnsignedTransactionStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("UnsignedTransactions"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		RewardTxStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("RewardTransactions"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		BlockHeaderStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("BlockHeaders"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		HeartbeatV2: config.HeartbeatV2Config{
			PeerAuthenticationTimeBetweenSendsInSec:          1,
			PeerAuthenticationTimeBetweenSendsWhenErrorInSec: 1,
			PeerAuthenticationTimeThresholdBetweenSends:      0.1,
			HeartbeatTimeBetweenSendsInSec:                   1,
			HeartbeatTimeBetweenSendsDuringBootstrapInSec:    1,
			HeartbeatTimeBetweenSendsWhenErrorInSec:          1,
			HeartbeatTimeThresholdBetweenSends:               0.1,
			PeerShardTimeBetweenSendsInSec:                   5,
			PeerShardTimeThresholdBetweenSends:               0.1,
			HeartbeatExpiryTimespanInSec:                     30,
			MaxDurationPeerUnresponsiveInSec:                 10,
			HideInactiveValidatorIntervalInSec:               60,
			HardforkTimeBetweenSendsInSec:                    5,
			TimeBetweenConnectionsMetricsUpdateInSec:         10,
			PeerAuthenticationTimeBetweenChecksInSec:         1,
			HeartbeatPool:                                    getLRUCacheConfig(),
		},
		StatusMetricsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("StatusMetricsStorageDB"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		SmartContractsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("SmartContractsStorage"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		SmartContractsStorageSimulate: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("SmartContractsStorageSimulate"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerBlockBodyStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("PeerBlocks"),
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		BootstrapStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("BootstrapData"),
				Type:              string(storageunit.MemoryDB),
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
					Type:              string(storageunit.MemoryDB),
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
				Type:              string(storageunit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		ScheduledSCRsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          AddTimestampSuffix("ScheduledSCRs"),
				Type:              string(storageunit.MemoryDB),
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
			CheckNodesOnDisk:          false,
		},
		Antiflood: config.AntifloodConfig{
			NumConcurrentResolverJobs:           2,
			NumConcurrentResolvingTrieNodesJobs: 1,
			TxAccumulator: config.TxAccumulatorConfig{
				MaxAllowedTimeInMilliseconds:   10,
				MaxDeviationTimeInMilliseconds: 1,
			},
		},
		Requesters: config.RequesterConfig{
			NumCrossShardPeers:  2,
			NumTotalPeers:       3,
			NumFullHistoryPeers: 4,
		},
		VirtualMachine: config.VirtualMachineServicesConfig{
			Execution: config.VirtualMachineConfig{
				WasmVMVersions: []config.WasmVMVersionByEpoch{
					{StartEpoch: 0, Version: "*"},
				},
			},
			Querying: config.QueryVirtualMachineConfig{
				NumConcurrentVMs: 1,
				VirtualMachineConfig: config.VirtualMachineConfig{
					WasmVMVersions: []config.WasmVMVersionByEpoch{
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
		PoolsCleanersConfig: config.PoolsCleanersConfig{
			MaxRoundsToKeepUnprocessedMiniBlocks:   50,
			MaxRoundsToKeepUnprocessedTransactions: 50,
		},
		BuiltInFunctions: config.BuiltInFunctionsConfig{
			AutomaticCrawlerAddresses: []string{
				"erd1he8wwxn4az3j82p7wwqsdk794dm7hcrwny6f8dfegkfla34udx7qrf7xje", //shard 0
				"erd1fpkcgel4gcmh8zqqdt043yfcn5tyx8373kg6q2qmkxzu4dqamc0swts65c", //shard 1
				"erd1najnxxweyw6plhg8efql330nttrj6l5cf87wqsuym85s9ha0hmdqnqgenp", //shard 2
			},
			MaxNumAddressesInTransferRole: 100,
			DNSV2Addresses: []string{
				"erd1he8wwxn4az3j82p7wwqsdk794dm7hcrwny6f8dfegkfla34udx7qrf7xje", //shard 0
				"erd1fpkcgel4gcmh8zqqdt043yfcn5tyx8373kg6q2qmkxzu4dqamc0swts65c", //shard 1
				"erd1najnxxweyw6plhg8efql330nttrj6l5cf87wqsuym85s9ha0hmdqnqgenp", //shard 2
			},
		},
		ResourceStats: config.ResourceStatsConfig{
			RefreshIntervalInSec: 1,
		},
		SovereignConfig: config.SovereignConfig{
			NotifierConfig: config.NotifierConfig{
				SubscribedEvents: []config.SubscribedEvent{
					{
						Identifier: "bridgeOps",
						Addresses:  []string{"erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"},
					},
				},
			},
			OutgoingSubscribedEvents: config.OutgoingSubscribedEvents{
				SubscribedEvents: []config.SubscribedEvent{
					{
						Identifier: "bridgeOps",
						Addresses:  []string{"erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"},
					},
				},
			},
			OutGoingBridge: config.OutGoingBridge{
				Hasher: "sha256",
			},
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
