package components

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/common"
	commonFactory "github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/factory"
	bootstrapComp "github.com/multiversx/mx-chain-go/factory/bootstrap"
	consensusComp "github.com/multiversx/mx-chain-go/factory/consensus"
	coreComp "github.com/multiversx/mx-chain-go/factory/core"
	cryptoComp "github.com/multiversx/mx-chain-go/factory/crypto"
	dataComp "github.com/multiversx/mx-chain-go/factory/data"
	"github.com/multiversx/mx-chain-go/factory/mock"
	networkComp "github.com/multiversx/mx-chain-go/factory/network"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	stateComp "github.com/multiversx/mx-chain-go/factory/state"
	statusComp "github.com/multiversx/mx-chain-go/factory/status"
	"github.com/multiversx/mx-chain-go/factory/statusCore"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/p2p"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/trie"
	logger "github.com/multiversx/mx-chain-logger-go"
	wasmConfig "github.com/multiversx/mx-chain-vm-v1_4-go/config"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("componentsMock")

// TestHasher -
const TestHasher = "blake2b"

// TestMarshalizer -
const TestMarshalizer = "json"

// SignedBlocksThreshold -
const SignedBlocksThreshold = 0.025

// ConsecutiveMissedBlocksPenalty -
const ConsecutiveMissedBlocksPenalty = 1.1

// DummyPk -
const DummyPk = "629e1245577afb7717ccb46b6ff3649bdd6a1311514ad4a7695da13f801cc277ee24e730a7fa8aa6c612159b4328db17" +
	"35692d0bded3a2264ba621d6bda47a981d60e17dd306d608e0875a0ba19639fb0844661f519472a175ca9ed2f33fbe16"

// DummySk -
const DummySk = "cea01c0bf060187d90394802ff223078e47527dc8aa33a922744fb1d06029c4b"

// LoadKeysFunc -
type LoadKeysFunc func(string, int) ([]byte, string, error)

// GetCoreArgs -
func GetCoreArgs() coreComp.CoreComponentsFactoryArgs {
	return coreComp.CoreComponentsFactoryArgs{
		Config: GetGeneralConfig(),
		ConfigPathsHolder: config.ConfigurationPathsHolder{
			GasScheduleDirectoryName: "../../cmd/node/config/gasSchedules",
		},
		RatingsConfig:   CreateDummyRatingsConfig(),
		EconomicsConfig: CreateDummyEconomicsConfig(),
		NodesConfig: config.NodesConfig{
			StartTime: 0,
			InitialNodes: []*config.InitialNodeConfig{
				{
					PubKey:  "227a5a5ec0c58171b7f4ee9ecc304ea7b176fb626741a25c967add76d6cd361d6995929f9b60a96237381091cefb1b061225e5bb930b40494a5ac9d7524fd67dfe478e5ccd80f17b093cff5722025761fb0217c39dbd5ae45e01eb5a3113be93",
					Address: "erd1ulhw20j7jvgfgak5p05kv667k5k9f320sgef5ayxkt9784ql0zssrzyhjp",
				},
				{
					PubKey:  "ef9522d654bc08ebf2725468f41a693aa7f3cf1cb93922cff1c8c81fba78274016010916f4a7e5b0855c430a724a2d0b3acd1fe8e61e37273a17d58faa8c0d3ef6b883a33ec648950469a1e9757b978d9ae662a019068a401cff56eea059fd08",
					Address: "erd17c4fs6mz2aa2hcvva2jfxdsrdknu4220496jmswer9njznt22eds0rxlr4",
				},
				{
					PubKey:  "e91ab494cedd4da346f47aaa1a3e792bea24fb9f6cc40d3546bc4ca36749b8bfb0164e40dbad2195a76ee0fd7fb7da075ecbf1b35a2ac20638d53ea5520644f8c16952225c48304bb202867e2d71d396bff5a5971f345bcfe32c7b6b0ca34c84",
					Address: "erd10d2gufxesrp8g409tzxljlaefhs0rsgjle3l7nq38de59txxt8csj54cd3",
				},
			},
		},
		WorkingDirectory:    "home",
		ChanStopNodeProcess: make(chan endProcess.ArgEndProcess),
		EpochConfig: config.EpochConfig{
			GasSchedule: config.GasScheduleConfig{
				GasScheduleByEpochs: []config.GasScheduleByEpochs{
					{
						StartEpoch: 0,
						FileName:   "gasScheduleV1.toml",
					},
				},
			},
		},
		RoundConfig: config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				"Example": {
					Round: "18446744073709551615",
				},
			},
		},
	}
}

