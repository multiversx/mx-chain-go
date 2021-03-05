package factory_test

import (
	"math/big"
	"testing"
	"time"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	factory2 "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

// ------------ Test TestProcessComponents --------------------
func TestProcessComponents_Close_ShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	processArgs := getProcessComponentsArgs(shardCoordinator)
	pcf, _ := factory.NewProcessComponentsFactory(processArgs)
	pc, _ := pcf.Create()

	err := pc.Close()
	require.NoError(t, err)
}

func getProcessComponentsArgs(shardCoordinator sharding.Coordinator) factory.ProcessComponentsFactoryArgs {
	coreComponents := getCoreComponents()
	networkComponents := getNetworkComponents()
	dataComponents := getDataComponents(coreComponents, shardCoordinator)
	cryptoComponents := getCryptoComponents(coreComponents)
	stateComponents := getStateComponents(coreComponents, shardCoordinator)
	processArgs := getProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)
	return processArgs
}

func getProcessArgs(
	shardCoordinator sharding.Coordinator,
	coreComponents factory.CoreComponentsHolder,
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
	gasSchedule["BuiltInCost"]["ESDTBurn"] = 1
	gasSchedule[core.MetaChainSystemSCsCost] = FillGasMapMetaChainSystemSCsCosts(1)

	epochStartConfig := getEpochStartConfig()

	gasScheduleNotifier := &mock.GasScheduleNotifierMock{
		GasSchedule: gasSchedule,
	}

	return factory.ProcessComponentsFactoryArgs{
		Config: testscommon.GetGeneralConfig(),
		AccountsParser: &mock.AccountsParserStub{
			InitialAccountsCalled: func() []genesis.InitialAccountHandler {
				addrConverter, _ := factory2.NewPubkeyConverter(config.PubkeyConfig{
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
		},
		SmartContractParser: &mock.SmartContractParserStub{},
		EconomicsData:       CreateEconomicsData(),
		GasSchedule:         gasScheduleNotifier,
		RoundHandler: &mock.RoundHandlerMock{
			RoundTimeDuration: time.Second,
		},
		ShardCoordinator:          shardCoordinator,
		NodesCoordinator:          &mock.NodesCoordinatorMock{},
		Data:                      dataComponents,
		CoreData:                  coreComponents,
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
				BaseIssuingCost:    "100",
				MinCreationDeposit: "100",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinStakeAmount: "100",
				MinServiceFee:  0,
				MaxServiceFee:  100,
			},
		},
		Version:                 "v1.0.0",
		Indexer:                 &mock.IndexerMock{},
		TpsBenchmark:            &testscommon.TpsBenchmarkMock{},
		HistoryRepo:             &testscommon.HistoryRepositoryStub{},
		HeaderIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
	}
}

// CreateEconomicsData creates a mock EconomicsData object
func CreateEconomicsData() process.EconomicsDataHandler {
	economicsConfig := createDummyEconomicsConfig()
	args := economics.ArgsNewEconomicsData{
		Economics:                      &economicsConfig,
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  &mock.EpochNotifierStub{},
	}
	economicsData, _ := economics.NewEconomicsData(args)
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
	gasMap["DelegationOps"] = value
	gasMap["UnStakeTokens"] = value
	gasMap["UnBondTokens"] = value
	gasMap["DelegationMgrOps"] = value
	gasMap["GetAllNodeStates"] = value

	return gasMap
}
