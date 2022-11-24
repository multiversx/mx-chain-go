package components

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	commonFactory "github.com/ElrondNetwork/elrond-go/common/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/factory"
	bootstrapComp "github.com/ElrondNetwork/elrond-go/factory/bootstrap"
	consensusComp "github.com/ElrondNetwork/elrond-go/factory/consensus"
	coreComp "github.com/ElrondNetwork/elrond-go/factory/core"
	cryptoComp "github.com/ElrondNetwork/elrond-go/factory/crypto"
	dataComp "github.com/ElrondNetwork/elrond-go/factory/data"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	networkComp "github.com/ElrondNetwork/elrond-go/factory/network"
	processComp "github.com/ElrondNetwork/elrond-go/factory/processing"
	stateComp "github.com/ElrondNetwork/elrond-go/factory/state"
	statusComp "github.com/ElrondNetwork/elrond-go/factory/status"
	"github.com/ElrondNetwork/elrond-go/factory/statusCore"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/p2p"
	p2pConfig "github.com/ElrondNetwork/elrond-go/p2p/config"
	p2pFactory "github.com/ElrondNetwork/elrond-go/p2p/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/ElrondNetwork/elrond-go/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	arwenConfig "github.com/ElrondNetwork/wasm-vm-v1_4/config"
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
		NodesConfig: sharding.NodesSetupDTO{
			StartTime:     0,
			RoundDuration: 4000,
			ShardConsensus: []sharding.ConsensusConfiguration{
				{
					EnableEpoch:        0,
					MinNodes:           1,
					ConsensusGroupSize: 1,
				},
			},
			MetaConsensus: []sharding.ConsensusConfiguration{
				{
					EnableEpoch:        0,
					MinNodes:           1,
					ConsensusGroupSize: 1,
				},
			},
			Hysteresis: 0,
			Adaptivity: false,
			InitialNodes: []*sharding.InitialNode{
				{
					PubKey:  "41378f754e2c7b2745208c3ed21b151d297acdc84c3aca00b9e292cf28ec2d444771070157ea7760ed83c26f4fed387d0077e00b563a95825dac2cbc349fc0025ccf774e37b0a98ad9724d30e90f8c29b4091ccb738ed9ffc0573df776ee9ea30b3c038b55e532760ea4a8f152f2a52848020e5cee1cc537f2c2323399723081",
					Address: "9e95a4e46da335a96845b4316251fc1bb197e1b8136d96ecc62bf6604eca9e49",
				},
				{
					PubKey:  "52f3bf5c01771f601ec2137e267319ab6716ef6ff5dfddaea48b42d955f631167f2ce19296a202bb8fd174f4e94f8c85f619df85a7f9f8de0f3768e5e6d8c48187b767deccf9829be246aa331aa86d182eb8fa28ea8a3e45d357ed1647a9be020a5569d686253a6f89e9123c7f21f302e82f67d3e3cd69cf267b9910a663ef32",
					Address: "7a330039e77ca06bc127319fd707cc4911a80db489a39fcfb746283a05f61836",
				},
				{
					PubKey:  "5e91c426c5c8f5f805f86de1e0653e2ec33853772e583b88e9f0f201089d03d8570759c3c3ab610ce573493c33ba0adf954c8939dba5d5ef7f2be4e87145d8153fc5b4fb91cecb8d9b1f62e080743fbf69c8c3096bf07980bb82cb450ba9b902673373d5b671ea73620cc5bc4d36f7a0f5ca3684d4c8aa5c1b425ab2a8673140",
					Address: "131e2e717f2d33bdf7850c12b03dfe41ea8a5e76fdd6d4f23aebe558603e746f",
				},
				{
					PubKey:  "73972bf46dca59fba211c58f11b530f8e9d6392c499655ce760abc6458fd9c6b54b9676ee4b95aa32f6c254c9aad2f63a6195cd65d837a4320d7b8e915ba3a7123c8f4983b201035573c0752bb54e9021eb383b40d302447b62ea7a3790c89c47f5ab81d183f414e87611a31ff635ad22e969495356d5bc44eec7917aaad4c5e",
					Address: "4c9e66b605882c1099088f26659692f084e41dc0dedfaedf6a6409af21c02aac",
				},
				{
					PubKey:  "7391ccce066ab5674304b10220643bc64829afa626a165f1e7a6618e260fa68f8e79018ac5964f7a1b8dd419645049042e34ebe7f2772def71e6176ce9daf50a57c17ee2a7445b908fe47e8f978380fcc2654a19925bf73db2402b09dde515148081f8ca7c331fbedec689de1b7bfce6bf106e4433557c29752c12d0a009f47a",
					Address: "90a66900634b206d20627fbaec432ebfbabeaf30b9e338af63191435e2e37022",
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
	networkComponents := GetNetworkComponents()
	stateComponents := GetStateComponents(coreComponents, shardCoordinator)
	cryptoComponents := GetCryptoComponents(coreComponents)
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
		dataComponents,
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
		EpochStartNotifier:            &mock.EpochStartNotifierStub{},
		CurrentEpoch:                  0,
		CreateTrieEpochRootHashStorer: false,
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
	}
}

func getNewTrieStorageManagerArgs() trie.NewTrieStorageManagerArgs {
	return trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.CreateMemUnit(),
		CheckpointsStorer:      testscommon.CreateMemUnit(),
		Marshalizer:            &mock.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		GeneralConfig:          config.TrieStorageManagerConfig{SnapshotsGoroutineNum: 1},
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10, 32),
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
	}
}