// GetStatusCoreArgs -
func GetStatusCoreArgs(coreComponents factory.CoreComponentsHolder) statusCore.StatusCoreComponentsFactoryArgs {
	return statusCore.StatusCoreComponentsFactoryArgs{
		Config: GetGeneralConfig(),
		EpochConfig: config.EpochConfig{
			GasSchedule: config.GasScheduleConfig{
				GasScheduleByEpochs: []config.GasScheduleByEpochs{
					{
						StartEpoch: 0,
						FileName:   "gasScheduleV1.toml",
					},
				},
			},
		},
		RoundConfig: config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				"Example": {
					Round: "18446744073709551615",
				},
			},
		},
		RatingsConfig:   CreateDummyRatingsConfig(),
		EconomicsConfig: CreateDummyEconomicsConfig(),
		CoreComp:        coreComponents,
	}
}

// GetConsensusArgs -
func GetConsensusArgs(shardCoordinator sharding.Coordinator) consensusComp.ConsensusComponentsFactoryArgs {
	coreComponents := GetCoreComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
	networkComponents := GetNetworkComponents(cryptoComponents)
	stateComponents := GetStateComponents(coreComponents)
	dataComponents := GetDataComponents(coreComponents, shardCoordinator)
	processComponents := GetProcessComponents(
		shardCoordinator,
		coreComponents,
		networkComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
	)
	statusComponents := GetStatusComponents(
		coreComponents,
		networkComponents,
		stateComponents,
		shardCoordinator,
		processComponents.NodesCoordinator(),
	)

	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                coreComponents.SyncTimer(),
		Processor:                processComponents.BlockProcessor(),
		RoundTimeDurationHandler: coreComponents.RoundHandler(),
	}
	scheduledProcessor, _ := spos.NewScheduledProcessorWrapper(args)

	return consensusComp.ConsensusComponentsFactoryArgs{
		Config:               testscommon.GetGeneralConfig(),
		FlagsConfig:          config.ContextFlagsConfig{},
		BootstrapRoundIndex:  0,
		CoreComponents:       coreComponents,
		NetworkComponents:    networkComponents,
		CryptoComponents:     cryptoComponents,
		DataComponents:       dataComponents,
		ProcessComponents:    processComponents,
		StateComponents:      stateComponents,
		StatusComponents:     statusComponents,
		StatusCoreComponents: GetStatusCoreComponents(),
		ScheduledProcessor:   scheduledProcessor,
	}
}

// GetCryptoArgs -
func GetCryptoArgs(coreComponents factory.CoreComponentsHolder) cryptoComp.CryptoComponentsFactoryArgs {
	args := cryptoComp.CryptoComponentsFactoryArgs{
		Config: config.Config{
			GeneralSettings: config.GeneralSettingsConfig{ChainID: "undefined"},
			Consensus: config.ConsensusConfig{
				Type: "bls",
			},
			MultisigHasher: config.TypeConfig{Type: "blake2b"},
			PublicKeyPIDSignature: config.CacheConfig{
				Capacity: 1000,
				Type:     "LRU",
			},
			Hasher: config.TypeConfig{Type: "blake2b"},
		},
		SkIndex:                              0,
		ValidatorKeyPemFileName:              "validatorKey.pem",
		CoreComponentsHolder:                 coreComponents,
		ActivateBLSPubKeyMessageVerification: false,
		KeyLoader: &mock.KeyLoaderStub{
			LoadKeyCalled: DummyLoadSkPkFromPemFile([]byte(DummySk), DummyPk, nil),
		},
		EnableEpochs: config.EnableEpochs{
			BLSMultiSignerEnableEpoch: []config.MultiSignerConfig{{EnableEpoch: 0, Type: "no-KOSK"}},
		},
	}

	return args
}

