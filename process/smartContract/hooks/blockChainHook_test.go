package hooks_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/process"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/pkg/errors"
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

func TestNewBlockChainHookImpl_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.Accounts = nil
	bh, err := hooks.NewBlockChainHookImpl(args)

	assert.Nil(t, bh)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewBlockChainHookImpl_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.AddrConv = nil
	bh, err := hooks.NewBlockChainHookImpl(args)

	assert.Nil(t, bh)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewBlockChainHookImpl_NilStorageServiceShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.StorageService = nil
	bh, err := hooks.NewBlockChainHookImpl(args)

	assert.Nil(t, bh)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewBlockChainHookImpl_NilBlockChainShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.BlockChain = nil
	bh, err := hooks.NewBlockChainHookImpl(args)

	assert.Nil(t, bh)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewBlockChainHookImpl_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.ShardCoordinator = nil
	bh, err := hooks.NewBlockChainHookImpl(args)

	assert.Nil(t, bh)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewBlockChainHookImpl_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.Marshalizer = nil
	bh, err := hooks.NewBlockChainHookImpl(args)

	assert.Nil(t, bh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewBlockChainHookImpl_NilUint64ConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.Uint64Converter = nil
	bh, err := hooks.NewBlockChainHookImpl(args)

	assert.Nil(t, bh)
	assert.Equal(t, process.ErrNilUint64Converter, err)
}

func TestNewBlockChainHookImpl_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	bh, err := hooks.NewBlockChainHookImpl(args)

	assert.NotNil(t, bh)
	assert.Nil(t, err)
}

//------- AccountExists

func TestBlockChainHookImpl_AccountExistsErrorsShouldRetFalseAndErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	args.AddrConv = mock.NewAddressConverterFake(32, "")
	bh, _ := hooks.NewBlockChainHookImpl(args)

	accountsExists, err := bh.AccountExists(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.False(t, accountsExists)
}

func TestBlockChainHookImpl_AccountExistsDoesNotExistsRetFalseAndNil(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}
	args.AddrConv = mock.NewAddressConverterFake(32, "")
	bh, _ := hooks.NewBlockChainHookImpl(args)

	accountsExists, err := bh.AccountExists(make([]byte, 0))

	assert.False(t, accountsExists)
	assert.Nil(t, err)
}

func TestBlockChainHookImpl_AccountExistsDoesExistsRetTrueAndNil(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	bh, _ := hooks.NewBlockChainHookImpl(args)

	accountsExists, err := bh.AccountExists(make([]byte, 0))

	assert.Nil(t, err)
	assert.True(t, accountsExists)
}

//------- GetBalance

func TestBlockChainHookImpl_GetBalanceWrongAccountTypeShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	bh, _ := hooks.NewBlockChainHookImpl(args)

	balance, err := bh.GetBalance(make([]byte, 0))

	assert.Equal(t, state.ErrWrongTypeAssertion, err)
	assert.Nil(t, balance)
}

func TestBlockChainHookImpl_GetBalanceGetAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	balance, err := bh.GetBalance(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Nil(t, balance)
}

func TestBlockChainHookImpl_GetBalanceShouldWork(t *testing.T) {
	t.Parallel()

	accnt := &state.Account{
		Nonce:   1,
		Balance: big.NewInt(2),
	}

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	balance, err := bh.GetBalance(make([]byte, 0))

	assert.Nil(t, err)
	assert.Equal(t, accnt.Balance, balance)
}

//------- GetNonce

func TestBlockChainHookImpl_GetNonceGetAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	nonce, err := bh.GetNonce(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Equal(t, nonce, uint64(0))
}

func TestBlockChainHookImpl_GetNonceShouldWork(t *testing.T) {
	t.Parallel()

	accnt := &state.Account{
		Nonce:   1,
		Balance: big.NewInt(2),
	}

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	nonce, err := bh.GetNonce(make([]byte, 0))

	assert.Nil(t, err)
	assert.Equal(t, accnt.Nonce, nonce)
}

//------- GetStorageData

func TestBlockChainHookImpl_GetStorageAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	value, err := bh.GetStorageData(make([]byte, 0), make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Nil(t, value)
}

func TestBlockChainHookImpl_GetStorageDataShouldWork(t *testing.T) {
	t.Parallel()

	variableIdentifier := []byte("variable")
	variableValue := []byte("value")
	accnt := mock.NewAccountWrapMock(nil, nil)
	accnt.DataTrieTracker().SaveKeyValue(variableIdentifier, variableValue)

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	value, err := bh.GetStorageData(make([]byte, 0), variableIdentifier)

	assert.Nil(t, err)
	assert.Equal(t, variableValue, value)
}

//------- IsCodeEmpty

