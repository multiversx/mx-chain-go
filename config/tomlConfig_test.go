package config

import (
	"fmt"
	"strconv"
	"testing"

	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTomlParser(t *testing.T) {
	txBlockBodyStorageSize := 170
	txBlockBodyStorageType := "type1"
	txBlockBodyStorageShards := 5
	txBlockBodyStorageFile := "path1/file1"
	txBlockBodyStorageTypeDB := "type2"

	receiptsStorageSize := 171
	receiptsStorageType := "type3"
	receiptsStorageFile := "path1/file2"
	receiptsStorageTypeDB := "type4"

	scheduledSCRsStorageSize := 174
	scheduledSCRsStorageType := "type7"
	scheduledSCRsStorageFile := "path1/file4"
	scheduledSCRsStorageTypeDB := "type8"

	logsPath := "pathLogger"
	logsStackDepth := 1010

	accountsStorageSize := 172
	accountsStorageType := "type5"
	accountsStorageFile := "path1/file3"
	accountsStorageTypeDB := "type6"

	hasherType := "hashFunc4"
	multiSigHasherType := "hashFunc5"

	consensusType := "bls"

	wasmVMVersions := []WasmVMVersionByEpoch{
		{StartEpoch: 12, Version: "v0.3"},
		{StartEpoch: 88, Version: "v1.2"},
	}

	cfgExpected := Config{
		MiniBlocksStorage: StorageConfig{
			Cache: CacheConfig{
				Capacity: uint32(txBlockBodyStorageSize),
				Type:     txBlockBodyStorageType,
				Shards:   uint32(txBlockBodyStorageShards),
			},
			DB: DBConfig{
				FilePath: txBlockBodyStorageFile,
				Type:     txBlockBodyStorageTypeDB,
			},
		},
		ReceiptsStorage: StorageConfig{
			Cache: CacheConfig{
				Capacity: uint32(receiptsStorageSize),
				Type:     receiptsStorageType,
			},
			DB: DBConfig{
				FilePath: receiptsStorageFile,
				Type:     receiptsStorageTypeDB,
			},
		},
		ScheduledSCRsStorage: StorageConfig{
			Cache: CacheConfig{
				Capacity: uint32(scheduledSCRsStorageSize),
				Type:     scheduledSCRsStorageType,
			},
			DB: DBConfig{
				FilePath: scheduledSCRsStorageFile,
				Type:     scheduledSCRsStorageTypeDB,
			},
		},
		AccountsTrieStorage: StorageConfig{
			Cache: CacheConfig{
				Capacity: uint32(accountsStorageSize),
				Type:     accountsStorageType,
			},
			DB: DBConfig{
				FilePath: accountsStorageFile,
				Type:     accountsStorageTypeDB,
			},
		},
		Hasher: TypeConfig{
			Type: hasherType,
		},
		MultisigHasher: TypeConfig{
			Type: multiSigHasherType,
		},
		Consensus: ConsensusConfig{
			Type: consensusType,
		},
		VirtualMachine: VirtualMachineServicesConfig{
			Execution: VirtualMachineConfig{
				WasmVMVersions:                      wasmVMVersions,
				TimeOutForSCExecutionInMilliseconds: 10000,
				WasmerSIGSEGVPassthrough:            true,
			},
			Querying: QueryVirtualMachineConfig{
				NumConcurrentVMs:     16,
				VirtualMachineConfig: VirtualMachineConfig{WasmVMVersions: wasmVMVersions},
			},
			GasConfig: VirtualMachineGasConfig{
				ShardMaxGasPerVmQuery: 1_500_000_000,
				MetaMaxGasPerVmQuery:  0,
			},
		},
		Debug: DebugConfig{
			InterceptorResolver: InterceptorResolverDebugConfig{
				Enabled:                    true,
				EnablePrint:                true,
				CacheSize:                  10000,
				IntervalAutoPrintInSeconds: 20,
				NumRequestsThreshold:       9,
				NumResolveFailureThreshold: 3,
				DebugLineExpiration:        10,
			},
			Antiflood: AntifloodDebugConfig{
				Enabled:                    true,
				CacheSize:                  10000,
				IntervalAutoPrintInSeconds: 20,
			},
			ShuffleOut: ShuffleOutDebugConfig{
				CallGCWhenShuffleOut:    true,
				ExtraPrintsOnShuffleOut: true,
				DoProfileOnShuffleOut:   true,
			},
		},
	}
	testString := `
[MiniBlocksStorage]
    [MiniBlocksStorage.Cache]
        Capacity = ` + strconv.Itoa(txBlockBodyStorageSize) + `
        Type = "` + txBlockBodyStorageType + `"
        Shards = ` + strconv.Itoa(txBlockBodyStorageShards) + `
    [MiniBlocksStorage.DB]
        FilePath = "` + txBlockBodyStorageFile + `"
        Type = "` + txBlockBodyStorageTypeDB + `"

[ReceiptsStorage]
    [ReceiptsStorage.Cache]
        Capacity = ` + strconv.Itoa(receiptsStorageSize) + `
        Type = "` + receiptsStorageType + `"
    [ReceiptsStorage.DB]
        FilePath = "` + receiptsStorageFile + `"
        Type = "` + receiptsStorageTypeDB + `"

[ScheduledSCRsStorage]
    [ScheduledSCRsStorage.Cache]
        Capacity = ` + strconv.Itoa(scheduledSCRsStorageSize) + `
        Type = "` + scheduledSCRsStorageType + `"
    [ScheduledSCRsStorage.DB]
        FilePath = "` + scheduledSCRsStorageFile + `"
        Type = "` + scheduledSCRsStorageTypeDB + `"

[Logger]
    Path = "` + logsPath + `"
    StackTraceDepth = ` + strconv.Itoa(logsStackDepth) + `

[AccountsTrieStorage]
    [AccountsTrieStorage.Cache]
        Capacity = ` + strconv.Itoa(accountsStorageSize) + `
        Type = "` + accountsStorageType + `"
    [AccountsTrieStorage.DB]
        FilePath = "` + accountsStorageFile + `"
        Type = "` + accountsStorageTypeDB + `"

[Hasher]
    Type = "` + hasherType + `"

[MultisigHasher]
    Type = "` + multiSigHasherType + `"

[Consensus]
    Type = "` + consensusType + `"

[VirtualMachine]
    [VirtualMachine.Execution]
        TimeOutForSCExecutionInMilliseconds = 10000 # 10 seconds = 10000 milliseconds
        WasmerSIGSEGVPassthrough            = true
        WasmVMVersions = [
            { StartEpoch = 12, Version = "v0.3" },
            { StartEpoch = 88, Version = "v1.2" },
        ]

    [VirtualMachine.Querying]
        NumConcurrentVMs = 16
        WasmVMVersions = [
            { StartEpoch = 12, Version = "v0.3" },
            { StartEpoch = 88, Version = "v1.2" },
        ]

    [VirtualMachine.GasConfig]
        ShardMaxGasPerVmQuery = 1500000000
        MetaMaxGasPerVmQuery = 0

[Debug]
    [Debug.InterceptorResolver]
        Enabled = true
        CacheSize = 10000
        EnablePrint = true
        IntervalAutoPrintInSeconds = 20
        NumRequestsThreshold = 9
        NumResolveFailureThreshold = 3
        DebugLineExpiration = 10
    [Debug.Antiflood]
        Enabled = true
        CacheSize = 10000
        IntervalAutoPrintInSeconds = 20
    [Debug.ShuffleOut]
        CallGCWhenShuffleOut = true
        ExtraPrintsOnShuffleOut = true
        DoProfileOnShuffleOut = true
`
	cfg := Config{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	require.Nil(t, err)
	require.Equal(t, cfgExpected, cfg)
}

func TestTomlEconomicsParser(t *testing.T) {
	protocolSustainabilityPercentage := 0.1
	leaderPercentage1 := 0.1
	leaderPercentage2 := 0.2
	epoch0 := uint32(0)
	epoch1 := uint32(1)
	developerPercentage := 0.3
	maxGasLimitPerBlock := "18446744073709551615"
	minGasPrice := "18446744073709551615"
	minGasLimit := "18446744073709551615"
	extraGasLimitGuardedTx := "50000"
	maxGasPriceSetGuardian := "1234567"
	protocolSustainabilityAddress := "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp"
	denomination := 18

	cfgEconomicsExpected := EconomicsConfig{
		GlobalSettings: GlobalSettings{
			Denomination: denomination,
		},
		RewardsSettings: RewardsSettings{
			RewardsConfigByEpoch: []EpochRewardSettings{
				{
					EpochEnable:                      epoch0,
					LeaderPercentage:                 leaderPercentage1,
					ProtocolSustainabilityPercentage: protocolSustainabilityPercentage,
					ProtocolSustainabilityAddress:    protocolSustainabilityAddress,
					DeveloperPercentage:              developerPercentage,
				},
				{
					EpochEnable:                      epoch1,
					LeaderPercentage:                 leaderPercentage2,
					ProtocolSustainabilityPercentage: protocolSustainabilityPercentage,
					ProtocolSustainabilityAddress:    protocolSustainabilityAddress,
					DeveloperPercentage:              developerPercentage,
				},
			},
		},
		FeeSettings: FeeSettings{
			GasLimitSettings: []GasLimitSetting{
				{
					MaxGasLimitPerBlock:    maxGasLimitPerBlock,
					MinGasLimit:            minGasLimit,
					ExtraGasLimitGuardedTx: extraGasLimitGuardedTx,
				},
			},
			MinGasPrice:            minGasPrice,
			MaxGasPriceSetGuardian: maxGasPriceSetGuardian,
		},
	}

	testString := `
[GlobalSettings]
    Denomination = ` + fmt.Sprintf("%d", denomination) + `
[RewardsSettings]
    [[RewardsSettings.RewardsConfigByEpoch]]
    EpochEnable = ` + fmt.Sprintf("%d", epoch0) + `
    LeaderPercentage = ` + fmt.Sprintf("%.6f", leaderPercentage1) + `
    DeveloperPercentage = ` + fmt.Sprintf("%.6f", developerPercentage) + `
    ProtocolSustainabilityPercentage = ` + fmt.Sprintf("%.6f", protocolSustainabilityPercentage) + ` #fraction of value 0.1 - 10%
    ProtocolSustainabilityAddress = "` + protocolSustainabilityAddress + `"

    [[RewardsSettings.RewardsConfigByEpoch]]
    EpochEnable = ` + fmt.Sprintf("%d", epoch1) + `
    LeaderPercentage = ` + fmt.Sprintf("%.6f", leaderPercentage2) + `
    DeveloperPercentage = ` + fmt.Sprintf("%.6f", developerPercentage) + `
    ProtocolSustainabilityPercentage = ` + fmt.Sprintf("%.6f", protocolSustainabilityPercentage) + ` #fraction of value 0.1 - 10%
    ProtocolSustainabilityAddress = "` + protocolSustainabilityAddress + `"

[FeeSettings]
    GasLimitSettings = [{EnableEpoch = 0, MaxGasLimitPerBlock = "` + maxGasLimitPerBlock + `", MaxGasLimitPerMiniBlock = "", MaxGasLimitPerMetaBlock = "", MaxGasLimitPerMetaMiniBlock = "", MaxGasLimitPerTx = "", MinGasLimit = "` + minGasLimit + `", ExtraGasLimitGuardedTx = "` + extraGasLimitGuardedTx + `"}] 
    MinGasPrice = "` + minGasPrice + `"
	MaxGasPriceSetGuardian = "` + maxGasPriceSetGuardian + `"
`
	cfg := EconomicsConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, cfgEconomicsExpected, cfg)
}

func TestTomlPreferencesParser(t *testing.T) {
	nodeDisplayName := "test-name"
	destinationShardAsObs := "3"
	identity := "test-identity"
	redundancyLevel := int64(0)
	prefPubKey0 := "preferred pub key 0"
	prefPubKey1 := "preferred pub key 1"

	cfgPreferencesExpected := Preferences{
		Preferences: PreferencesConfig{
			NodeDisplayName:            nodeDisplayName,
			DestinationShardAsObserver: destinationShardAsObs,
			Identity:                   identity,
			RedundancyLevel:            redundancyLevel,
			PreferredConnections:       []string{prefPubKey0, prefPubKey1},
		},
	}

	testString := `
[Preferences]
    NodeDisplayName = "` + nodeDisplayName + `"
    DestinationShardAsObserver = "` + destinationShardAsObs + `"
    Identity = "` + identity + `"
    RedundancyLevel = ` + fmt.Sprintf("%d", redundancyLevel) + `
    PreferredConnections = [
        "` + prefPubKey0 + `",
        "` + prefPubKey1 + `"
    ]
`
	cfg := Preferences{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, cfgPreferencesExpected, cfg)
}

func TestTomlExternalParser(t *testing.T) {
	indexerURL := "url"
	elasticUsername := "user"
	elasticPassword := "pass"

	cfgExternalExpected := ExternalConfig{
		ElasticSearchConnector: ElasticSearchConfig{
			Enabled:  true,
			URL:      indexerURL,
			Username: elasticUsername,
			Password: elasticPassword,
		},
	}

	testString := `
[ElasticSearchConnector]
    Enabled = true
    URL = "` + indexerURL + `"
    Username = "` + elasticUsername + `"
    Password = "` + elasticPassword + `"`

	cfg := ExternalConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, cfgExternalExpected, cfg)
}