// GetDataArgs -
func GetDataArgs(coreComponents factory.CoreComponentsHolder, shardCoordinator sharding.Coordinator) dataComp.DataComponentsFactoryArgs {
	return dataComp.DataComponentsFactoryArgs{
		Config: testscommon.GetGeneralConfig(),
		PrefsConfig: config.PreferencesConfig{
			FullArchive: false,
		},
		ShardCoordinator:              shardCoordinator,
		Core:                          coreComponents,
		StatusCore:                    GetStatusCoreComponents(),
		Crypto:                        GetCryptoComponents(coreComponents),
		CurrentEpoch:                  0,
		CreateTrieEpochRootHashStorer: false,
		NodeProcessingMode:            common.Normal,
		FlagsConfigs:                  config.ContextFlagsConfig{},
	}
}

// GetCoreComponents -
func GetCoreComponents() factory.CoreComponentsHolder {
	coreArgs := GetCoreArgs()
	coreComponentsFactory, _ := coreComp.NewCoreComponentsFactory(coreArgs)
	coreComponents, err := coreComp.NewManagedCoreComponents(coreComponentsFactory)
	if err != nil {
		fmt.Println("getCoreComponents NewManagedCoreComponents", "error", err.Error())
		return nil
	}
	err = coreComponents.Create()
	if err != nil {
		fmt.Println("getCoreComponents Create", "error", err.Error())
	}
	return coreComponents
}

// GetNetworkFactoryArgs -
func GetNetworkFactoryArgs() networkComp.NetworkComponentsFactoryArgs {
	p2pCfg := p2pConfig.P2PConfig{
		Node: p2pConfig.NodeConfig{
			Port: "0",
		},
		KadDhtPeerDiscovery: p2pConfig.KadDhtPeerDiscoveryConfig{
			Enabled:                          false,
			Type:                             "optimized",
			RefreshIntervalInSec:             10,
			ProtocolID:                       "erd/kad/1.0.0",
			InitialPeerList:                  []string{"peer0", "peer1"},
			BucketSize:                       10,
			RoutingTableRefreshIntervalInSec: 5,
		},
		Sharding: p2pConfig.ShardingConfig{
			TargetPeerCount:         10,
			MaxIntraShardValidators: 10,
			MaxCrossShardValidators: 10,
			MaxIntraShardObservers:  10,
			MaxCrossShardObservers:  10,
			MaxSeeders:              2,
			Type:                    "NilListSharder",
			AdditionalConnections: p2pConfig.AdditionalConnectionsConfig{
				MaxFullHistoryObservers: 10,
			},
		},
	}

	mainConfig := config.Config{
		PeerHonesty: config.CacheConfig{
			Type:     "LRU",
			Capacity: 5000,
			Shards:   16,
		},
		Debug: config.DebugConfig{
			Antiflood: config.AntifloodDebugConfig{
				Enabled:                    true,
				CacheSize:                  100,
				IntervalAutoPrintInSeconds: 1,
			},
		},
		PeersRatingConfig: config.PeersRatingConfig{
			TopRatedCacheCapacity: 1000,
			BadRatedCacheCapacity: 1000,
		},
		PoolsCleanersConfig: config.PoolsCleanersConfig{
			MaxRoundsToKeepUnprocessedMiniBlocks:   50,
			MaxRoundsToKeepUnprocessedTransactions: 50,
		},
	}

	appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()

	cryptoCompMock := GetDefaultCryptoComponents()

	return networkComp.NetworkComponentsFactoryArgs{
		P2pConfig:     p2pCfg,
		MainConfig:    mainConfig,
		StatusHandler: appStatusHandler,
		Marshalizer:   &mock.MarshalizerMock{},
		RatingsConfig: config.RatingsConfig{
			General:    config.General{},
			ShardChain: config.ShardChain{},
			MetaChain:  config.MetaChain{},
			PeerHonesty: config.PeerHonestyConfig{
				DecayCoefficient:             0.9779,
				DecayUpdateIntervalInSeconds: 10,
				MaxScore:                     100,
				MinScore:                     -100,
				BadPeerThreshold:             -80,
				UnitValue:                    1.0,
			},
		},
		Syncer:            &p2pFactory.LocalSyncTimer{},
		NodeOperationMode: p2p.NormalOperation,
		CryptoComponents:  cryptoCompMock,
	}
}

