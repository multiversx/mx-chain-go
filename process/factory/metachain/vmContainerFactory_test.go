package metachain

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
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
	)

	assert.NotNil(t, vmf)
	assert.Nil(t, err)
}

func TestVmContainerFactory_Create(t *testing.T) {
	t.Parallel()

	economicsData, _ := economics.NewEconomicsData(
		&config.ConfigEconomics{
			EconomicsAddresses: config.EconomicsAddresses{
				CommunityAddress: "addr1",
				BurnAddress:      "addr2",
			},
			RewardsSettings: config.RewardsSettings{
				RewardsValue:        "1000",
				CommunityPercentage: 0.10,
				LeaderPercentage:    0.50,
				BurnPercentage:      0.40,
			},
			FeeSettings: config.FeeSettings{
				MaxGasLimitPerBlock: "10000000000",
				MinGasPrice:         "10",
				MinGasLimit:         "10",
			},
			ValidatorSettings: config.ValidatorSettings{
				StakeValue:    "500",
				UnBoundPeriod: "1000",
			},
		},
	)

	vmf, err := NewVMContainerFactory(
		createMockVMAccountsArguments(),
		economicsData,
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
