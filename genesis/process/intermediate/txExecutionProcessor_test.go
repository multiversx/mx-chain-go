package intermediate_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/genesis/process/intermediate"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func TestNewTxExecutionProcessor_NilTxProcessorShouldErr(t *testing.T) {
	t.Parallel()

	tep, err := intermediate.NewTxExecutionProcessor(nil, &stateMock.AccountsStub{})

	assert.True(t, check.IfNil(tep))
	assert.Equal(t, process.ErrNilTxProcessor, err)
}

func TestNewTxExecutionProcessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	tep, err := intermediate.NewTxExecutionProcessor(&testscommon.TxProcessorStub{}, nil)

	assert.True(t, check.IfNil(tep))
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewTxExecutionProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	tep, err := intermediate.NewTxExecutionProcessor(&testscommon.TxProcessorStub{}, &stateMock.AccountsStub{})

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
		&testscommon.TxProcessorStub{
			ProcessTransactionCalled: func(tx *transaction.Transaction) (vmcommon.ReturnCode, error) {
				if tx.Nonce == nonce && bytes.Equal(tx.SndAddr, sndAddr) && bytes.Equal(tx.RcvAddr, recvAddr) &&
					value.Cmp(tx.Value) == 0 && bytes.Equal(tx.Data, data) {
					return 0, nil
				}

				return 0, errors.New("should not happened")
			},
		},
		&stateMock.AccountsStub{},
	)

	err := tep.ExecuteTransaction(nonce, sndAddr, recvAddr, value, data)

	assert.Nil(t, err)
}

//------- GetNonce

func TestTxExecutionProcessor_GetNonceAccountsErrShouldErr(t *testing.T) {
	t.Parallel()

	tep, _ := intermediate.NewTxExecutionProcessor(
		&testscommon.TxProcessorStub{},
		&stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
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
		&testscommon.TxProcessorStub{},
		&stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
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

func TestTxExecutionProcessor_GetAccount(t *testing.T) {
	t.Parallel()

	t.Run("GetExistingAccount error should error", func(t *testing.T) {
		t.Parallel()

		tep, _ := intermediate.NewTxExecutionProcessor(
			&testscommon.TxProcessorStub{},
			&stateMock.AccountsStub{
				GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
					return nil, expectedErr
				},
			},
		)

		acc, found := tep.GetAccount([]byte("address"))
		assert.False(t, found)
		assert.Nil(t, acc)
	})
	t.Run("GetExistingAccount returns invalid data should error", func(t *testing.T) {
		t.Parallel()

		tep, _ := intermediate.NewTxExecutionProcessor(
			&testscommon.TxProcessorStub{},
			&stateMock.AccountsStub{
				GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
					return &stateMock.PeerAccountHandlerMock{}, nil
				},
			},
		)

		acc, found := tep.GetAccount([]byte("address"))
		assert.False(t, found)
		assert.Nil(t, acc)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tep, _ := intermediate.NewTxExecutionProcessor(
			&testscommon.TxProcessorStub{},
			&stateMock.AccountsStub{
				GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
					return &stateMock.UserAccountStub{}, nil
				},
			},
		)

		acc, found := tep.GetAccount([]byte("address"))
		assert.True(t, found)
		assert.Equal(t, &stateMock.UserAccountStub{}, acc)
	})
}

//------- AddBalance

func TestTxExecutionProcessor_AddBalanceAccountsErrShouldErr(t *testing.T) {
	t.Parallel()

	tep, _ := intermediate.NewTxExecutionProcessor(
		&testscommon.TxProcessorStub{},
		&stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
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
		&testscommon.TxProcessorStub{},
		&stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
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
		&testscommon.TxProcessorStub{},
		&stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
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
		&testscommon.TxProcessorStub{},
		&stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return &mock.UserAccountMock{
					BalanceField: big.NewInt(0).Set(initialBalance),
				}, nil
			},
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
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

	tep, _ := intermediate.NewTxExecutionProcessor(
		&testscommon.TxProcessorStub{},
		&stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
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
		&testscommon.TxProcessorStub{},
		&stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return &mock.BaseAccountMock{
					Nonce: initialNonce,
				}, nil
			},
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
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

func TestTxExecutionProcessor_GetExecutedTransactionsEmpty(t *testing.T) {
	t.Parallel()

	tep, err := intermediate.NewTxExecutionProcessor(&testscommon.TxProcessorStub{}, &stateMock.AccountsStub{})
	assert.Nil(t, err)

	txs := tep.GetExecutedTransactions()
	assert.Equal(t, 0, len(txs))
}

func TestTxExecutionProcessor_GetExecutedTransactionsNonEmpty(t *testing.T) {
	t.Parallel()

	nonce := uint64(0)
	sndAddr := []byte("sender")
	recvAddr := []byte("receiver")
	value := big.NewInt(1000)
	data := []byte("data")

	tep, _ := intermediate.NewTxExecutionProcessor(
		&testscommon.TxProcessorStub{
			ProcessTransactionCalled: func(tx *transaction.Transaction) (vmcommon.ReturnCode, error) {
				if tx.Nonce == nonce && bytes.Equal(tx.SndAddr, sndAddr) && bytes.Equal(tx.RcvAddr, recvAddr) &&
					value.Cmp(tx.Value) == 0 && bytes.Equal(tx.Data, data) {
					return 0, nil
				}

				return 0, errors.New("should not happened")
			},
		},
		&stateMock.AccountsStub{},
	)

	err := tep.ExecuteTransaction(nonce, sndAddr, recvAddr, value, data)
	assert.Nil(t, err)

	txs := tep.GetExecutedTransactions()
	assert.Equal(t, 1, len(txs))
	assert.Equal(t, []byte("sender"), txs[0].GetSndAddr())
	assert.Equal(t, []byte("receiver"), txs[0].GetRcvAddr())
	assert.Equal(t, uint64(0), txs[0].GetNonce())
	assert.Equal(t, big.NewInt(1000), txs[0].GetValue())
}