// GetStateFactoryArgs -
func GetStateFactoryArgs(coreComponents factory.CoreComponentsHolder) stateComp.StateComponentsFactoryArgs {
	tsm, _ := trie.NewTrieStorageManager(storage.GetStorageManagerArgs())
	storageManagerUser, _ := trie.NewTrieStorageManagerWithoutPruning(tsm)
	tsm, _ = trie.NewTrieStorageManager(storage.GetStorageManagerArgs())
	storageManagerPeer, _ := trie.NewTrieStorageManagerWithoutPruning(tsm)

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[dataRetriever.UserAccountsUnit.String()] = storageManagerUser
	trieStorageManagers[dataRetriever.PeerAccountsUnit.String()] = storageManagerPeer

	triesHolder := state.NewDataTriesHolder()
	trieUsers, _ := trie.NewTrie(storageManagerUser, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), coreComponents.EnableEpochsHandler(), 5)
	triePeers, _ := trie.NewTrie(storageManagerPeer, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), coreComponents.EnableEpochsHandler(), 5)
	triesHolder.Put([]byte(dataRetriever.UserAccountsUnit.String()), trieUsers)
	triesHolder.Put([]byte(dataRetriever.PeerAccountsUnit.String()), triePeers)

	stateComponentsFactoryArgs := stateComp.StateComponentsFactoryArgs{
		Config:         GetGeneralConfig(),
		Core:           coreComponents,
		StatusCore:     GetStatusCoreComponents(),
		StorageService: disabled.NewChainStorer(),
		ProcessingMode: common.Normal,
		ChainHandler:   &testscommon.ChainHandlerStub{},
	}

	return stateComponentsFactoryArgs
}

// GetProcessComponentsFactoryArgs -
func GetProcessComponentsFactoryArgs(shardCoordinator sharding.Coordinator) processComp.ProcessComponentsFactoryArgs {
	coreComponents := GetCoreComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
	networkComponents := GetNetworkComponents(cryptoComponents)
	dataComponents := GetDataComponents(coreComponents, shardCoordinator)
	stateComponents := GetStateComponents(coreComponents)
	processArgs := GetProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)
	return processArgs
}

//GetBootStrapFactoryArgs -
func GetBootStrapFactoryArgs() bootstrapComp.BootstrapComponentsFactoryArgs {
	coreComponents := GetCoreComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
	networkComponents := GetNetworkComponents(cryptoComponents)
	statusCoreComponents := GetStatusCoreComponents()
	return bootstrapComp.BootstrapComponentsFactoryArgs{
		Config:               testscommon.GetGeneralConfig(),
		WorkingDir:           "home",
		CoreComponents:       coreComponents,
		CryptoComponents:     cryptoComponents,
		NetworkComponents:    networkComponents,
		StatusCoreComponents: statusCoreComponents,
		PrefConfig: config.Preferences{
			Preferences: config.PreferencesConfig{
				DestinationShardAsObserver: "0",
				ConnectionWatcherType:      "print",
			},
		},
		ImportDbConfig: config.ImportDbConfig{
			IsImportDBMode: false,
		},
		RoundConfig: config.RoundConfig{},
		FlagsConfig: config.ContextFlagsConfig{
			ForceStartFromNetwork: false,
		},
	}
}

