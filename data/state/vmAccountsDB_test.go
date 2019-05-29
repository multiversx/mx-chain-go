package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewVMAccountsDB_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	vadb, err := state.NewVMAccountsDB(nil, mock.NewAddressConverterFake(32, ""))

	assert.Nil(t, vadb)
	assert.Equal(t, state.ErrNilAccountsAdapter, err)
}

func TestNewVMAccountsDB_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	vadb, err := state.NewVMAccountsDB(&mock.AccountsStub{}, nil)

	assert.Nil(t, vadb)
	assert.Equal(t, state.ErrNilAddressConverter, err)
}

func TestNewVMAccountsDB_ShouldWork(t *testing.T) {
	t.Parallel()

	vadb, err := state.NewVMAccountsDB(&mock.AccountsStub{}, mock.NewAddressConverterFake(32, ""))

	assert.NotNil(t, vadb)
	assert.Nil(t, err)
}

//------- AccountExists

func TestVMAccountsDB_AccountExistsErrorsShouldRetFalseAndErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}, mock.NewAddressConverterFake(32, ""))

	accountsExists, err := vadb.AccountExists(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.False(t, accountsExists)
}

func TestVMAccountsDB_AccountExistsDoesNotExistsRetFalseAndNil(t *testing.T) {
	t.Parallel()

	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}, mock.NewAddressConverterFake(32, ""))

	accountsExists, err := vadb.AccountExists(make([]byte, 0))

	assert.False(t, accountsExists)
	assert.Nil(t, err)
}

func TestVMAccountsDB_AccountExistsDoesExistsRetTrueAndNil(t *testing.T) {
	t.Parallel()

	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return &mock.AccountWrapMock{}, nil
		},
	}, mock.NewAddressConverterFake(32, ""))

	accountsExists, err := vadb.AccountExists(make([]byte, 0))

	assert.Nil(t, err)
	assert.True(t, accountsExists)
}

//------- GetBalance

func TestVMAccountsDB_GetBalanceWrongAccountTypeShouldErr(t *testing.T) {
	t.Parallel()

	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return &mock.AccountWrapMock{}, nil
		},
	}, mock.NewAddressConverterFake(32, ""))

	balance, err := vadb.GetBalance(make([]byte, 0))

	assert.Equal(t, state.ErrWrongTypeAssertion, err)
	assert.Nil(t, balance)
}

func TestVMAccountsDB_GetBalanceGetAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}, mock.NewAddressConverterFake(32, ""))

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
	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}, mock.NewAddressConverterFake(32, ""))

	balance, err := vadb.GetBalance(make([]byte, 0))

	assert.Nil(t, err)
	assert.Equal(t, accnt.Balance, balance)
}

//------- GetNonce

func TestVMAccountsDB_GetNonceGetAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}, mock.NewAddressConverterFake(32, ""))

	nonce, err := vadb.GetNonce(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Nil(t, nonce)
}

func TestVMAccountsDB_GetNonceShouldWork(t *testing.T) {
	t.Parallel()

	accnt := &state.Account{
		Nonce:   1,
		Balance: big.NewInt(2),
	}
	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}, mock.NewAddressConverterFake(32, ""))

	nonce, err := vadb.GetNonce(make([]byte, 0))

	assert.Nil(t, err)
	assert.Equal(t, accnt.Nonce, nonce.Uint64())
}

//------- GetStorageData

func TestVMAccountsDB_GetStorageAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}, mock.NewAddressConverterFake(32, ""))

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

	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}, mock.NewAddressConverterFake(32, ""))

	value, err := vadb.GetStorageData(make([]byte, 0), variableIdentifier)

	assert.Nil(t, err)
	assert.Equal(t, variableValue, value)
}

//------- IsCodeEmpty

func TestVMAccountsDB_IsCodeEmptyAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}, mock.NewAddressConverterFake(32, ""))

	isEmpty, err := vadb.IsCodeEmpty(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.False(t, isEmpty)
}

func TestVMAccountsDB_IsCodeEmptyShouldWork(t *testing.T) {
	t.Parallel()

	accnt := mock.NewAccountWrapMock(nil, nil)

	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}, mock.NewAddressConverterFake(32, ""))

	isEmpty, err := vadb.IsCodeEmpty(make([]byte, 0))

	assert.Nil(t, err)
	assert.True(t, isEmpty)
}

//------- GetCode

func TestVMAccountsDB_GetCodeAccountErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, errExpected
		},
	}, mock.NewAddressConverterFake(32, ""))

	retrievedCode, err := vadb.GetCode(make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Nil(t, retrievedCode)
}

func TestVMAccountsDB_GetCodeShouldWork(t *testing.T) {
	t.Parallel()

	code := []byte("code")
	accnt := mock.NewAccountWrapMock(nil, nil)
	accnt.SetCode(code)

	vadb, _ := state.NewVMAccountsDB(&mock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return accnt, nil
		},
	}, mock.NewAddressConverterFake(32, ""))

	retrievedCode, err := vadb.GetCode(make([]byte, 0))

	assert.Nil(t, err)
	assert.Equal(t, code, retrievedCode)
}
