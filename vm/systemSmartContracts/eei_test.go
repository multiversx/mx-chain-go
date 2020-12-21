package systemSmartContracts

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewVMContext_NilBlockChainHook(t *testing.T) {
	t.Parallel()

	vmCtx, err := NewVMContext(
		nil,
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	assert.Nil(t, vmCtx)
	assert.Equal(t, vm.ErrNilBlockchainHook, err)
}

func TestNewVMContext_NilCryptoHook(t *testing.T) {
	t.Parallel()

	vmCtx, err := NewVMContext(
		&mock.BlockChainHookStub{},
		nil,
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	assert.Nil(t, vmCtx)
	assert.Equal(t, vm.ErrNilCryptoHook, err)
}

func TestNewVMContext(t *testing.T) {
	t.Parallel()

	vmCtx, err := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	assert.NotNil(t, vmCtx)
	assert.Nil(t, err)
}

func TestVmContext_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	vmCtx, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})
	assert.False(t, check.IfNil(vmCtx))

	vmCtx = nil
	assert.True(t, check.IfNil(vmCtx))
}

func TestVmContext_CleanCache(t *testing.T) {
	t.Parallel()

	vmCtx, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	vmCtx.CleanCache()

	vmOutput := vmCtx.CreateVMOutput()
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

	vmCtx, _ := NewVMContext(
		blockChainHook,
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	res := vmCtx.GetBalance(addr)
	assert.Equal(t, res.Uint64(), balance.Uint64())
}

func TestVmContext_CreateVMOutput_Empty(t *testing.T) {
	t.Parallel()

	vmCtx, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	vmOutput := vmCtx.CreateVMOutput()
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

	vmCtx, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	addr := "smartcontract"
	vmCtx.SetSCAddress([]byte(addr))

	key := []byte("key")
	data := []byte("data")
	vmCtx.SetStorage(key, data)

	res := vmCtx.GetStorage(key)
	assert.True(t, bytes.Equal(data, res))

	vmOutput := vmCtx.CreateVMOutput()
	assert.Equal(t, 1, len(vmOutput.OutputAccounts))

	assert.True(t, bytes.Equal(vmOutput.OutputAccounts[addr].StorageUpdates[string(key)].Data, data))
}

func TestVmContext_Transfer(t *testing.T) {
	t.Parallel()

	vmCtx, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{},
		&mock.RaterMock{})

	destination := []byte("dest")
	sender := []byte("sender")
	value := big.NewInt(999)
	input := []byte("input")

	err := vmCtx.Transfer(destination, sender, value, input, 0)
	assert.Nil(t, err)

	balance := vmCtx.GetBalance(destination)
	assert.Equal(t, value.Uint64(), balance.Uint64())

	balance = vmCtx.GetBalance(sender)
	assert.Equal(t, value.Int64(), -1*balance.Int64())

	vmOutput := vmCtx.CreateVMOutput()
	assert.Equal(t, 2, len(vmOutput.OutputAccounts))
}

func TestVmContext_IsValidatorNonexistentAccountShouldRetFalse(t *testing.T) {
	t.Parallel()

	vmCtx, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (state.AccountHandler, error) {
				return nil, errors.New("not found")
			},
		},
		&mock.RaterMock{})

	assert.False(t, vmCtx.IsValidator([]byte("bls key")))
}

func TestVmContext_IsValidatorInvalidAccountTypeShouldRetFalse(t *testing.T) {
	t.Parallel()

	vmCtx, _ := NewVMContext(
		&mock.BlockChainHookStub{},
		hooks.NewVMCryptoHook(),
		&mock.ArgumentParserMock{},
		&mock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (state.AccountHandler, error) {
				return state.NewEmptyUserAccount(), nil
			},
		},
		&mock.RaterMock{})

	assert.False(t, vmCtx.IsValidator([]byte("bls key")))
}

func TestVmContext_IsValidator(t *testing.T) {
	t.Parallel()

	type testIO struct {
		peerType       core.PeerType
		expectedResult bool
	}

	testData := []testIO{
		{
			peerType:       core.LeavingList,
			expectedResult: true,
		},
		{
			peerType:       core.EligibleList,
			expectedResult: true,
		},
		{
			peerType:       core.WaitingList,
			expectedResult: true,
		},
		{
			peerType:       core.NewList,
			expectedResult: false,
		},
		{
			peerType:       core.JailedList,
			expectedResult: false,
		},
	}

	for _, tio := range testData {
		blsKey := []byte("bls key")
		vmCtx, _ := NewVMContext(
			&mock.BlockChainHookStub{},
			hooks.NewVMCryptoHook(),
			&mock.ArgumentParserMock{},
			&mock.AccountsStub{
				GetExistingAccountCalled: func(address []byte) (state.AccountHandler, error) {
					assert.Equal(t, blsKey, address)

					acnt := state.NewEmptyPeerAccount()
					acnt.List = string(tio.peerType)

					return acnt, nil
				},
			},
			&mock.RaterMock{})

		assert.Equal(t, tio.expectedResult, vmCtx.IsValidator(blsKey))
	}
}
