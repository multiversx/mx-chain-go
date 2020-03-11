package metachain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/stretchr/testify/assert"
)

func createMockVMAccountsArguments() hooks.ArgBlockChainHook {
	arguments := hooks.ArgBlockChainHook{
		Accounts: &mock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
				return &mock.AccountWrapMock{}, nil
			},
		},
		AddrConv:         mock.NewAddressConverterFake(32, ""),
		StorageService:   &mock.ChainStorerMock{},
		BlockChain:       &mock.BlockChainMock{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		Marshalizer:      &mock.MarshalizerMock{},
		Uint64Converter:  &mock.Uint64ByteSliceConverterMock{},
	}
	return arguments
}

func TestNewVMContainerFactory_OkValues(t *testing.T) {
	t.Parallel()

	gasSchedule := make(map[string]map[string]uint64)
	vmf, err := NewVMContainerFactory(
		createMockVMAccountsArguments(),
		&economics.EconomicsData{},
		&mock.MessageSignVerifierMock{},
		gasSchedule,
	)

	assert.NotNil(t, vmf)
	assert.Nil(t, err)
	assert.False(t, vmf.IsInterfaceNil())
}

func TestVmContainerFactory_Create(t *testing.T) {
	t.Parallel()

	economicsData, _ := economics.NewEconomicsData(
		&config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				TotalSupply:      "2000000000000000000000",
				MinimumInflation: 0,
				MaximumInflation: 0.5,
			},
			RewardsSettings: config.RewardsSettings{
				LeaderPercentage: 0.10,
			},
			FeeSettings: config.FeeSettings{
				MaxGasLimitPerBlock:  "10000000000",
				MinGasPrice:          "10",
				MinGasLimit:          "10",
				GasPerDataByte:       "1",
				DataLimitForBaseCalc: "10000",
			},
			ValidatorSettings: config.ValidatorSettings{
				GenesisNodePrice:         "500",
				UnBondPeriod:             "1000",
				TotalSupply:              "200000000000",
				MinStepValue:             "100000",
				NumNodes:                 1000,
				AuctionEnableNonce:       "0",
				StakeEnableNonce:         "0",
				NumRoundsWithoutBleed:    "1000",
				MaximumPercentageToBleed: "0.5",
				BleedPercentagePerRound:  "0.00001",
				UnJailValue:              "1000",
			},
			RatingSettings: config.RatingSettings{
				StartRating:                 5,
				MaxRating:                   10,
				MinRating:                   1,
				ProposerIncreaseRatingStep:  2,
				ProposerDecreaseRatingStep:  4,
				ValidatorIncreaseRatingStep: 1,
				ValidatorDecreaseRatingStep: 2,
			},
		},
	)

	vmf, err := NewVMContainerFactory(
		createMockVMAccountsArguments(),
		economicsData,
		&mock.MessageSignVerifierMock{},
		makeGasSchedule(),
	)
	assert.NotNil(t, vmf)
	assert.Nil(t, err)

	container, err := vmf.Create()
	assert.Nil(t, err)
	assert.NotNil(t, container)

	vm, err := container.Get(factory.SystemVirtualMachine)
	assert.Nil(t, err)
	assert.NotNil(t, vm)

	acc := vmf.BlockChainHookImpl()
	assert.NotNil(t, acc)
}

func makeGasSchedule() map[string]map[string]uint64 {
	gasSchedule := make(map[string]map[string]uint64)
	FillGasMapInternal(gasSchedule, 1)
	return gasSchedule
}

func FillGasMapInternal(gasMap map[string]map[string]uint64, value uint64) map[string]map[string]uint64 {
	gasMap[core.BaseOperationCost] = FillGasMapBaseOperationCosts(value)
	gasMap[core.MetaChainSystemSCsCost] = FillGasMapMetaChainSystemSCsCosts(value)

	return gasMap
}

func FillGasMapBaseOperationCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["StorePerByte"] = value
	gasMap["DataCopyPerByte"] = value
	gasMap["ReleasePerByte"] = value
	gasMap["PersistPerByte"] = value
	gasMap["CompilePerByte"] = value

	return gasMap
}

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

	return gasMap
}
