package systemSmartContracts

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
)

func TestNewVMContext_NilBlockChainHook(t *testing.T) {
	t.Parallel()

	args := createDefaultEeiArgs()
	args.BlockChainHook = nil
	vmCtx, err := NewVMContext(args)

	assert.Nil(t, vmCtx)
	assert.Equal(t, vm.ErrNilBlockchainHook, err)
}

func TestNewVMContext_NilCryptoHook(t *testing.T) {
	t.Parallel()

	args := createDefaultEeiArgs()
	args.CryptoHook = nil
	vmCtx, err := NewVMContext(args)

	assert.Nil(t, vmCtx)
	assert.Equal(t, vm.ErrNilCryptoHook, err)
}

func TestNewVMContext_NilValidatorsAccountsDB(t *testing.T) {
	t.Parallel()

	args := createDefaultEeiArgs()
	args.ValidatorAccountsDB = nil
	vmCtx, err := NewVMContext(args)

	assert.Nil(t, vmCtx)
	assert.Equal(t, vm.ErrNilValidatorAccountsDB, err)
}

func TestNewVMContext_NilChanceComputer(t *testing.T) {
	t.Parallel()

	args := createDefaultEeiArgs()
	args.ChanceComputer = nil
	vmCtx, err := NewVMContext(args)

	assert.Nil(t, vmCtx)
	assert.Equal(t, vm.ErrNilChanceComputer, err)
}

func TestNewVMContext_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	args := createDefaultEeiArgs()
	args.EnableEpochsHandler = nil
	vmCtx, err := NewVMContext(args)

	assert.Nil(t, vmCtx)
	assert.Equal(t, vm.ErrNilEnableEpochsHandler, err)
	assert.True(t, check.IfNil(vmCtx))
}

func TestNewVMContext(t *testing.T) {
	t.Parallel()

	args := createDefaultEeiArgs()
	vmCtx, err := NewVMContext(args)
	assert.False(t, check.IfNil(vmCtx))
	assert.Nil(t, err)
}

func TestVmContext_CleanCache(t *testing.T) {
	t.Parallel()

	vmCtx, _ := NewVMContext(createDefaultEeiArgs())

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

	args := createDefaultEeiArgs()
	args.BlockChainHook = blockChainHook
	vmCtx, _ := NewVMContext(args)

	res := vmCtx.GetBalance(addr)
	assert.Equal(t, res.Uint64(), balance.Uint64())
}

func TestVmContext_CreateVMOutput_Empty(t *testing.T) {
	t.Parallel()

	vmCtx, _ := NewVMContext(createDefaultEeiArgs())

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

	vmCtx, _ := NewVMContext(createDefaultEeiArgs())

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

	vmCtx, _ := NewVMContext(createDefaultEeiArgs())

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

	args := createDefaultEeiArgs()
	args.ValidatorAccountsDB = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return nil, errors.New("not found")
		},
	}
	vmCtx, _ := NewVMContext(args)

	assert.False(t, vmCtx.IsValidator([]byte("bls key")))
}

func TestVmContext_IsValidatorInvalidAccountTypeShouldRetFalse(t *testing.T) {
	t.Parallel()

	args := createDefaultEeiArgs()
	args.ValidatorAccountsDB = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return state.NewEmptyUserAccount(), nil
		},
	}
	vmCtx, _ := NewVMContext(args)

	assert.False(t, vmCtx.IsValidator([]byte("bls key")))
}

func TestVmContext_IsValidator(t *testing.T) {
	t.Parallel()

	type testIO struct {
		peerType       common.PeerType
		expectedResult bool
	}

	testData := []testIO{
		{
			peerType:       common.LeavingList,
			expectedResult: true,
		},
		{
			peerType:       common.EligibleList,
			expectedResult: true,
		},
		{
			peerType:       common.WaitingList,
			expectedResult: true,
		},
		{
			peerType:       common.NewList,
			expectedResult: false,
		},
		{
			peerType:       common.JailedList,
			expectedResult: false,
		},
	}

	for _, tio := range testData {
		blsKey := []byte("bls key")
		args := createDefaultEeiArgs()
		args.ValidatorAccountsDB = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				assert.Equal(t, blsKey, address)

				acnt := state.NewEmptyPeerAccount()
				acnt.List = string(tio.peerType)

				return acnt, nil
			},
		}
		vmCtx, _ := NewVMContext(args)

		assert.Equal(t, tio.expectedResult, vmCtx.IsValidator(blsKey))
	}
}

func TestVmContext_CleanStorage(t *testing.T) {
	t.Parallel()

	vmCtx, _ := NewVMContext(createDefaultEeiArgs())

	vmCtx.CleanCache()
	vmCtx.storageUpdate["address"] = make(map[string][]byte)
	vmCtx.storageUpdate["address"]["key"] = []byte("someData")
	vmCtx.CleanStorageUpdates()
	assert.Equal(t, 0, len(vmCtx.storageUpdate))
}