func TestAPIRoutesToml(t *testing.T) {
	package0 := "testPackage0"
	route0 := "testRoute0"
	route1 := "testRoute1"

	package1 := "testPackage1"
	route2 := "testRoute2"

	loggingThreshold := 10

	expectedCfg := ApiRoutesConfig{
		Logging: ApiLoggingConfig{
			LoggingEnabled:          true,
			ThresholdInMicroSeconds: loggingThreshold,
		},
		APIPackages: map[string]APIPackageConfig{
			package0: {
				Routes: []RouteConfig{
					{Name: route0, Open: true},
					{Name: route1, Open: true},
				},
			},
			package1: {
				Routes: []RouteConfig{
					{Name: route2, Open: false},
				},
			},
		},
	}

	testString := `
[Logging]
    LoggingEnabled = true
    ThresholdInMicroSeconds = 10

     # API routes configuration
[APIPackages]

[APIPackages.` + package0 + `]
    Routes = [
        # test comment
        { Name = "` + route0 + `", Open = true },

        # test comment
        { Name = "` + route1 + `", Open = true },
    ]

[APIPackages.` + package1 + `]
    Routes = [
         # test comment
        { Name = "` + route2 + `", Open = false }
    ]
 `

	cfg := ApiRoutesConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

func TestP2pConfig(t *testing.T) {
	initialPeersList := "/ip4/127.0.0.1/tcp/9999/p2p/16Uiu2HAkw5SNNtSvH1zJiQ6Gc3WoGNSxiyNueRKe6fuAuh57G3Bk"
	protocolID := "test protocol id"
	shardingType := "ListSharder"
	port := "37373-38383"

	testString := `
#P2P config file
[Node]
    Port = "` + port + `"
    ThresholdMinConnectedPeers = 0

[KadDhtPeerDiscovery]
    Enabled = false
    Type = ""
    RefreshIntervalInSec = 0
    ProtocolID = "` + protocolID + `"
    InitialPeerList = ["` + initialPeersList + `"]

    #kademlia's routing table bucket size
    BucketSize = 0

    #RoutingTableRefreshIntervalInSec defines how many seconds should pass between 2 kad routing table auto refresh calls
    RoutingTableRefreshIntervalInSec = 0

[Sharding]
    # The targeted number of peer connections
    TargetPeerCount = 0
    MaxIntraShardValidators = 0
    MaxCrossShardValidators = 0
    MaxIntraShardObservers = 0
    MaxCrossShardObservers = 0
    MaxSeeders = 0
    Type = "` + shardingType + `"
    [AdditionalConnections]
        MaxFullHistoryObservers = 0`

	expectedCfg := p2pConfig.P2PConfig{
		Node: p2pConfig.NodeConfig{
			Port: port,
		},
		KadDhtPeerDiscovery: p2pConfig.KadDhtPeerDiscoveryConfig{
			ProtocolID:      protocolID,
			InitialPeerList: []string{initialPeersList},
		},
		Sharding: p2pConfig.ShardingConfig{
			Type: shardingType,
		},
	}
	cfg := p2pConfig.P2PConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

func TestEnableEpochConfig(t *testing.T) {
	testString := `
[EnableEpochs]
    # SCDeployEnableEpoch represents the epoch when the deployment of smart contracts will be enabled
    SCDeployEnableEpoch = 1

    # BuiltInFunctionsEnableEpoch represents the epoch when the built in functions will be enabled
    BuiltInFunctionsEnableEpoch = 2

    # RelayedTransactionsEnableEpoch represents the epoch when the relayed transactions will be enabled
    RelayedTransactionsEnableEpoch = 3

    # PenalizedTooMuchGasEnableEpoch represents the epoch when the penalization for using too much gas will be enabled
    PenalizedTooMuchGasEnableEpoch = 4

    # SwitchJailWaitingEnableEpoch represents the epoch when the system smart contract processing at end of epoch is enabled
    SwitchJailWaitingEnableEpoch = 5

    # BelowSignedThresholdEnableEpoch represents the epoch when the change for computing rating for validators below signed rating is enabled
    BelowSignedThresholdEnableEpoch = 6

    # SwitchHysteresisForMinNodesEnableEpoch represents the epoch when the system smart contract changes its config to consider
    # also (minimum) hysteresis nodes for the minimum number of nodes
    SwitchHysteresisForMinNodesEnableEpoch = 7

    # TransactionSignedWithTxHashEnableEpoch represents the epoch when the node will also accept transactions that are
    # signed with the hash of transaction
    TransactionSignedWithTxHashEnableEpoch = 8

    # MetaProtectionEnableEpoch represents the epoch when the transactions to the metachain are checked to have enough gas
    MetaProtectionEnableEpoch = 9

    # AheadOfTimeGasUsageEnableEpoch represents the epoch when the cost of smart contract prepare changes from compiler per byte to ahead of time prepare per byte
    AheadOfTimeGasUsageEnableEpoch = 10

    # GasPriceModifierEnableEpoch represents the epoch when the gas price modifier in fee computation is enabled
    GasPriceModifierEnableEpoch = 11

    # RepairCallbackEnableEpoch represents the epoch when the callback repair is activated for scrs
    RepairCallbackEnableEpoch = 12

    # BlockGasAndFeesReCheckEnableEpoch represents the epoch when gas and fees used in each created or processed block are re-checked
    BlockGasAndFeesReCheckEnableEpoch = 13

    # BalanceWaitingListsEnableEpoch represents the epoch when the shard waiting lists are balanced at the start of an epoch
    BalanceWaitingListsEnableEpoch = 14

    # ReturnDataToLastTransferEnableEpoch represents the epoch when returned data is added to last output transfer for callbacks
    ReturnDataToLastTransferEnableEpoch = 15

    # SenderInOutTransferEnableEpoch represents the epoch when the feature of having different senders in output transfer is enabled
    SenderInOutTransferEnableEpoch = 16

    # StakeEnableEpoch represents the epoch when staking is enabled
    StakeEnableEpoch = 17

    # StakingV2EnableEpoch represents the epoch when staking v2 is enabled
    StakingV2EnableEpoch = 18

    DoubleKeyProtectionEnableEpoch = 19

    # ESDTEnableEpoch represents the epoch when ESDT is enabled
    ESDTEnableEpoch = 20

    # GovernanceEnableEpoch represents the epoch when governance is enabled
    GovernanceEnableEpoch = 21

    # DelegationManagerEnableEpoch represents the epoch when the delegation manager is enabled
    # epoch should not be 0
    DelegationManagerEnableEpoch = 22

    # DelegationSmartContractEnableEpoch represents the epoch when delegation smart contract is enabled
    # epoch should not be 0
    DelegationSmartContractEnableEpoch = 23

    # CorrectLastUnjailedEnableEpoch represents the epoch when the fix regaring the last unjailed node should apply
    CorrectLastUnjailedEnableEpoch = 24

    # RelayedTransactionsV2EnableEpoch represents the epoch when the relayed transactions V2 will be enabled
    RelayedTransactionsV2EnableEpoch = 25

    # UnbondTokensV2EnableEpoch represents the epoch when the new implementation of the unbond tokens function is available
    UnbondTokensV2EnableEpoch = 26

    # SaveJailedAlwaysEnableEpoch represents the epoch when saving jailed status at end of epoch will happen in all cases
    SaveJailedAlwaysEnableEpoch = 27

    # ReDelegateBelowMinCheckEnableEpoch represents the epoch when the check for the re-delegated value will be enabled
    ReDelegateBelowMinCheckEnableEpoch = 28

    # ValidatorToDelegationEnableEpoch represents the epoch when the validator-to-delegation feature will be enabled
    ValidatorToDelegationEnableEpoch = 29

    # WaitingListFixEnableEpoch represents the epoch when the 6 epoch waiting list fix is enabled
    WaitingListFixEnableEpoch = 30

    # IncrementSCRNonceInMultiTransferEnableEpoch represents the epoch when the fix for preventing the generation of the same SCRs
    # is enabled. The fix is done by adding an extra increment.
    IncrementSCRNonceInMultiTransferEnableEpoch = 31

    # ESDTMultiTransferEnableEpoch represents the epoch when esdt multitransfer built in function is enabled
    ESDTMultiTransferEnableEpoch = 32

    # GlobalMintBurnDisableEpoch represents the epoch when the global mint and burn functions are disabled
    GlobalMintBurnDisableEpoch = 33

    # ESDTTransferRoleEnableEpoch represents the epoch when esdt transfer role set is enabled
    ESDTTransferRoleEnableEpoch = 34

    # BuiltInFunctionOnMetaEnableEpoch represents the epoch when built in function processing on metachain is enabled
    BuiltInFunctionOnMetaEnableEpoch = 35

    # ComputeRewardCheckpointEnableEpoch represents the epoch when compute rewards checkpoint epoch is enabled
    ComputeRewardCheckpointEnableEpoch = 36

    # SCRSizeInvariantCheckEnableEpoch represents the epoch when the scr size invariant check is enabled
    SCRSizeInvariantCheckEnableEpoch = 37

    # BackwardCompSaveKeyValueEnableEpoch represents the epoch when backward compatibility save key value is enabled
    BackwardCompSaveKeyValueEnableEpoch = 38

    # ESDTNFTCreateOnMultiShardEnableEpoch represents the epoch when esdt nft creation on multiple shards is enabled
    ESDTNFTCreateOnMultiShardEnableEpoch = 39

    # MetaESDTSetEnableEpoch represents the epoch when the backward compatibility for save key value error is enabled
    MetaESDTSetEnableEpoch = 40

    # AddTokensToDelegationEnableEpoch represents the epoch when adding tokens to delegation is enabled for whitelisted address
    AddTokensToDelegationEnableEpoch = 41

    # MultiESDTTransferFixOnCallBackOnEnableEpoch represents the epoch when multi esdt transfer on callback fix is enabled
    MultiESDTTransferFixOnCallBackOnEnableEpoch = 42

    # OptimizeGasUsedInCrossMiniBlocksEnableEpoch represents the epoch when gas used in cross shard mini blocks will be optimized
    OptimizeGasUsedInCrossMiniBlocksEnableEpoch = 43

    # FixOOGReturnCodeEnableEpoch represents the epoch when the backward compatibility returning out of gas error is enabled
    FixOOGReturnCodeEnableEpoch = 44

    # RemoveNonUpdatedStorageEnableEpoch represents the epoch when the backward compatibility for removing non updated storage is enabled
    RemoveNonUpdatedStorageEnableEpoch = 45

    # OptimizeNFTStoreEnableEpoch represents the epoch when optimizations on NFT metadata store and send are enabled
    OptimizeNFTStoreEnableEpoch = 46

    # CreateNFTThroughExecByCallerEnableEpoch represents the epoch when nft creation through execution on destination by caller is enabled
    CreateNFTThroughExecByCallerEnableEpoch = 47

    # IsPayableBySCEnableEpoch represents the epoch when a new flag isPayable by SC is enabled
    IsPayableBySCEnableEpoch = 48

    # CleanUpInformativeSCRsEnableEpoch represents the epoch when the scrs which contain only information are cleaned from miniblocks and logs are created from it
    CleanUpInformativeSCRsEnableEpoch = 49

    # StorageAPICostOptimizationEnableEpoch represents the epoch when new storage helper functions are enabled and cost is reduced in Wasm VM
    StorageAPICostOptimizationEnableEpoch = 50

    # TransformToMultiShardCreateEnableEpoch represents the epoch when the new function on esdt system sc is enabled to transfer create role into multishard
    TransformToMultiShardCreateEnableEpoch = 51

    # ESDTRegisterAndSetAllRolesEnableEpoch represents the epoch when new function to register tickerID and set all roles is enabled
    ESDTRegisterAndSetAllRolesEnableEpoch = 52

    # FailExecutionOnEveryAPIErrorEnableEpoch represent the epoch when new protection in VM is enabled to fail all wrong API calls
    FailExecutionOnEveryAPIErrorEnableEpoch = 53

    # ManagedCryptoAPIsEnableEpoch represents the epoch when the new managed crypto APIs are enabled
    ManagedCryptoAPIsEnableEpoch = 54

    # ESDTMetadataContinuousCleanupEnableEpoch represents the epoch when esdt metadata is automatically deleted according to inshard liquidity
    ESDTMetadataContinuousCleanupEnableEpoch = 55

    # FixAsyncCallBackArgsListEnableEpoch represents the epoch when the async callback arguments lists fix will be enabled
    FixAsyncCallBackArgsListEnableEpoch = 56

    # FixOldTokenLiquidityEnableEpoch represents the epoch when the fix for old token liquidity is enabled
    FixOldTokenLiquidityEnableEpoch = 57

    # SetSenderInEeiOutputTransferEnableEpoch represents the epoch when setting the sender in eei output transfers will be enabled
    SetSenderInEeiOutputTransferEnableEpoch = 58

    # MaxBlockchainHookCountersEnableEpoch represents the epoch when the max blockchainhook counters are enabled
    MaxBlockchainHookCountersEnableEpoch = 59

    # WipeSingleNFTLiquidityDecreaseEnableEpoch represents the epoch when the system account liquidity is decreased for wipeSingleNFT as well
    WipeSingleNFTLiquidityDecreaseEnableEpoch = 60

    # AlwaysSaveTokenMetaDataEnableEpoch represents the epoch when the token metadata is always saved
    AlwaysSaveTokenMetaDataEnableEpoch = 61

    # RuntimeCodeSizeFixEnableEpoch represents the epoch when the code size fix in the VM is enabled
    RuntimeCodeSizeFixEnableEpoch = 62

    # RuntimeMemStoreLimitEnableEpoch represents the epoch when the condition for Runtime MemStore is enabled
    RuntimeMemStoreLimitEnableEpoch = 63

	# SetGuardianEnableEpoch represents the epoch when guard account feature is enabled
	SetGuardianEnableEpoch = 64

    # RelayedNonceFixEnableEpoch represents the epoch when the nonce fix for relayed txs is enabled
    RelayedNonceFixEnableEpoch = 65

	# DynamicGasCostForDataTrieStorageLoadEnableEpoch represents the epoch when dynamic gas cost for data trie storage load will be enabled
    DynamicGasCostForDataTrieStorageLoadEnableEpoch = 66

    # MaxNodesChangeEnableEpoch holds configuration for changing the maximum number of nodes and the enabling epoch
    MaxNodesChangeEnableEpoch = [
        { EpochEnable = 44, MaxNumNodes = 2169, NodesToShufflePerShard = 80 },
        { EpochEnable = 45, MaxNumNodes = 3200, NodesToShufflePerShard = 80 }
    ]

    BLSMultiSignerEnableEpoch = [
        {EnableEpoch = 0, Type = "no-KOSK"},
        {EnableEpoch = 3, Type = "KOSK"}
    ]

[GasSchedule]
    GasScheduleByEpochs = [
        { StartEpoch = 46, FileName = "gasScheduleV1.toml" },
        { StartEpoch = 47, FileName = "gasScheduleV3.toml" },
    ]
`

	expectedCfg := EpochConfig{
		EnableEpochs: EnableEpochs{
			SCDeployEnableEpoch:                    1,
			BuiltInFunctionsEnableEpoch:            2,
			RelayedTransactionsEnableEpoch:         3,
			PenalizedTooMuchGasEnableEpoch:         4,
			SwitchJailWaitingEnableEpoch:           5,
			SwitchHysteresisForMinNodesEnableEpoch: 7,
			BelowSignedThresholdEnableEpoch:        6,
			TransactionSignedWithTxHashEnableEpoch: 8,
			MetaProtectionEnableEpoch:              9,
			AheadOfTimeGasUsageEnableEpoch:         10,
			GasPriceModifierEnableEpoch:            11,
			RepairCallbackEnableEpoch:              12,
			MaxNodesChangeEnableEpoch: []MaxNodesChangeConfig{
				{
					EpochEnable:            44,
					MaxNumNodes:            2169,
					NodesToShufflePerShard: 80,
				},
				{
					EpochEnable:            45,
					MaxNumNodes:            3200,
					NodesToShufflePerShard: 80,
				},
			},
			BlockGasAndFeesReCheckEnableEpoch:               13,
			StakingV2EnableEpoch:                            18,
			StakeEnableEpoch:                                17,
			DoubleKeyProtectionEnableEpoch:                  19,
			ESDTEnableEpoch:                                 20,
			GovernanceEnableEpoch:                           21,
			DelegationManagerEnableEpoch:                    22,
			DelegationSmartContractEnableEpoch:              23,
			CorrectLastUnjailedEnableEpoch:                  24,
			BalanceWaitingListsEnableEpoch:                  14,
			ReturnDataToLastTransferEnableEpoch:             15,
			SenderInOutTransferEnableEpoch:                  16,
			RelayedTransactionsV2EnableEpoch:                25,
			UnbondTokensV2EnableEpoch:                       26,
			SaveJailedAlwaysEnableEpoch:                     27,
			ValidatorToDelegationEnableEpoch:                29,
			ReDelegateBelowMinCheckEnableEpoch:              28,
			WaitingListFixEnableEpoch:                       30,
			IncrementSCRNonceInMultiTransferEnableEpoch:     31,
			ESDTMultiTransferEnableEpoch:                    32,
			GlobalMintBurnDisableEpoch:                      33,
			ESDTTransferRoleEnableEpoch:                     34,
			BuiltInFunctionOnMetaEnableEpoch:                35,
			ComputeRewardCheckpointEnableEpoch:              36,
			SCRSizeInvariantCheckEnableEpoch:                37,
			BackwardCompSaveKeyValueEnableEpoch:             38,
			ESDTNFTCreateOnMultiShardEnableEpoch:            39,
			MetaESDTSetEnableEpoch:                          40,
			AddTokensToDelegationEnableEpoch:                41,
			MultiESDTTransferFixOnCallBackOnEnableEpoch:     42,
			OptimizeGasUsedInCrossMiniBlocksEnableEpoch:     43,
			FixOOGReturnCodeEnableEpoch:                     44,
			RemoveNonUpdatedStorageEnableEpoch:              45,
			OptimizeNFTStoreEnableEpoch:                     46,
			CreateNFTThroughExecByCallerEnableEpoch:         47,
			IsPayableBySCEnableEpoch:                        48,
			CleanUpInformativeSCRsEnableEpoch:               49,
			StorageAPICostOptimizationEnableEpoch:           50,
			TransformToMultiShardCreateEnableEpoch:          51,
			ESDTRegisterAndSetAllRolesEnableEpoch:           52,
			FailExecutionOnEveryAPIErrorEnableEpoch:         53,
			ManagedCryptoAPIsEnableEpoch:                    54,
			ESDTMetadataContinuousCleanupEnableEpoch:        55,
			FixAsyncCallBackArgsListEnableEpoch:             56,
			FixOldTokenLiquidityEnableEpoch:                 57,
			SetSenderInEeiOutputTransferEnableEpoch:         58,
			MaxBlockchainHookCountersEnableEpoch:            59,
			WipeSingleNFTLiquidityDecreaseEnableEpoch:       60,
			AlwaysSaveTokenMetaDataEnableEpoch:              61,
			RuntimeCodeSizeFixEnableEpoch:                   62,
			RuntimeMemStoreLimitEnableEpoch:                 63,
			SetGuardianEnableEpoch:                          64,
			RelayedNonceFixEnableEpoch:                      65,
			DynamicGasCostForDataTrieStorageLoadEnableEpoch: 66,
			BLSMultiSignerEnableEpoch: []MultiSignerConfig{
				{
					EnableEpoch: 0,
					Type:        "no-KOSK",
				},
				{
					EnableEpoch: 3,
					Type:        "KOSK",
				},
			},
		},

		GasSchedule: GasScheduleConfig{
			GasScheduleByEpochs: []GasScheduleByEpochs{
				{
					StartEpoch: 46,
					FileName:   "gasScheduleV1.toml",
				},
				{
					StartEpoch: 47,
					FileName:   "gasScheduleV3.toml",
				},
			},
		},
	}
	cfg := EpochConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	assert.Nil(t, err)
	assert.Equal(t, expectedCfg, cfg)
}