// GetProcessArgs -
func GetProcessArgs(
	shardCoordinator sharding.Coordinator,
	coreComponents factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	cryptoComponents factory.CryptoComponentsHolder,
	stateComponents factory.StateComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
) processComp.ProcessComponentsFactoryArgs {

	gasSchedule := wasmConfig.MakeGasMapForTests()
	// TODO: check if these could be initialized by MakeGasMapForTests()
	gasSchedule["BuiltInCost"]["SaveUserName"] = 1
	gasSchedule["BuiltInCost"]["SaveKeyValue"] = 1
	gasSchedule["BuiltInCost"]["ESDTTransfer"] = 1
	gasSchedule["BuiltInCost"]["ESDTBurn"] = 1
	gasSchedule[common.MetaChainSystemSCsCost] = FillGasMapMetaChainSystemSCsCosts(1)

	gasScheduleNotifier := &testscommon.GasScheduleNotifierMock{
		GasSchedule: gasSchedule,
	}

	nc := &shardingMocks.NodesCoordinatorMock{}
	statusComponents := GetStatusComponents(
		coreComponents,
		networkComponents,
		stateComponents,
		shardCoordinator,
		nc,
	)

	bootstrapComponentsFactoryArgs := GetBootStrapFactoryArgs()
	bootstrapComponentsFactory, _ := bootstrapComp.NewBootstrapComponentsFactory(bootstrapComponentsFactoryArgs)
	bootstrapComponents, _ := bootstrapComp.NewTestManagedBootstrapComponents(bootstrapComponentsFactory)
	_ = bootstrapComponents.Create()
	_ = bootstrapComponents.SetShardCoordinator(shardCoordinator)

	return processComp.ProcessComponentsFactoryArgs{
		Config: testscommon.GetGeneralConfig(),
		AccountsParser: &mock.AccountsParserStub{
			InitialAccountsCalled: func() []genesis.InitialAccountHandler {
				addrConverter, _ := commonFactory.NewPubkeyConverter(config.PubkeyConfig{
					Length:          32,
					Type:            "bech32",
					SignatureLength: 0,
					Hrp:             "erd",
				})
				balance := big.NewInt(0)
				acc1 := data.InitialAccount{
					Address:      "erd1ulhw20j7jvgfgak5p05kv667k5k9f320sgef5ayxkt9784ql0zssrzyhjp",
					Supply:       big.NewInt(0).Mul(big.NewInt(2500000000), big.NewInt(1000000000000)),
					Balance:      balance,
					StakingValue: big.NewInt(0).Mul(big.NewInt(2500000000), big.NewInt(1000000000000)),
					Delegation: &data.DelegationData{
						Address: "",
						Value:   big.NewInt(0),
					},
				}
				acc2 := data.InitialAccount{
					Address:      "erd17c4fs6mz2aa2hcvva2jfxdsrdknu4220496jmswer9njznt22eds0rxlr4",
					Supply:       big.NewInt(0).Mul(big.NewInt(2500000000), big.NewInt(1000000000000)),
					Balance:      balance,
					StakingValue: big.NewInt(0).Mul(big.NewInt(2500000000), big.NewInt(1000000000000)),
					Delegation: &data.DelegationData{
						Address: "",
						Value:   big.NewInt(0),
					},
				}
				acc3 := data.InitialAccount{
					Address:      "erd10d2gufxesrp8g409tzxljlaefhs0rsgjle3l7nq38de59txxt8csj54cd3",
					Supply:       big.NewInt(0).Mul(big.NewInt(2500000000), big.NewInt(1000000000000)),
					Balance:      balance,
					StakingValue: big.NewInt(0).Mul(big.NewInt(2500000000), big.NewInt(1000000000000)),
					Delegation: &data.DelegationData{
						Address: "",
						Value:   big.NewInt(0),
					},
				}

				acc1Bytes, _ := addrConverter.Decode(acc1.Address)
				acc1.SetAddressBytes(acc1Bytes)
				acc2Bytes, _ := addrConverter.Decode(acc2.Address)
				acc2.SetAddressBytes(acc2Bytes)
				acc3Bytes, _ := addrConverter.Decode(acc3.Address)
				acc3.SetAddressBytes(acc3Bytes)
				initialAccounts := []genesis.InitialAccountHandler{&acc1, &acc2, &acc3}

				return initialAccounts
			},
			GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*block.MiniBlock, map[uint32]*outport.TransactionPool, error) {
				txsPool := make(map[uint32]*outport.TransactionPool)
				for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
					txsPool[i] = &outport.TransactionPool{}
				}

				return make([]*block.MiniBlock, 4), txsPool, nil
			},
		},
		SmartContractParser:    &mock.SmartContractParserStub{},
		GasSchedule:            gasScheduleNotifier,
		NodesCoordinator:       nc,
		Data:                   dataComponents,
		CoreData:               coreComponents,
		Crypto:                 cryptoComponents,
		State:                  stateComponents,
		Network:                networkComponents,
		StatusComponents:       statusComponents,
		BootstrapComponents:    bootstrapComponents,
		StatusCoreComponents:   GetStatusCoreComponents(),
		RequestedItemsHandler:  &testscommon.RequestedItemsHandlerStub{},
		WhiteListHandler:       &testscommon.WhiteListHandlerStub{},
		WhiteListerVerifiedTxs: &testscommon.WhiteListHandlerStub{},
		MaxRating:              100,
		ImportStartHandler:     &testscommon.ImportStartHandlerStub{},
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "1000",
				OwnerAddress:    "erd1fpkcgel4gcmh8zqqdt043yfcn5tyx8373kg6q2qmkxzu4dqamc0swts65c",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				V1: config.GovernanceSystemSCConfigV1{
					ProposalCost:     "500",
					NumNodes:         100,
					MinQuorum:        50,
					MinPassThreshold: 50,
					MinVetoThreshold: 50,
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
					LostProposalFee:  "1",
				},
				OwnerAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "2500000000000000000000",
				MinStakeValue:                        "1",
				UnJailValue:                          "1",
				MinStepValue:                         "1",
				UnBondPeriod:                         0,
				NumRoundsWithoutBleed:                0,
				MaximumPercentageToBleed:             0,
				BleedPercentagePerRound:              0,
				MaxNumberOfNodesForStake:             10,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "100",
				MinStakeAmount:      "100",
				ConfigChangeAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 100,
			},
		},
		HistoryRepo: &dblookupext.HistoryRepositoryStub{},
		FlagsConfig: config.ContextFlagsConfig{
			Version: "v1.0.0",
		},
	}
}