func TestBlockChainHookImpl_IsCodeEmptyAccountErrorsShouldErrAndRetFalse(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	isEmpty, err := bh.IsCodeEmpty(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.False(t, isEmpty)
}

func TestBlockChainHookImpl_IsCodeEmptyShouldWork(t *testing.T) {
	t.Parallel()

	accnt := mock.NewAccountWrapMock(nil, nil)

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	isEmpty, err := bh.IsCodeEmpty(make([]byte, 0))

	assert.Nil(t, err)
	assert.True(t, isEmpty)
}

//------- GetCode

func TestBlockChainHookImpl_GetCodeAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	retrievedCode, err := bh.GetCode(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Nil(t, retrievedCode)
}

func TestBlockChainHookImpl_GetCodeShouldWork(t *testing.T) {
	t.Parallel()

	code := []byte("code")
	accnt := mock.NewAccountWrapMock(nil, nil)
	accnt.SetCode(code)

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	retrievedCode, err := bh.GetCode(make([]byte, 0))

	assert.Nil(t, err)
	assert.Equal(t, code, retrievedCode)
}

func TestBlockChainHookImpl_CleanFakeAccounts(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	bh, _ := hooks.NewBlockChainHookImpl(args)

	address := []byte("test")
	bh.AddTempAccount(address, big.NewInt(10), 10)
	bh.CleanTempAccounts()

	acc := bh.TempAccount(address)
	assert.Nil(t, acc)
}

func TestBlockChainHookImpl_CreateAndGetFakeAccounts(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	bh, _ := hooks.NewBlockChainHookImpl(args)

	address := []byte("test")
	nonce := uint64(10)
	bh.AddTempAccount(address, big.NewInt(10), nonce)

	acc := bh.TempAccount(address)
	assert.NotNil(t, acc)
	assert.Equal(t, nonce, acc.GetNonce())
}

func TestBlockChainHookImpl_GetNonceFromFakeAccount(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	bh, _ := hooks.NewBlockChainHookImpl(args)

	address := []byte("test")
	nonce := uint64(10)
	bh.AddTempAccount(address, big.NewInt(10), nonce)

	getNonce, err := bh.GetNonce(address)
	assert.Nil(t, err)
	assert.Equal(t, nonce, getNonce)
}

func TestBlockChainHookImpl_NewAddressLengthNoGood(t *testing.T) {
	t.Parallel()

	adrConv := mock.NewAddressConverterFake(32, "")
	acnts := &mock.AccountsStub{}
	acnts.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (state.AccountHandler, error) {
		return &state.Account{
			Nonce:    0,
			Balance:  nil,
			CodeHash: nil,
			RootHash: nil,
		}, nil
	}
	args := createMockVMAccountsArguments()
	args.AddrConv = adrConv
	args.Accounts = acnts
	bh, _ := hooks.NewBlockChainHookImpl(args)

	address := []byte("test")
	nonce := uint64(10)

	scAddress, err := bh.NewAddress(address, nonce, []byte("00"))
	assert.Equal(t, hooks.ErrAddressLengthNotCorrect, err)
	assert.Nil(t, scAddress)

	address = []byte("1234567890123456789012345678901234567890")
	scAddress, err = bh.NewAddress(address, nonce, []byte("00"))
	assert.Equal(t, hooks.ErrAddressLengthNotCorrect, err)
	assert.Nil(t, scAddress)
}

func TestBlockChainHookImpl_NewAddressShardIdIncorrect(t *testing.T) {
	t.Parallel()

	adrConv := mock.NewAddressConverterFake(32, "")
	acnts := &mock.AccountsStub{}
	testErr := errors.New("testErr")
	acnts.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (state.AccountHandler, error) {
		return nil, testErr
	}
	args := createMockVMAccountsArguments()
	args.AddrConv = adrConv
	args.Accounts = acnts
	bh, _ := hooks.NewBlockChainHookImpl(args)

	address := []byte("012345678901234567890123456789ff")
	nonce := uint64(10)

	scAddress, err := bh.NewAddress(address, nonce, []byte("00"))
	assert.Equal(t, testErr, err)
	assert.Nil(t, scAddress)
}

func TestBlockChainHookImpl_NewAddressVMTypeTooLong(t *testing.T) {
	t.Parallel()

	adrConv := mock.NewAddressConverterFake(32, "")
	acnts := &mock.AccountsStub{}
	acnts.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (state.AccountHandler, error) {
		return &state.Account{
			Nonce:    0,
			Balance:  nil,
			CodeHash: nil,
			RootHash: nil,
		}, nil
	}
	args := createMockVMAccountsArguments()
	args.AddrConv = adrConv
	args.Accounts = acnts
	bh, _ := hooks.NewBlockChainHookImpl(args)

	address := []byte("01234567890123456789012345678900")
	nonce := uint64(10)

	vmType := []byte("010")
	scAddress, err := bh.NewAddress(address, nonce, vmType)
	assert.Equal(t, hooks.ErrVMTypeLengthIsNotCorrect, err)
	assert.Nil(t, scAddress)
}

func TestBlockChainHookImpl_NewAddress(t *testing.T) {
	t.Parallel()

	adrConv := mock.NewAddressConverterFake(32, "")
	acnts := &mock.AccountsStub{}
	acnts.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (state.AccountHandler, error) {
		return &state.Account{
			Nonce:    0,
			Balance:  nil,
			CodeHash: nil,
			RootHash: nil,
		}, nil
	}
	args := createMockVMAccountsArguments()
	args.AddrConv = adrConv
	args.Accounts = acnts
	bh, _ := hooks.NewBlockChainHookImpl(args)

	address := []byte("01234567890123456789012345678900")
	nonce := uint64(10)

	vmType := []byte("11")
	scAddress1, err := bh.NewAddress(address, nonce, vmType)
	assert.Nil(t, err)

	for i := 0; i < 8; i++ {
		assert.Equal(t, scAddress1[i], uint8(0))
	}
	assert.True(t, bytes.Equal(vmType, scAddress1[8:10]))

	nonce++
	scAddress2, err := bh.NewAddress(address, nonce, []byte("00"))
	assert.Nil(t, err)

	assert.False(t, bytes.Equal(scAddress1, scAddress2))

	fmt.Printf("%s \n%s \n", hex.EncodeToString(scAddress1), hex.EncodeToString(scAddress2))
}
