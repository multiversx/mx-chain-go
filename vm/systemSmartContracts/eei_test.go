package systemSmartContracts

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestNewVMContext_NilBlockChainHook(t *testing.T) {
	t.Parallel()

	vmContext, err := NewVMContext(
		nil,
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	assert.Nil(t, vmContext)
	assert.Equal(t, vm.ErrNilBlockchainHook, err)
}

func TestNewVMContext_NilCryptoHook(t *testing.T) {
	t.Parallel()

	vmContext, err := NewVMContext(
		&mock.BlockChainHookStub{},
		nil,
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	assert.Nil(t, vmContext)
	assert.Equal(t, vm.ErrNilCryptoHook, err)
}

func TestNewVMContext(t *testing.T) {
	t.Parallel()

	vmContext, err := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	assert.NotNil(t, vmContext)
	assert.Nil(t, err)
}

func TestVmContext_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	vmContext, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	assert.False(t, vmContext.IsInterfaceNil())

	vmContext = nil
	assert.True(t, vmContext.IsInterfaceNil())
}

func TestVmContext_CleanCache(t *testing.T) {
	t.Parallel()

	vmContext, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	vmContext.CleanCache()

	vmOutput := vmContext.CreateVMOutput()
	assert.Equal(t, 0, len(vmOutput.OutputAccounts))
}

func TestVmContext_GetBalance(t *testing.T) {
	t.Parallel()

	addr := []byte("addr")
	balance := big.NewInt(10)
	account, _ := state.NewUserAccount([]byte("123"))
	_ = account.AddToBalance(balance)

	blockChainHook := &mock.BlockChainHookStub{GetUserAccountCalled: func(address []byte) (a vmcommon.UserAccountHandler, e error) {
		if bytes.Equal(address, addr) {
			return account, nil
		}
		return nil, errors.New("get balance error")
	},
	}

	vmContext, _ := NewVMContext(
		blockChainHook,
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	res := vmContext.GetBalance(addr)
	assert.Equal(t, res.Uint64(), balance.Uint64())
}

func TestVmContext_CreateVMOutput_Empty(t *testing.T) {
	t.Parallel()

	vmContext, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	vmOutput := vmContext.CreateVMOutput()
	assert.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	assert.Equal(t, 0, len(vmOutput.ReturnData))
	assert.Equal(t, 0, len(vmOutput.OutputAccounts))
	assert.Equal(t, 0, len(vmOutput.Logs))
	assert.Equal(t, 0, len(vmOutput.DeletedAccounts))
	assert.Equal(t, 0, len(vmOutput.TouchedAccounts))
	assert.Equal(t, uint64(0), vmOutput.GasRefund.Uint64())
	assert.Equal(t, uint64(0), vmOutput.GasRemaining)
}

func TestVmContext_SetStorage(t *testing.T) {
	t.Parallel()

	vmContext, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	addr := "smartcontract"
	vmContext.SetSCAddress([]byte(addr))

	key := []byte("key")
	data := []byte("data")
	vmContext.SetStorage(key, data)

	res := vmContext.GetStorage(key)
	assert.True(t, bytes.Equal(data, res))

	vmOutput := vmContext.CreateVMOutput()
	assert.Equal(t, 1, len(vmOutput.OutputAccounts))

	assert.True(t, bytes.Equal(vmOutput.OutputAccounts[addr].StorageUpdates[string(key)].Data, data))
}

func TestVmContext_Transfer(t *testing.T) {
	t.Parallel()

	vmContext, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	destination := []byte("dest")
	sender := []byte("sender")
	value := big.NewInt(999)
	input := []byte("input")

	err := vmContext.Transfer(destination, sender, value, input, 0)
	assert.Nil(t, err)

	balance := vmContext.GetBalance(destination)
	assert.Equal(t, value.Uint64(), balance.Uint64())

	balance = vmContext.GetBalance(sender)
	assert.Equal(t, value.Int64(), -1*balance.Int64())

	vmOutput := vmContext.CreateVMOutput()
	assert.Equal(t, 2, len(vmOutput.OutputAccounts))
}