// GetStatusComponents -
func GetStatusComponents(
	coreComponents factory.CoreComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
	stateComponents factory.StateComponentsHolder,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
) factory.StatusComponentsHandler {
	indexerURL := "url"
	elasticUsername := "user"
	elasticPassword := "pass"
	statusArgs := statusComp.StatusComponentsFactoryArgs{
		Config: testscommon.GetGeneralConfig(),
		ExternalConfig: config.ExternalConfig{
			ElasticSearchConnector: config.ElasticSearchConfig{
				Enabled:        false,
				URL:            indexerURL,
				Username:       elasticUsername,
				Password:       elasticPassword,
				EnabledIndexes: []string{"transactions", "blocks"},
			},
			EventNotifierConnector: config.EventNotifierConfig{
				Enabled:        false,
				ProxyUrl:       "https://localhost:5000",
				MarshallerType: "json",
			},
		},
		EconomicsConfig:      config.EconomicsConfig{},
		ShardCoordinator:     shardCoordinator,
		NodesCoordinator:     nodesCoordinator,
		EpochStartNotifier:   coreComponents.EpochStartNotifierWithConfirm(),
		CoreComponents:       coreComponents,
		NetworkComponents:    networkComponents,
		StateComponents:      stateComponents,
		IsInImportMode:       false,
		StatusCoreComponents: GetStatusCoreComponents(),
	}

	statusComponentsFactory, _ := statusComp.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, err := statusComp.NewManagedStatusComponents(statusComponentsFactory)
	if err != nil {
		log.Error("getStatusComponents NewManagedStatusComponents", "error", err.Error())
		return nil
	}
	err = managedStatusComponents.Create()
	if err != nil {
		log.Error("getStatusComponents Create", "error", err.Error())
		return nil
	}
	return managedStatusComponents
}

