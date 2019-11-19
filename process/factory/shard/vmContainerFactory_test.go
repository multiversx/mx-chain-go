package shard

import (
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
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

func TestNewVMContainerFactory_NilGasScheduleShouldErr(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		10000,
		nil,
		createMockVMAccountsArguments(),
	)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilGasSchedule, err)
}

func TestNewVMContainerFactory_OkValues(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		10000,
		arwenConfig.MakeGasMap(1),
		createMockVMAccountsArguments(),
	)

	assert.NotNil(t, vmf)
	assert.Nil(t, err)
}

func TestVmContainerFactory_Create(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		10000,
		arwenConfig.MakeGasMap(1),
		createMockVMAccountsArguments(),
	)
	assert.NotNil(t, vmf)
	assert.Nil(t, err)

	container, err := vmf.Create()
	assert.Nil(t, err)
	assert.NotNil(t, container)

	/*vm, err := container.Get(factory.IELEVirtualMachine)
	assert.Nil(t, err)
	assert.NotNil(t, vm)*/

	acc := vmf.BlockChainHookImpl()
	assert.NotNil(t, acc)
}
