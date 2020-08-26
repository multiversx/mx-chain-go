package factory_test

import (
	"strconv"
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

var minTxGasPrice = uint64(10)
var minTxGasLimit = uint64(1000)
var maxGasLimitPerBlock = uint64(3000000)

// TODO: write unit tests

// ------------ Test ManagedCoreComponents --------------------
func TestManagedProcessComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	processArgs := getProcessComponentsArgs()
	_ = processArgs.CoreData.SetInternalMarshalizer(nil)
	managedCoreComponents, err := factory.NewManagedProcessComponents(processArgs)
	require.NoError(t, err)
	err = managedCoreComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedCoreComponents.NodesCoordinator())
}

func TestManagedProcessComponents_Create_ShouldWork(t *testing.T) {
	processArgs := getProcessComponentsArgs()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(0)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return 0
	}
	shardCoordinator.CurrentShard = core.MetachainShardId
	processArgs.ShardCoordinator = shardCoordinator
	managedProcessComponents, err := factory.NewManagedProcessComponents(processArgs)
	require.NoError(t, err)
	require.Nil(t, managedProcessComponents.NodesCoordinator())
	require.Nil(t, managedProcessComponents.InterceptorsContainer())
	require.Nil(t, managedProcessComponents.ResolversFinder())
	require.Nil(t, managedProcessComponents.Rounder())
	require.Nil(t, managedProcessComponents.ForkDetector())
	require.Nil(t, managedProcessComponents.BlockProcessor())
	require.Nil(t, managedProcessComponents.EpochStartTrigger())
	require.Nil(t, managedProcessComponents.EpochStartNotifier())
	require.Nil(t, managedProcessComponents.BlackListHandler())
	require.Nil(t, managedProcessComponents.BootStorer())
	require.Nil(t, managedProcessComponents.HeaderSigVerifier())
	require.Nil(t, managedProcessComponents.ValidatorsStatistics())
	require.Nil(t, managedProcessComponents.ValidatorsProvider())
	require.Nil(t, managedProcessComponents.BlockTracker())
	require.Nil(t, managedProcessComponents.PendingMiniBlocksHandler())
	require.Nil(t, managedProcessComponents.RequestHandler())
	require.Nil(t, managedProcessComponents.TxLogsProcessor())
	require.Nil(t, managedProcessComponents.HeaderConstructionValidator())
	require.Nil(t, managedProcessComponents.HeaderIntegrityVerifier())

	err = managedProcessComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedProcessComponents.NodesCoordinator())
	require.NotNil(t, managedProcessComponents.InterceptorsContainer())
	require.NotNil(t, managedProcessComponents.ResolversFinder())
	require.NotNil(t, managedProcessComponents.Rounder())
	require.NotNil(t, managedProcessComponents.ForkDetector())
	require.NotNil(t, managedProcessComponents.BlockProcessor())
	require.NotNil(t, managedProcessComponents.EpochStartTrigger())
	require.NotNil(t, managedProcessComponents.EpochStartNotifier())
	require.NotNil(t, managedProcessComponents.BlackListHandler())
	require.NotNil(t, managedProcessComponents.BootStorer())
	require.NotNil(t, managedProcessComponents.HeaderSigVerifier())
	require.NotNil(t, managedProcessComponents.ValidatorsStatistics())
	require.NotNil(t, managedProcessComponents.ValidatorsProvider())
	require.NotNil(t, managedProcessComponents.BlockTracker())
	require.NotNil(t, managedProcessComponents.PendingMiniBlocksHandler())
	require.NotNil(t, managedProcessComponents.RequestHandler())
	require.NotNil(t, managedProcessComponents.TxLogsProcessor())
	require.NotNil(t, managedProcessComponents.HeaderConstructionValidator())
	require.NotNil(t, managedProcessComponents.HeaderIntegrityVerifier())
}

func TestManagedProcessComponents_Close(t *testing.T) {
	processArgs := getProcessComponentsArgs()
	managedCoreComponents, _ := factory.NewManagedProcessComponents(processArgs)
	err := managedCoreComponents.Create()
	require.NoError(t, err)

	err = managedCoreComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedCoreComponents.NodesCoordinator())
}

// ------------ Test CoreComponents --------------------
func TestProcessComponents_Close_ShouldWork(t *testing.T) {
	t.Parallel()

	processArgs := getProcessComponentsArgs()
	pcf, _ := factory.NewProcessComponentsFactory(processArgs)
	pc, _ := pcf.Create()

	err := pc.Close()
	require.NoError(t, err)
}

func getProcessComponentsArgs() factory.ProcessComponentsFactoryArgs {
	coreComponents := getCoreComponents()
	networkComponents := getNetworkComponents()
	dataComponents := getDataComponents(coreComponents)
	cryptoComponents := getCryptoComponents(coreComponents)
	stateComponents := getStateComponents(coreComponents)
	processArgs := getProcessArgs(
		getCoreArgs(),
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)
	return processArgs
}

