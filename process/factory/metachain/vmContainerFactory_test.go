package metachain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
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

	vmf, err := NewVMContainerFactory(
		createMockVMAccountsArguments(),
		&economics.EconomicsData{},
		&mock.MessageSignVerifierMock{},
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