// GetStatusComponentsFactoryArgsAndProcessComponents -
func GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator sharding.Coordinator) (statusComp.StatusComponentsFactoryArgs, factory.ProcessComponentsHolder) {
	coreComponents := GetCoreComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
	networkComponents := GetNetworkComponents(cryptoComponents)
	dataComponents := GetDataComponents(coreComponents, shardCoordinator)
	stateComponents := GetStateComponents(coreComponents)
	processComponents := GetProcessComponents(
		shardCoordinator,
		coreComponents,
		networkComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
	)
	statusCoreComponents := GetStatusCoreComponents()

	indexerURL := "url"
	elasticUsername := "user"
	elasticPassword := "pass"
	return statusComp.StatusComponentsFactoryArgs{
		Config: testscommon.GetGeneralConfig(),
		ExternalConfig: config.ExternalConfig{
			ElasticSearchConnector: config.ElasticSearchConfig{
				Enabled:        false,
				URL:            indexerURL,
				Username:       elasticUsername,
				Password:       elasticPassword,
				EnabledIndexes: []string{"transactions", "blocks"},
			},
			EventNotifierConnector: config.EventNotifierConfig{
				Enabled:           false,
				ProxyUrl:          "http://localhost:5000",
				RequestTimeoutSec: 30,
				MarshallerType:    "json",
			},
			HostDriverConfig: config.HostDriverConfig{
				MarshallerType:     "json",
				Mode:               "client",
				URL:                "localhost:12345",
				RetryDurationInSec: 1,
			},
		},
		EconomicsConfig:      config.EconomicsConfig{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(2),
		NodesCoordinator:     &shardingMocks.NodesCoordinatorMock{},
		EpochStartNotifier:   &mock.EpochStartNotifierStub{},
		CoreComponents:       coreComponents,
		NetworkComponents:    networkComponents,
		StateComponents:      stateComponents,
		StatusCoreComponents: statusCoreComponents,
		IsInImportMode:       false,
	}, processComponents
}

// GetNetworkComponents -
func GetNetworkComponents(cryptoComp factory.CryptoComponentsHolder) factory.NetworkComponentsHolder {
	networkArgs := GetNetworkFactoryArgs()
	networkArgs.CryptoComponents = cryptoComp
	networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
	networkComponents, _ := networkComp.NewManagedNetworkComponents(networkComponentsFactory)

	_ = networkComponents.Create()

	return networkComponents
}

// GetDataComponents -
func GetDataComponents(coreComponents factory.CoreComponentsHolder, shardCoordinator sharding.Coordinator) factory.DataComponentsHolder {
	dataArgs := GetDataArgs(coreComponents, shardCoordinator)
	dataComponentsFactory, err := dataComp.NewDataComponentsFactory(dataArgs)
	if err != nil {
		fmt.Println(err.Error())
	}
	dataComponents, _ := dataComp.NewManagedDataComponents(dataComponentsFactory)
	_ = dataComponents.Create()
	return dataComponents
}

// GetCryptoComponents -
func GetCryptoComponents(coreComponents factory.CoreComponentsHolder) factory.CryptoComponentsHolder {
	cryptoArgs := GetCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := cryptoComp.NewCryptoComponentsFactory(cryptoArgs)
	cryptoComponents, err := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
	if err != nil {
		log.Error("getCryptoComponents NewManagedCryptoComponents", "error", err.Error())
		return nil
	}

	err = cryptoComponents.Create()
	if err != nil {
		log.Error("getCryptoComponents Create", "error", err.Error())
		return nil
	}
	return cryptoComponents
}

// GetStateComponents -
func GetStateComponents(coreComponents factory.CoreComponentsHolder) factory.StateComponentsHolder {
	stateArgs := GetStateFactoryArgs(coreComponents)
	stateComponentsFactory, err := stateComp.NewStateComponentsFactory(stateArgs)
	if err != nil {
		log.Error("getStateComponents NewStateComponentsFactory", "error", err.Error())
		return nil
	}

	stateComponents, err := stateComp.NewManagedStateComponents(stateComponentsFactory)
	if err != nil {
		log.Error("getStateComponents NewManagedStateComponents", "error", err.Error())
		return nil
	}
	err = stateComponents.Create()
	if err != nil {
		log.Error("getStateComponents Create", "error", err.Error())
		return nil
	}
	return stateComponents
}

// GetStatusCoreComponents -
func GetStatusCoreComponents() factory.StatusCoreComponentsHolder {
	args := GetStatusCoreArgs(GetCoreComponents())
	statusCoreFactory, err := statusCore.NewStatusCoreComponentsFactory(args)
	if err != nil {
		log.Error("GetStatusCoreComponents NewStatusCoreComponentsFactory", "error", err.Error())
		return nil
	}

	statusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreFactory)
	if err != nil {
		log.Error("GetStatusCoreComponents NewManagedStatusCoreComponents", "error", err.Error())
		return nil
	}

	err = statusCoreComponents.Create()
	if err != nil {
		log.Error("statusCoreComponents Create", "error", err.Error())
		return nil
	}

	return statusCoreComponents
}

