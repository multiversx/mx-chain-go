package intermediate_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/genesis/process/intermediate"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func TestNewTxExecutionProcessor_NilTxProcessorShouldErr(t *testing.T) {
	t.Parallel()

	tep, err := intermediate.NewTxExecutionProcessor(nil, &mock.AccountsStub{})

	assert.True(t, check.IfNil(tep))
	assert.Equal(t, process.ErrNilTxProcessor, err)
}

func TestNewTxExecutionProcessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	tep, err := intermediate.NewTxExecutionProcessor(&mock.TxProcessorStub{}, nil)

	assert.True(t, check.IfNil(tep))
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewTxExecutionProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	tep, err := intermediate.NewTxExecutionProcessor(&mock.TxProcessorStub{}, &mock.AccountsStub{})

	assert.False(t, check.IfNil(tep))
	assert.Nil(t, err)
}

//------- ExecuteTransaction

func TestTxExecutionProcessor_ExecuteTransaction(t *testing.T) {
	t.Parallel()

	nonce := uint64(445)
	sndAddr := []byte("sender")
	recvAddr := []byte("receiver")
	value := big.NewInt(3343)
	data := []byte("data")

	tep, _ := intermediate.NewTxExecutionProcessor(
		&mock.TxProcessorStub{
			ProcessTransactionCalled: func(tx *transaction.Transaction) (vmcommon.ReturnCode, error) {
				if tx.Nonce == nonce && bytes.Equal(tx.SndAddr, sndAddr) && bytes.Equal(tx.RcvAddr, recvAddr) &&
					value.Cmp(tx.Value) == 0 && bytes.Equal(tx.Data, data) {
					return 0, nil
				}

				return 0, errors.New("should not happened")
			},
		},
		&mock.AccountsStub{},
	)

	err := tep.ExecuteTransaction(nonce, sndAddr, recvAddr, value, data)

	assert.Nil(t, err)
}

//------- GetNonce

func TestTxExecutionProcessor_GetNonceAccountsErrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	tep, _ := intermediate.NewTxExecutionProcessor(
		&mock.TxProcessorStub{},
		&mock.AccountsStub{
			LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
				return nil, expectedErr
			},
		},
	)

	nonce, err := tep.GetNonce([]byte("sender"))

	assert.Equal(t, uint64(0), nonce)
	assert.Equal(t, expectedErr, err)
}

func TestTxExecutionProcessor_GetNonceShouldWork(t *testing.T) {
	t.Parallel()

	nonce := uint64(224323)
	tep, _ := intermediate.NewTxExecutionProcessor(
		&mock.TxProcessorStub{},
		&mock.AccountsStub{
			LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
				return &mock.BaseAccountMock{
					Nonce: nonce,
				}, nil
			},
		},
	)

	recovered, err := tep.GetNonce([]byte("sender"))

	assert.Equal(t, nonce, recovered)
	assert.Nil(t, err)
}

//------- AddBalance

func TestTxExecutionProcessor_AddBalanceAccountsErrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	tep, _ := intermediate.NewTxExecutionProcessor(
		&mock.TxProcessorStub{},
		&mock.AccountsStub{
			LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
				return nil, expectedErr
			},
		},
	)

	err := tep.AddBalance([]byte("sender"), big.NewInt(0))

	assert.Equal(t, expectedErr, err)
}

func TestTxExecutionProcessor_AddBalanceWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	nonce := uint64(224323)
	tep, _ := intermediate.NewTxExecutionProcessor(
		&mock.TxProcessorStub{},
		&mock.AccountsStub{
			LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
				return &mock.BaseAccountMock{
					Nonce: nonce,
				}, nil
			},
		},
	)

	err := tep.AddBalance([]byte("sender"), big.NewInt(0))

	assert.Equal(t, genesis.ErrWrongTypeAssertion, err)
}

func TestTxExecutionProcessor_AddBalanceNegativeValueShouldErr(t *testing.T) {
	t.Parallel()

	tep, _ := intermediate.NewTxExecutionProcessor(
		&mock.TxProcessorStub{},
		&mock.AccountsStub{
			LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
				return &mock.UserAccountMock{
					BalanceField: big.NewInt(0),
				}, nil
			},
		},
	)

	err := tep.AddBalance([]byte("sender"), big.NewInt(-1))

	assert.Equal(t, mock.ErrNegativeValue, err)
}

func TestTxExecutionProcessor_AddBalanceShouldWork(t *testing.T) {
	t.Parallel()

	initialBalance := big.NewInt(22322)
	added := big.NewInt(37843)
	tep, _ := intermediate.NewTxExecutionProcessor(
		&mock.TxProcessorStub{},
		&mock.AccountsStub{
			LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
				return &mock.UserAccountMock{
					BalanceField: big.NewInt(0).Set(initialBalance),
				}, nil
			},
			SaveAccountCalled: func(account state.AccountHandler) error {
				expectedBalance := big.NewInt(0).Set(initialBalance)
				expectedBalance.Add(expectedBalance, added)

				if expectedBalance.Cmp(account.(state.UserAccountHandler).GetBalance()) == 0 {
					return nil
				}

				return errors.New("should not happened")
			},
		},
	)

	err := tep.AddBalance([]byte("sender"), added)

	assert.Nil(t, err)
}

//------- AddNonce

func TestTxExecutionProcessor_AddNonceAccountsErrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	tep, _ := intermediate.NewTxExecutionProcessor(
		&mock.TxProcessorStub{},
		&mock.AccountsStub{
			LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
				return nil, expectedErr
			},
		},
	)

	err := tep.AddNonce([]byte("sender"), 0)

	assert.Equal(t, expectedErr, err)
}

func TestTxExecutionProcessor_AddNonceShouldWork(t *testing.T) {
	t.Parallel()

	initialNonce := uint64(83927)
	addedNonce := uint64(27826)
	tep, _ := intermediate.NewTxExecutionProcessor(
		&mock.TxProcessorStub{},
		&mock.AccountsStub{
			LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
				return &mock.BaseAccountMock{
					Nonce: initialNonce,
				}, nil
			},
			SaveAccountCalled: func(account state.AccountHandler) error {
				if account.GetNonce() == initialNonce+addedNonce {
					return nil
				}

				return errors.New("should not happen")
			},
		},
	)

	err := tep.AddNonce([]byte("sender"), addedNonce)

	assert.Nil(t, err)
}
