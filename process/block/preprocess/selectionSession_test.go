package preprocess

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewSelectionSession(t *testing.T) {
	t.Parallel()

	session, err := NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:       nil,
		TransactionsProcessor: &testscommon.TxProcessorStub{},
	})
	require.Nil(t, session)
	require.ErrorIs(t, err, process.ErrNilAccountsAdapter)

	session, err = NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:       &stateMock.AccountsStub{},
		TransactionsProcessor: nil,
	})
	require.Nil(t, session)
	require.ErrorIs(t, err, process.ErrNilTxProcessor)

	session, err = NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:       &stateMock.AccountsStub{},
		TransactionsProcessor: &testscommon.TxProcessorStub{},
	})
	require.NoError(t, err)
	require.NotNil(t, session)
}

func TestSelectionSession_GetAccountState(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}
	processor := &testscommon.TxProcessorStub{}

	accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, []byte("alice")) {
			return &stateMock.UserAccountStub{
				Address: []byte("alice"),
				Nonce:   42,
			}, nil
		}

		if bytes.Equal(address, []byte("bob")) {
			return &stateMock.UserAccountStub{
				Address: []byte("bob"),
				Nonce:   7,
				IsGuardedCalled: func() bool {
					return true
				},
			}, nil
		}

		return nil, fmt.Errorf("account not found: %s", address)
	}

	session, err := NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:       accounts,
		TransactionsProcessor: processor,
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	state, err := session.GetAccountState([]byte("alice"))
	require.NoError(t, err)
	require.Equal(t, uint64(42), state.Nonce)

	state, err = session.GetAccountState([]byte("bob"))
	require.NoError(t, err)
	require.Equal(t, uint64(7), state.Nonce)

	state, err = session.GetAccountState([]byte("carol"))
	require.ErrorContains(t, err, "account not found: carol")
	require.Nil(t, state)
}

func TestSelectionSession_IsIncorrectlyGuarded(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}
	processor := &testscommon.TxProcessorStub{}

	accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, []byte("bob")) {
			return &stateMock.BaseAccountMock{}, nil
		}

		return &stateMock.UserAccountStub{}, nil
	}

	processor.VerifyGuardianCalled = func(tx *transaction.Transaction, account state.UserAccountHandler) error {
		if tx.Nonce == 43 {
			return process.ErrTransactionNotExecutable
		}
		if tx.Nonce == 44 {
			return fmt.Errorf("arbitrary processing error")
		}

		return nil
	}

	session, err := NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:       accounts,
		TransactionsProcessor: processor,
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	isIncorrectlyGuarded := session.IsIncorrectlyGuarded(&transaction.Transaction{Nonce: 42, SndAddr: []byte("alice")})
	require.False(t, isIncorrectlyGuarded)

	isIncorrectlyGuarded = session.IsIncorrectlyGuarded(&transaction.Transaction{Nonce: 43, SndAddr: []byte("alice")})
	require.True(t, isIncorrectlyGuarded)

	isIncorrectlyGuarded = session.IsIncorrectlyGuarded(&transaction.Transaction{Nonce: 44, SndAddr: []byte("alice")})
	require.False(t, isIncorrectlyGuarded)

	isIncorrectlyGuarded = session.IsIncorrectlyGuarded(&transaction.Transaction{Nonce: 45, SndAddr: []byte("bob")})
	require.True(t, isIncorrectlyGuarded)
}

func TestSelectionSession_ephemeralAccountsCache_IsSharedAmongCalls(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}
	processor := &testscommon.TxProcessorStub{}

	numCallsGetExistingAccount := 0

	accounts.GetExistingAccountCalled = func(_ []byte) (vmcommon.AccountHandler, error) {
		numCallsGetExistingAccount++
		return &stateMock.UserAccountStub{}, nil
	}

	session, err := NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:       accounts,
		TransactionsProcessor: processor,
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	_, _ = session.GetAccountState([]byte("alice"))
	require.Equal(t, 1, numCallsGetExistingAccount)

	_, _ = session.GetAccountState([]byte("alice"))
	require.Equal(t, 1, numCallsGetExistingAccount)

	_ = session.IsIncorrectlyGuarded(&transaction.Transaction{Nonce: 42, SndAddr: []byte("alice")})
	require.Equal(t, 1, numCallsGetExistingAccount)

	_, _ = session.GetAccountState([]byte("bob"))
	require.Equal(t, 2, numCallsGetExistingAccount)

	_, _ = session.GetAccountState([]byte("bob"))
	require.Equal(t, 2, numCallsGetExistingAccount)

	_ = session.IsIncorrectlyGuarded(&transaction.Transaction{Nonce: 42, SndAddr: []byte("bob")})
	require.Equal(t, 2, numCallsGetExistingAccount)
}
