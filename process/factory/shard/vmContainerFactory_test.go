package shard

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewVMContainerFactory_NilAccounts(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(nil, &mock.AddressConverterMock{})

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewVMContainerFactory_NilAddressConverter(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(&mock.AccountsStub{}, nil)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewVMContainerFactory_OkValues(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(&mock.AccountsStub{}, &mock.AddressConverterMock{})

	assert.NotNil(t, vmf)
	assert.Nil(t, err)
}

func TestVmContainerFactory_Create(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(&mock.AccountsStub{}, &mock.AddressConverterMock{})
	assert.NotNil(t, vmf)
	assert.Nil(t, err)

	container, err := vmf.Create()
	assert.Nil(t, err)
	assert.NotNil(t, container)

	vm, err := container.Get([]byte(factory.IELEVirtualMachine))
	assert.Nil(t, err)
	assert.NotNil(t, vm)

	acc := vmf.VMAccountsDB()
	assert.NotNil(t, acc)
}