// GetProcessComponents -
func GetProcessComponents(
	shardCoordinator sharding.Coordinator,
	coreComponents factory.CoreComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	cryptoComponents factory.CryptoComponentsHolder,
	stateComponents factory.StateComponentsHolder,
) factory.ProcessComponentsHolder {
	processArgs := GetProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)
	processComponentsFactory, _ := processComp.NewProcessComponentsFactory(processArgs)
	managedProcessComponents, err := processComp.NewManagedProcessComponents(processComponentsFactory)
	if err != nil {
		log.Error("getProcessComponents NewManagedProcessComponents", "error", err.Error())
		return nil
	}
	err = managedProcessComponents.Create()
	if err != nil {
		log.Error("getProcessComponents Create", "error", err.Error())
		return nil
	}
	return managedProcessComponents
}

// DummyLoadSkPkFromPemFile -
func DummyLoadSkPkFromPemFile(sk []byte, pk string, err error) LoadKeysFunc {
	return func(_ string, _ int) ([]byte, string, error) {
		return sk, pk, err
	}
}

// FillGasMapMetaChainSystemSCsCosts -
func FillGasMapMetaChainSystemSCsCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["Stake"] = value
	gasMap["UnStake"] = value
	gasMap["UnBond"] = value
	gasMap["Claim"] = value
	gasMap["Get"] = value
	gasMap["ChangeRewardAddress"] = value
	gasMap["ChangeValidatorKeys"] = value
	gasMap["UnJail"] = value
	gasMap["ESDTIssue"] = value
	gasMap["ESDTOperations"] = value
	gasMap["Proposal"] = value
	gasMap["Vote"] = value
	gasMap["DelegateVote"] = value
	gasMap["RevokeVote"] = value
	gasMap["CloseProposal"] = value
	gasMap["DelegationOps"] = value
	gasMap["UnStakeTokens"] = value
	gasMap["UnBondTokens"] = value
	gasMap["DelegationMgrOps"] = value
	gasMap["GetAllNodeStates"] = value
	gasMap["ValidatorToDelegation"] = value
	gasMap["GetActiveFund"] = value
	gasMap["FixWaitingListSize"] = value

	return gasMap
}

// SetShardCoordinator -
func SetShardCoordinator(t *testing.T, bootstrapComponents factory.BootstrapComponentsHolder, coordinator sharding.Coordinator) {
	type testBootstrapComponents interface {
		SetShardCoordinator(shardCoordinator sharding.Coordinator) error
	}

	testBootstrap, ok := bootstrapComponents.(testBootstrapComponents)
	require.True(t, ok)

	_ = testBootstrap.SetShardCoordinator(coordinator)
}