// GetStateFactoryArgs -
func GetStateFactoryArgs(coreComponents factory.CoreComponentsHolder, shardCoordinator sharding.Coordinator) stateComp.StateComponentsFactoryArgs {
	tsm, _ := trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
	storageManagerUser, _ := trie.NewTrieStorageManagerWithoutPruning(tsm)
	tsm, _ = trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
	storageManagerPeer, _ := trie.NewTrieStorageManagerWithoutPruning(tsm)

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[trieFactory.UserAccountTrie] = storageManagerUser
	trieStorageManagers[trieFactory.PeerAccountTrie] = storageManagerPeer

	triesHolder := state.NewDataTriesHolder()
	trieUsers, _ := trie.NewTrie(storageManagerUser, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), 5)
	triePeers, _ := trie.NewTrie(storageManagerPeer, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), 5)
	triesHolder.Put([]byte(trieFactory.UserAccountTrie), trieUsers)
	triesHolder.Put([]byte(trieFactory.PeerAccountTrie), triePeers)

	stateComponentsFactoryArgs := stateComp.StateComponentsFactoryArgs{
		Config:           GetGeneralConfig(),
		ShardCoordinator: shardCoordinator,
		Core:             coreComponents,
		StatusCore:       GetStatusCoreComponents(),
		StorageService:   disabled.NewChainStorer(),
		ProcessingMode:   common.Normal,
		ChainHandler:     &testscommon.ChainHandlerStub{},
	}

	return stateComponentsFactoryArgs
}

// GetProcessComponentsFactoryArgs -
func GetProcessComponentsFactoryArgs(shardCoordinator sharding.Coordinator) processComp.ProcessComponentsFactoryArgs {
	coreComponents := GetCoreComponents()
	networkComponents := GetNetworkComponents()
	dataComponents := GetDataComponents(coreComponents, shardCoordinator)
	cryptoComponents := GetCryptoComponents(coreComponents)
	stateComponents := GetStateComponents(coreComponents, shardCoordinator)
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
	networkComponents := GetNetworkComponents()
	cryptoComponents := GetCryptoComponents(coreComponents)
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

	gasSchedule := arwenConfig.MakeGasMapForTests()
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
		dataComponents,
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
			GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*block.MiniBlock, map[uint32]*indexer.Pool, error) {
				txsPool := make(map[uint32]*indexer.Pool)
				for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
					txsPool[i] = &indexer.Pool{}
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
					MinQuorum:        "50",
					MinPassThreshold: "50",
					MinVetoThreshold: "50",
				},
				FirstWhitelistedAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
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
		Version:     "v1.0.0",
		HistoryRepo: &dblookupext.HistoryRepositoryStub{},
	}
}

// GetStatusComponents -
func GetStatusComponents(
	coreComponents factory.CoreComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
	dataComponents factory.DataComponentsHolder,
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
		},
		EconomicsConfig:      config.EconomicsConfig{},
		ShardCoordinator:     shardCoordinator,
		NodesCoordinator:     nodesCoordinator,
		EpochStartNotifier:   coreComponents.EpochStartNotifierWithConfirm(),
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
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
	networkComponents := GetNetworkComponents()
	dataComponents := GetDataComponents(coreComponents, shardCoordinator)
	cryptoComponents := GetCryptoComponents(coreComponents)
	stateComponents := GetStateComponents(coreComponents, shardCoordinator)
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
		},
		EconomicsConfig:      config.EconomicsConfig{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(2),
		NodesCoordinator:     &shardingMocks.NodesCoordinatorMock{},
		EpochStartNotifier:   &mock.EpochStartNotifierStub{},
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		NetworkComponents:    networkComponents,
		StateComponents:      stateComponents,
		StatusCoreComponents: statusCoreComponents,
		IsInImportMode:       false,
	}, processComponents
}

// GetNetworkComponents -
func GetNetworkComponents() factory.NetworkComponentsHolder {
	networkArgs := GetNetworkFactoryArgs()
	networkComponentsFactory, _ := networkComp.NewNetworkComponentsFactory(networkArgs)
	networkComponents, _ := networkComp.NewManagedNetworkComponents(networkComponentsFactory)

	_ = networkComponents.Create()

	return networkComponents
}

// GetDataComponents -
func GetDataComponents(coreComponents factory.CoreComponentsHolder, shardCoordinator sharding.Coordinator) factory.DataComponentsHolder {
	dataArgs := GetDataArgs(coreComponents, shardCoordinator)
	dataComponentsFactory, _ := dataComp.NewDataComponentsFactory(dataArgs)
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
func GetStateComponents(coreComponents factory.CoreComponentsHolder, shardCoordinator sharding.Coordinator) factory.StateComponentsHolder {
	stateArgs := GetStateFactoryArgs(coreComponents, shardCoordinator)
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
