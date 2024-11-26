package preprocess

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewSelectionSession(t *testing.T) {
	t.Parallel()

	session, err := newSelectionSession(argsSelectionSession{
		accountsAdapter:       nil,
		transactionsProcessor: &testscommon.TxProcessorStub{},
		marshalizer:           &marshal.GogoProtoMarshalizer{},
	})
	require.Nil(t, session)
	require.ErrorIs(t, err, process.ErrNilAccountsAdapter)

	session, err = newSelectionSession(argsSelectionSession{
		accountsAdapter:       &stateMock.AccountsStub{},
		transactionsProcessor: nil,
		marshalizer:           &marshal.GogoProtoMarshalizer{},
	})
	require.Nil(t, session)
	require.ErrorIs(t, err, process.ErrNilTxProcessor)

	session, err = newSelectionSession(argsSelectionSession{
		accountsAdapter:       &stateMock.AccountsStub{},
		transactionsProcessor: &testscommon.TxProcessorStub{},
		marshalizer:           &marshal.GogoProtoMarshalizer{},
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

	session, err := newSelectionSession(argsSelectionSession{
		accountsAdapter:       accounts,
		transactionsProcessor: processor,
		marshalizer:           &marshal.GogoProtoMarshalizer{},
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

func TestSelectionSession_IsBadlyGuarded(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}
	processor := &testscommon.TxProcessorStub{}

	accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
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

	session, err := newSelectionSession(argsSelectionSession{
		accountsAdapter:       accounts,
		transactionsProcessor: processor,
		marshalizer:           &marshal.GogoProtoMarshalizer{},
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	isBadlyGuarded := session.IsBadlyGuarded(&transaction.Transaction{Nonce: 42, SndAddr: []byte("alice")})
	require.False(t, isBadlyGuarded)

	isBadlyGuarded = session.IsBadlyGuarded(&transaction.Transaction{Nonce: 43, SndAddr: []byte("alice")})
	require.True(t, isBadlyGuarded)

	isBadlyGuarded = session.IsBadlyGuarded(&transaction.Transaction{Nonce: 44, SndAddr: []byte("alice")})
	require.False(t, isBadlyGuarded)
}
