package components

import (
	"github.com/multiversx/mx-chain-go/config"
)

// GetGeneralConfig -
func GetGeneralConfig() config.Config {
	return config.Config{
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
		StateTriesConfig: config.StateTriesConfig{
			AccountsStatePruningEnabled: true,
			PeerStatePruningEnabled:     true,
			MaxStateTrieLevelInMemory:   5,
			MaxPeerTrieLevelInMemory:    5,
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
			PruningBufferLen:      1000,
			SnapshotsBufferLen:    10,
			SnapshotsGoroutineNum: 1,
		},
		VirtualMachine: config.VirtualMachineServicesConfig{
			Querying: config.QueryVirtualMachineConfig{
				NumConcurrentVMs: 1,
				VirtualMachineConfig: config.VirtualMachineConfig{
					WasmVMVersions: []config.WasmVMVersionByEpoch{
						{StartEpoch: 0, Version: "v0.3"},
					},
				},
			},
			Execution: config.VirtualMachineConfig{
				WasmVMVersions: []config.WasmVMVersionByEpoch{
					{StartEpoch: 0, Version: "v0.3"},
				},
			},
			GasConfig: config.VirtualMachineGasConfig{
				ShardMaxGasPerVmQuery: 1_500_000_000,
				MetaMaxGasPerVmQuery:  0,
			},
		},
		SmartContractsStorageForSCQuery: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
		},
		SmartContractDataPool: config.CacheConfig{
			Capacity: 10000,
			Type:     "LRU",
			Shards:   1,
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
			DNSV2Addresses: []string{
				"erd1he8wwxn4az3j82p7wwqsdk794dm7hcrwny6f8dfegkfla34udx7qrf7xje", //shard 0
				"erd1fpkcgel4gcmh8zqqdt043yfcn5tyx8373kg6q2qmkxzu4dqamc0swts65c", //shard 1
				"erd1najnxxweyw6plhg8efql330nttrj6l5cf87wqsuym85s9ha0hmdqnqgenp", //shard 2
			},
			MaxNumAddressesInTransferRole: 100,
		},
		EpochStartConfig: GetEpochStartConfig(),
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
		GeneralSettings: config.GeneralSettingsConfig{
			ChainID:                  "undefined",
			MinTransactionVersion:    1,
			GenesisMaxNumberOfShards: 3,
			SetGuardianEpochsDelay:   20,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           TestMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: TestHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: TestMarshalizer,
		},
		TxSignMarshalizer: config.TypeConfig{
			Type: TestMarshalizer,
		},
		TxSignHasher: config.TypeConfig{
			Type: TestHasher,
		},
		Consensus: config.ConsensusConfig{
			Type: "bls",
		},
		ValidatorStatistics: config.ValidatorStatisticsConfig{
			CacheRefreshIntervalInSec: uint32(100),
		},
		SoftwareVersionConfig: config.SoftwareVersionConfig{
			PollingIntervalInMinutes: 30,
		},
		Versions: config.VersionsConfig{
			DefaultVersion:   "1",
			VersionsByEpochs: nil,
			Cache: config.CacheConfig{
				Type:     "LRU",
				Capacity: 1000,
				Shards:   1,
			},
		},
		Hardfork: config.HardforkConfig{
			PublicKeyToListenFrom: DummyPk,
		},
		HeartbeatV2: config.HeartbeatV2Config{
			HeartbeatExpiryTimespanInSec: 10,
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
		},
	}
}

// GetEpochStartConfig -
func GetEpochStartConfig() config.EpochStartConfig {
	return config.EpochStartConfig{
		MinRoundsBetweenEpochs:            20,
		RoundsPerEpoch:                    20,
		MaxShuffledOutRestartThreshold:    0.2,
		MinShuffledOutRestartThreshold:    0.1,
		MinNumConnectedPeersToStart:       2,
		MinNumOfPeersToConsiderBlockValid: 2,
	}
}

// CreateDummyEconomicsConfig -
func CreateDummyEconomicsConfig() config.EconomicsConfig {
	return config.EconomicsConfig{
		GlobalSettings: config.GlobalSettings{
			GenesisTotalSupply: "20000000000000000000000000",
			MinimumInflation:   0,
			YearSettings: []*config.YearSetting{
				{
					Year:             0,
					MaximumInflation: 0.01,
				},
			},
		},
		RewardsSettings: config.RewardsSettings{
			RewardsConfigByEpoch: []config.EpochRewardSettings{
				{
					LeaderPercentage:                 0.1,
					ProtocolSustainabilityPercentage: 0.1,
					ProtocolSustainabilityAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
					TopUpFactor:                      0.25,
					TopUpGradientPoint:               "3000000000000000000000000",
				},
			},
		},
		FeeSettings: config.FeeSettings{
			GasLimitSettings: []config.GasLimitSetting{
				{
					MaxGasLimitPerBlock:         "1500000000",
					MaxGasLimitPerMiniBlock:     "1500000000",
					MaxGasLimitPerMetaBlock:     "15000000000",
					MaxGasLimitPerMetaMiniBlock: "15000000000",
					MaxGasLimitPerTx:            "1500000000",
					MinGasLimit:                 "50000",
					ExtraGasLimitGuardedTx:      "50000",
				},
			},
			MinGasPrice:            "1000000000",
			GasPerDataByte:         "1500",
			GasPriceModifier:       1,
			MaxGasPriceSetGuardian: "100000",
		},
	}
}

// CreateDummyRatingsConfig -
func CreateDummyRatingsConfig() config.RatingsConfig {
	return config.RatingsConfig{
		General: config.General{
			StartRating:           5000001,
			MaxRating:             10000000,
			MinRating:             1,
			SignedBlocksThreshold: SignedBlocksThreshold,
			SelectionChances: []*config.SelectionChance{
				{MaxThreshold: 0, ChancePercent: 5},
				{MaxThreshold: 2500000, ChancePercent: 19},
				{MaxThreshold: 7500000, ChancePercent: 20},
				{MaxThreshold: 10000000, ChancePercent: 21},
			},
		},
		ShardChain: config.ShardChain{
			RatingSteps: config.RatingSteps{
				HoursToMaxRatingFromStartRating: 2,
				ProposerValidatorImportance:     1,
				ProposerDecreaseFactor:          -4,
				ValidatorDecreaseFactor:         -4,
				ConsecutiveMissedBlocksPenalty:  ConsecutiveMissedBlocksPenalty,
			},
		},
		MetaChain: config.MetaChain{
			RatingSteps: config.RatingSteps{
				HoursToMaxRatingFromStartRating: 2,
				ProposerValidatorImportance:     1,
				ProposerDecreaseFactor:          -4,
				ValidatorDecreaseFactor:         -4,
				ConsecutiveMissedBlocksPenalty:  ConsecutiveMissedBlocksPenalty,
			},
		},
	}
}
