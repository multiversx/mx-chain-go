package factory_test

import (
	"math/big"
	"strings"
	"sync"
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go/common"
	commonFactory "github.com/ElrondNetwork/elrond-go/common/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/mainFactoryMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/outport"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ------------ Test TestProcessComponents --------------------
func TestProcessComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	processArgs := getProcessComponentsArgs(shardCoordinator)
	pcf, err := factory.NewProcessComponentsFactory(processArgs)
	require.Nil(t, err)

	pc, err := pcf.Create()
	require.Nil(t, err)

	err = pc.Close()
	require.NoError(t, err)
}

func TestProcessComponentsFactory_CreateWithInvalidTxAccumulatorTimeExpectError(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	processArgs := getProcessComponentsArgs(shardCoordinator)
	processArgs.Config.Antiflood.TxAccumulator.MaxAllowedTimeInMilliseconds = 0
	pcf, err := factory.NewProcessComponentsFactory(processArgs)
	require.Nil(t, err)

	instance, err := pcf.Create()
	require.Nil(t, instance)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), process.ErrInvalidValue.Error()))
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
	gasSchedule[common.MetaChainSystemSCsCost] = FillGasMapMetaChainSystemSCsCosts(1)

	gasScheduleNotifier := &testscommon.GasScheduleNotifierMock{
		GasSchedule: gasSchedule,
	}

	nodesCoordinator := &shardingMocks.NodesCoordinatorMock{}
	statusComponents := getStatusComponents(
		coreComponents,
		networkComponents,
		dataComponents,
		stateComponents,
		shardCoordinator,
		nodesCoordinator,
	)

	bootstrapComponentsFactoryArgs := getBootStrapArgs()

	bootstrapComponentsFactory, _ := factory.NewBootstrapComponentsFactory(bootstrapComponentsFactoryArgs)
	bootstrapComponents, _ := factory.NewManagedBootstrapComponents(bootstrapComponentsFactory)
	_ = bootstrapComponents.Create()
	factory.SetShardCoordinator(shardCoordinator, bootstrapComponents)

	return factory.ProcessComponentsFactoryArgs{
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
			GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*block.MiniBlock, map[uint32]*outportcore.Pool, error) {
				txsPool := make(map[uint32]*outportcore.Pool)
				for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
					txsPool[i] = &outportcore.Pool{}
				}

				return make([]*block.MiniBlock, 4), txsPool, nil
			},
		},
		SmartContractParser:    &mock.SmartContractParserStub{},
		GasSchedule:            gasScheduleNotifier,
		NodesCoordinator:       nodesCoordinator,
		Data:                   dataComponents,
		CoreData:               coreComponents,
		Crypto:                 cryptoComponents,
		State:                  stateComponents,
		Network:                networkComponents,
		StatusComponents:       statusComponents,
		BootstrapComponents:    bootstrapComponents,
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

func TestProcessComponents_IndexGenesisBlocks(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	processArgs := getProcessComponentsArgs(shardCoordinator)
	processArgs.Data = &mock.DataComponentsMock{
		Storage: &mock.ChainStorerMock{},
	}

	saveBlockCalledMutex := sync.Mutex{}

	outportHandler := &outport.OutportStub{
		HasDriversCalled: func() bool {
			return true
		},
		SaveBlockCalled: func(args *outportcore.ArgsSaveBlockData) {
			saveBlockCalledMutex.Lock()
			require.NotNil(t, args)

			bodyRequired := &dataBlock.Body{
				MiniBlocks: make([]*block.MiniBlock, 4),
			}

			txsPoolRequired := &outportcore.Pool{}

			assert.Equal(t, txsPoolRequired, args.TransactionsPool)
			assert.Equal(t, bodyRequired, args.Body)
			saveBlockCalledMutex.Unlock()
		},
	}

	processArgs.StatusComponents = &mainFactoryMocks.StatusComponentsStub{
		Outport: outportHandler,
	}

	pcf, err := factory.NewProcessComponentsFactory(processArgs)
	require.Nil(t, err)

	genesisBlocks := make(map[uint32]coreData.HeaderHandler)
	indexingData := make(map[uint32]*genesis.IndexingData)

	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		genesisBlocks[i] = &block.Header{}
	}

	err = pcf.IndexGenesisBlocks(genesisBlocks, indexingData)
	require.Nil(t, err)
}
