package factory_test

import (
	"strconv"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
)

var minTxGasPrice = uint64(10)
var minTxGasLimit = uint64(1000)
var maxGasLimitPerBlock = uint64(3000000)

// TODO: write unit tests

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
		GasSchedule:            gasSchedule,
		Rounder:                &mock.RounderMock{},
		ShardCoordinator:       mock.NewMultiShardsCoordinatorMock(2),
		NodesCoordinator:       &mock.NodesCoordinatorMock{},
		Data:                   dataComponents,
		CoreData:               coreComponets,
		Crypto:                 cryptoComponents,
		State:                  stateComponents,
		Network:                networkComponents,
		CoreServiceContainer:   &testscommon.ServiceContainerMock{},
		RequestedItemsHandler:  &testscommon.RequestedItemsHandlerStub{},
		WhiteListHandler:       &testscommon.WhiteListHandlerStub{},
		WhiteListerVerifiedTxs: &testscommon.WhiteListHandlerStub{},
		EpochStartNotifier:     &mock.EpochStartNotifierStub{},
		EpochStart: &config.EpochStartConfig{
			MinRoundsBetweenEpochs:            20,
			RoundsPerEpoch:                    20,
			ShuffledOutRestartThreshold:       0,
			ShuffleBetweenShards:              false,
			MinNumConnectedPeersToStart:       2,
			MinNumOfPeersToConsiderBlockValid: 2,
		},
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
		},
		Version: "v1.0.0",
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
				TotalSupply:      "2000000000000000000000",
				MinimumInflation: 0,
				MaximumInflation: 0.05,
			},
			RewardsSettings: config.RewardsSettings{
				LeaderPercentage:    0.1,
				DeveloperPercentage: 0.1,
				CommunityAddress:    "test address",
			},
			FeeSettings: config.FeeSettings{
				MaxGasLimitPerBlock:     maxGasLimitPerBlock,
				MaxGasLimitPerMetaBlock: maxGasLimitPerBlock,
				MinGasPrice:             minGasPrice,
				MinGasLimit:             minGasLimit,
				GasPerDataByte:          "1",
				DataLimitForBaseCalc:    "10000",
			},
			ValidatorSettings: config.ValidatorSettings{
				GenesisNodePrice:         "500000000",
				UnBondPeriod:             "5",
				TotalSupply:              "200000000000",
				MinStepValue:             "100000",
				AuctionEnableNonce:       "100000",
				StakeEnableNonce:         "0",
				NumRoundsWithoutBleed:    "1000",
				MaximumPercentageToBleed: "0.5",
				BleedPercentagePerRound:  "0.00001",
				UnJailValue:              "1000",
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

	return gasMap
}
