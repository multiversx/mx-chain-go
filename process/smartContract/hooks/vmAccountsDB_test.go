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

func TestNewVMAccountsDB_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.Accounts = nil
	vadb, err := hooks.NewVMAccountsDB(args)

	assert.Nil(t, vadb)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewVMAccountsDB_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.AddrConv = nil
	vadb, err := hooks.NewVMAccountsDB(args)

	assert.Nil(t, vadb)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewVMAccountsDB_NilStorageServiceShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.StorageService = nil
	vadb, err := hooks.NewVMAccountsDB(args)

	assert.Nil(t, vadb)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewVMAccountsDB_NilBlockChainShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.BlockChain = nil
	vadb, err := hooks.NewVMAccountsDB(args)

	assert.Nil(t, vadb)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewVMAccountsDB_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.ShardCoordinator = nil
	vadb, err := hooks.NewVMAccountsDB(args)

	assert.Nil(t, vadb)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewVMAccountsDB_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.Marshalizer = nil
	vadb, err := hooks.NewVMAccountsDB(args)

	assert.Nil(t, vadb)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewVMAccountsDB_NilUint64ConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.Uint64Converter = nil
	vadb, err := hooks.NewVMAccountsDB(args)

	assert.Nil(t, vadb)
	assert.Equal(t, process.ErrNilUint64Converter, err)
}

func TestNewVMAccountsDB_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	vadb, err := hooks.NewVMAccountsDB(args)

	assert.NotNil(t, vadb)
	assert.Nil(t, err)
}

//------- AccountExists

func TestVMAccountsDB_AccountExistsErrorsShouldRetFalseAndErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	args.AddrConv = mock.NewAddressConverterFake(32, "")
	vadb, _ := hooks.NewVMAccountsDB(args)

	accountsExists, err := vadb.AccountExists(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.False(t, accountsExists)
}

func TestVMAccountsDB_AccountExistsDoesNotExistsRetFalseAndNil(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}
	args.AddrConv = mock.NewAddressConverterFake(32, "")
	vadb, _ := hooks.NewVMAccountsDB(args)

	accountsExists, err := vadb.AccountExists(make([]byte, 0))

	assert.False(t, accountsExists)
	assert.Nil(t, err)
}

func TestVMAccountsDB_AccountExistsDoesExistsRetTrueAndNil(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	vadb, _ := hooks.NewVMAccountsDB(args)

	accountsExists, err := vadb.AccountExists(make([]byte, 0))

	assert.Nil(t, err)
	assert.True(t, accountsExists)
}

//------- GetBalance

func TestVMAccountsDB_GetBalanceWrongAccountTypeShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	vadb, _ := hooks.NewVMAccountsDB(args)

	balance, err := vadb.GetBalance(make([]byte, 0))

	assert.Equal(t, state.ErrWrongTypeAssertion, err)
	assert.Nil(t, balance)
}

func TestVMAccountsDB_GetBalanceGetAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	vadb, _ := hooks.NewVMAccountsDB(args)

	balance, err := vadb.GetBalance(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Nil(t, balance)
}

func TestVMAccountsDB_GetBalanceShouldWork(t *testing.T) {
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
	vadb, _ := hooks.NewVMAccountsDB(args)

	balance, err := vadb.GetBalance(make([]byte, 0))

	assert.Nil(t, err)
	assert.Equal(t, accnt.Balance, balance)
}

//------- GetNonce

func TestVMAccountsDB_GetNonceGetAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	vadb, _ := hooks.NewVMAccountsDB(args)

	nonce, err := vadb.GetNonce(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Equal(t, nonce, uint64(0))
}

func TestVMAccountsDB_GetNonceShouldWork(t *testing.T) {
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
	vadb, _ := hooks.NewVMAccountsDB(args)

	nonce, err := vadb.GetNonce(make([]byte, 0))

	assert.Nil(t, err)
	assert.Equal(t, accnt.Nonce, nonce)
}

//------- GetStorageData

func TestVMAccountsDB_GetStorageAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	vadb, _ := hooks.NewVMAccountsDB(args)

	value, err := vadb.GetStorageData(make([]byte, 0), make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Nil(t, value)
}

func TestVMAccountsDB_GetStorageDataShouldWork(t *testing.T) {
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
	vadb, _ := hooks.NewVMAccountsDB(args)

	value, err := vadb.GetStorageData(make([]byte, 0), variableIdentifier)

	assert.Nil(t, err)
	assert.Equal(t, variableValue, value)
}

//------- IsCodeEmpty

