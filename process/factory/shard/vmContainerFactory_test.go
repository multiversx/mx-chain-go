package shard

import (
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewVMContainerFactory_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		nil,
		&mock.AddressConverterMock{},
		10000,
		arwenConfig.MakeGasMap(1),
	)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewVMContainerFactory_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		&mock.AccountsStub{},
		nil,
		10000,
		arwenConfig.MakeGasMap(1),
	)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewVMContainerFactory_NilGasScheduleShouldErr(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		10000,
		nil,
	)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilGasSchedule, err)
}

func TestNewVMContainerFactory_OkValues(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		10000,
		arwenConfig.MakeGasMap(1),
	)

	assert.NotNil(t, vmf)
	assert.Nil(t, err)
}

func TestVmContainerFactory_Create(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		10000,
		arwenConfig.MakeGasMap(1),
	)
	assert.NotNil(t, vmf)
	assert.Nil(t, err)

	container, err := vmf.Create()
	assert.Nil(t, err)
	assert.NotNil(t, container)

	vm, err := container.Get(factory.IELEVirtualMachine)
	assert.Nil(t, err)
	assert.NotNil(t, vm)

	acc := vmf.VMAccountsDB()
	assert.NotNil(t, acc)
}