func getProcessArgs(
	coreArgs factory.CoreComponentsFactoryArgs,
	coreComponets factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	cryptoComponents factory.CryptoComponentsHolder,
	stateComponents factory.StateComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
) factory.ProcessComponentsFactoryArgs {

	gasSchedule := arwenConfig.MakeGasMapForTests()
	// TODO: check if these could be initialized by MakeGasMapForTests()
	gasSchedule["BuiltInCost"]["SaveUserName"] = 1
	gasSchedule["BuiltInCost"]["SaveKeyValue"] = 1
	gasSchedule["BuiltInCost"]["ESDTTransfer"] = 1
	gasSchedule[core.MetaChainSystemSCsCost] = FillGasMapMetaChainSystemSCsCosts(1)

	epochStartConfig := getEpochStartConfig()

	return factory.ProcessComponentsFactoryArgs{
		CoreFactoryArgs:     &coreArgs,
		AccountsParser:      &mock.AccountsParserStub{},
		SmartContractParser: &mock.SmartContractParserStub{},
		EconomicsData:       CreateEconomicsData(),
		NodesConfig: &sharding.NodesSetup{
			StartTime:                   0,
			RoundDuration:               5,
			ConsensusGroupSize:          3,
			MinNodesPerShard:            3,
			MetaChainConsensusGroupSize: 3,
			MetaChainMinNodes:           3,
			Hysteresis:                  0,
			Adaptivity:                  false,
		},
		GasSchedule:               gasSchedule,
		Rounder:                   &mock.RounderMock{},
		ShardCoordinator:          mock.NewMultiShardsCoordinatorMock(2),
		NodesCoordinator:          &mock.NodesCoordinatorMock{},
		Data:                      dataComponents,
		CoreData:                  coreComponets,
		Crypto:                    cryptoComponents,
		State:                     stateComponents,
		Network:                   networkComponents,
		RequestedItemsHandler:     &testscommon.RequestedItemsHandlerStub{},
		WhiteListHandler:          &testscommon.WhiteListHandlerStub{},
		WhiteListerVerifiedTxs:    &testscommon.WhiteListHandlerStub{},
		EpochStartNotifier:        &mock.EpochStartNotifierStub{},
		EpochStart:                &epochStartConfig,
		Rater:                     &testscommon.RaterMock{},
		RatingsData:               &testscommon.RatingsInfoMock{},
		SizeCheckDelta:            0,
		StateCheckpointModulus:    0,
		MaxComputableRounds:       1000,
		NumConcurrentResolverJobs: 2,
		MinSizeInBytes:            0,
		MaxSizeInBytes:            200,
		MaxRating:                 100,
		ImportStartHandler:        &testscommon.ImportStartHandlerStub{},
		ValidatorPubkeyConverter:  &testscommon.PubkeyConverterMock{},
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "1000",
				OwnerAddress:    "aaaaaa",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				ProposalCost:     "500",
				NumNodes:         100,
				MinQuorum:        50,
				MinPassThreshold: 50,
				MinVetoThreshold: 50,
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "100",
				MinStakeValue:                        "1",
				UnJailValue:                          "1",
				MinStepValue:                         "1",
				UnBondPeriod:                         0,
				AuctionEnableNonce:                   0,
				StakeEnableNonce:                     0,
				NumRoundsWithoutBleed:                0,
				MaximumPercentageToBleed:             0,
				BleedPercentagePerRound:              0,
				MaxNumberOfNodesForStake:             10,
				NodesToSelectInAuction:               100,
				ActivateBLSPubKeyMessageVerification: false,
			},
		},
		Version:                 "v1.0.0",
		Indexer:                 &mock.IndexerMock{},
		TpsBenchmark:            &testscommon.TpsBenchmarkMock{},
		HistoryRepo:             &testscommon.HistoryRepositoryStub{},
		EpochNotifier:           &mock.EpochNotifierStub{},
		HeaderIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
	}
}

// CreateEconomicsData creates a mock EconomicsData object
func CreateEconomicsData() *economics.EconomicsData {
	maxGasLimitPerBlock := strconv.FormatUint(maxGasLimitPerBlock, 10)
	minGasPrice := strconv.FormatUint(minTxGasPrice, 10)
	minGasLimit := strconv.FormatUint(minTxGasLimit, 10)

	economicsData, _ := economics.NewEconomicsData(
		&config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				GenesisTotalSupply: "2000000000000000000000",
				MinimumInflation:   0,
				YearSettings: []*config.YearSetting{
					{
						Year:             0,
						MaximumInflation: 0.01,
					},
				},
			},
			RewardsSettings: config.RewardsSettings{
				LeaderPercentage:              0.1,
				DeveloperPercentage:           0.1,
				ProtocolSustainabilityAddress: "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
			},
			FeeSettings: config.FeeSettings{
				MaxGasLimitPerBlock:     maxGasLimitPerBlock,
				MaxGasLimitPerMetaBlock: maxGasLimitPerBlock,
				MinGasPrice:             minGasPrice,
				MinGasLimit:             minGasLimit,
				GasPerDataByte:          "1",
				DataLimitForBaseCalc:    "10000",
			},
		},
	)
	return economicsData
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

	return gasMap
}