func TestVMAccountsDB_IsCodeEmptyAccountErrorsShouldErrAndRetFalse(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	vadb, _ := hooks.NewVMAccountsDB(args)

	isEmpty, err := vadb.IsCodeEmpty(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.False(t, isEmpty)
}

func TestVMAccountsDB_IsCodeEmptyShouldWork(t *testing.T) {
	t.Parallel()

	accnt := mock.NewAccountWrapMock(nil, nil)

	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}
	vadb, _ := hooks.NewVMAccountsDB(args)

	isEmpty, err := vadb.IsCodeEmpty(make([]byte, 0))

	assert.Nil(t, err)
	assert.True(t, isEmpty)
}

//------- GetCode

func TestVMAccountsDB_GetCodeAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	args := createMockVMAccountsArguments()
	args.Accounts = &mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	vadb, _ := hooks.NewVMAccountsDB(args)

	retrievedCode, err := vadb.GetCode(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Nil(t, retrievedCode)
}

func TestVMAccountsDB_GetCodeShouldWork(t *testing.T) {
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
	vadb, _ := hooks.NewVMAccountsDB(args)

	retrievedCode, err := vadb.GetCode(make([]byte, 0))

	assert.Nil(t, err)
	assert.Equal(t, code, retrievedCode)
}

func TestVMAccountsDB_CleanFakeAccounts(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	vadb, _ := hooks.NewVMAccountsDB(args)

	address := []byte("test")
	vadb.AddTempAccount(address, big.NewInt(10), 10)
	vadb.CleanTempAccounts()

	acc := vadb.TempAccount(address)
	assert.Nil(t, acc)
}

func TestVMAccountsDB_CreateAndGetFakeAccounts(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	vadb, _ := hooks.NewVMAccountsDB(args)

	address := []byte("test")
	nonce := uint64(10)
	vadb.AddTempAccount(address, big.NewInt(10), nonce)

	acc := vadb.TempAccount(address)
	assert.NotNil(t, acc)
	assert.Equal(t, nonce, acc.GetNonce())
}

func TestVMAccountsDB_GetNonceFromFakeAccount(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	vadb, _ := hooks.NewVMAccountsDB(args)

	address := []byte("test")
	nonce := uint64(10)
	vadb.AddTempAccount(address, big.NewInt(10), nonce)

	getNonce, err := vadb.GetNonce(address)
	assert.Nil(t, err)
	assert.Equal(t, nonce, getNonce)
}

func TestVMAccountsDB_NewAddressLengthNoGood(t *testing.T) {
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
	vadb, _ := hooks.NewVMAccountsDB(args)

	address := []byte("test")
	nonce := uint64(10)

	scAddress, err := vadb.NewAddress(address, nonce, []byte("00"))
	assert.Equal(t, hooks.ErrAddressLengthNotCorrect, err)
	assert.Nil(t, scAddress)

	address = []byte("1234567890123456789012345678901234567890")
	scAddress, err = vadb.NewAddress(address, nonce, []byte("00"))
	assert.Equal(t, hooks.ErrAddressLengthNotCorrect, err)
	assert.Nil(t, scAddress)
}

func TestVMAccountsDB_NewAddressShardIdIncorrect(t *testing.T) {
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
	vadb, _ := hooks.NewVMAccountsDB(args)

	address := []byte("012345678901234567890123456789ff")
	nonce := uint64(10)

	scAddress, err := vadb.NewAddress(address, nonce, []byte("00"))
	assert.Equal(t, testErr, err)
	assert.Nil(t, scAddress)
}

func TestVMAccountsDB_NewAddressVMTypeTooLong(t *testing.T) {
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
	vadb, _ := hooks.NewVMAccountsDB(args)

	address := []byte("01234567890123456789012345678900")
	nonce := uint64(10)

	vmType := []byte("010")
	scAddress, err := vadb.NewAddress(address, nonce, vmType)
	assert.Equal(t, hooks.ErrVMTypeLengthIsNotCorrect, err)
	assert.Nil(t, scAddress)
}

func TestVMAccountsDB_NewAddress(t *testing.T) {
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
	vadb, _ := hooks.NewVMAccountsDB(args)

	address := []byte("01234567890123456789012345678900")
	nonce := uint64(10)

	vmType := []byte("11")
	scAddress1, err := vadb.NewAddress(address, nonce, vmType)
	assert.Nil(t, err)

	for i := 0; i < 8; i++ {
		assert.Equal(t, scAddress1[i], uint8(0))
	}
	assert.True(t, bytes.Equal(vmType, scAddress1[8:10]))

	nonce++
	scAddress2, err := vadb.NewAddress(address, nonce, []byte("00"))
	assert.Nil(t, err)

	assert.False(t, bytes.Equal(scAddress1, scAddress2))

	fmt.Printf("%s \n%s \n", hex.EncodeToString(scAddress1), hex.EncodeToString(scAddress2))
}
