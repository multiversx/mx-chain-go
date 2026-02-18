package preprocess

import (
	"bytes"
	"fmt"
	"math/big"
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
		AccountsAdapter:         nil,
		TransactionsProcessor:   &testscommon.TxProcessorStub{},
		TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{},
	})
	require.Nil(t, session)
	require.ErrorIs(t, err, state.ErrNilAccountsAdapter)

	session, err = NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:         &stateMock.AccountsStub{},
		TransactionsProcessor:   nil,
		TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{},
	})
	require.Nil(t, session)
	require.ErrorIs(t, err, process.ErrNilTxProcessor)

	session, err = NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:         &stateMock.AccountsStub{},
		TransactionsProcessor:   &testscommon.TxProcessorStub{},
		TxVersionCheckerHandler: nil,
	})
	require.Nil(t, session)
	require.ErrorIs(t, err, process.ErrNilTransactionVersionChecker)

	session, err = NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:         &stateMock.AccountsStub{},
		TransactionsProcessor:   &testscommon.TxProcessorStub{},
		TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{},
	})
	require.NoError(t, err)
	require.NotNil(t, session)
}

func TestSelectionSession_GetAccountNonceAndBalance(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}
	processor := &testscommon.TxProcessorStub{}

	accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, []byte("alice")) {
			return &stateMock.UserAccountStub{
				Address: []byte("alice"),
				Nonce:   42,
				Balance: big.NewInt(3000000000000000000),
			}, nil
		}

		if bytes.Equal(address, []byte("bob")) {
			return &stateMock.UserAccountStub{
				Address: []byte("bob"),
				Nonce:   7,
				Balance: big.NewInt(1000000000000000000),
			}, nil
		}

		return nil, state.ErrAccNotFound
	}

	session, err := NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:         accounts,
		TransactionsProcessor:   processor,
		TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{},
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	nonce, balance, existing, err := session.GetAccountNonceAndBalance([]byte("alice"))
	require.NoError(t, err)
	require.Equal(t, uint64(42), nonce)
	require.Equal(t, "3000000000000000000", balance.String())
	require.True(t, existing)

	nonce, balance, existing, err = session.GetAccountNonceAndBalance([]byte("bob"))
	require.NoError(t, err)
	require.Equal(t, uint64(7), nonce)
	require.Equal(t, "1000000000000000000", balance.String())
	require.True(t, existing)

	nonce, balance, existing, err = session.GetAccountNonceAndBalance([]byte("carol"))
	require.NoError(t, err)
	require.Equal(t, uint64(0), nonce)
	require.Equal(t, "0", balance.String())
	require.False(t, existing)
}

func TestSelectionSession_GetRootHash(t *testing.T) {
	t.Parallel()

	processor := &testscommon.TxProcessorStub{}
	accounts := &stateMock.AccountsStub{}

	accounts.RootHashCalled = func() ([]byte, error) {
		return []byte("rootHash1"), nil
	}

	session, err := NewSelectionSession(ArgsSelectionSession{
		AccountsAdapter:         accounts,
		TransactionsProcessor:   processor,
		TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{},
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	rootHash, err := session.GetRootHash()
	require.NoError(t, err)
	require.Equal(t, []byte("rootHash1"), rootHash)
}

func TestSelectionSession_IsIncorrectlyGuarded(t *testing.T) {
	t.Parallel()

	t.Run("non-guarded tx from non-guarded account should return false without GetUserAccount", func(t *testing.T) {
		t.Parallel()

		getUserAccountCalled := false

		accounts := &stateMock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.UserAccountStub{
					Address:         address,
					IsGuardedCalled: func() bool { return false },
				}, nil
			},
		}

		processor := &testscommon.TxProcessorStub{
			VerifyGuardianCalled: func(_ *transaction.Transaction, _ state.UserAccountHandler) error {
				getUserAccountCalled = true
				return nil
			},
		}

		session, err := NewSelectionSession(ArgsSelectionSession{
			AccountsAdapter:       accounts,
			TransactionsProcessor: processor,
			TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{
				IsGuardedTransactionCalled: func(_ *transaction.Transaction) bool { return false },
			},
		})
		require.NoError(t, err)

		result := session.IsIncorrectlyGuarded(&transaction.Transaction{SndAddr: []byte("alice")})
		require.False(t, result)
		require.False(t, getUserAccountCalled, "GetUserAccount/VerifyGuardian should NOT be called for non-guarded tx + non-guarded account")
	})

	t.Run("guarded tx from non-guarded account should return true without VerifyGuardian", func(t *testing.T) {
		t.Parallel()

		verifyGuardianCalled := false

		accounts := &stateMock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.UserAccountStub{
					Address:         address,
					IsGuardedCalled: func() bool { return false },
				}, nil
			},
		}

		processor := &testscommon.TxProcessorStub{
			VerifyGuardianCalled: func(_ *transaction.Transaction, _ state.UserAccountHandler) error {
				verifyGuardianCalled = true
				return nil
			},
		}

		session, err := NewSelectionSession(ArgsSelectionSession{
			AccountsAdapter:       accounts,
			TransactionsProcessor: processor,
			TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{
				IsGuardedTransactionCalled: func(_ *transaction.Transaction) bool { return true },
			},
		})
		require.NoError(t, err)

		result := session.IsIncorrectlyGuarded(&transaction.Transaction{SndAddr: []byte("alice")})
		require.True(t, result)
		require.False(t, verifyGuardianCalled, "VerifyGuardian should NOT be called for guarded tx + non-guarded account")
	})

	t.Run("non-guarded tx from guarded account should delegate to VerifyGuardian", func(t *testing.T) {
		t.Parallel()

		verifyGuardianCalled := false

		accounts := &stateMock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.UserAccountStub{
					Address:         address,
					IsGuardedCalled: func() bool { return true },
				}, nil
			},
		}

		processor := &testscommon.TxProcessorStub{
			VerifyGuardianCalled: func(_ *transaction.Transaction, _ state.UserAccountHandler) error {
				verifyGuardianCalled = true
				return process.ErrTransactionNotExecutable
			},
		}

		session, err := NewSelectionSession(ArgsSelectionSession{
			AccountsAdapter:       accounts,
			TransactionsProcessor: processor,
			TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{
				IsGuardedTransactionCalled: func(_ *transaction.Transaction) bool { return false },
			},
		})
		require.NoError(t, err)

		result := session.IsIncorrectlyGuarded(&transaction.Transaction{SndAddr: []byte("alice")})
		require.True(t, result)
		require.True(t, verifyGuardianCalled)
	})

	t.Run("guarded tx from guarded account should delegate to VerifyGuardian", func(t *testing.T) {
		t.Parallel()

		accounts := &stateMock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.UserAccountStub{
					Address:         address,
					IsGuardedCalled: func() bool { return true },
				}, nil
			},
		}

		t.Run("guardian matches (no error)", func(t *testing.T) {
			t.Parallel()

			processor := &testscommon.TxProcessorStub{
				VerifyGuardianCalled: func(_ *transaction.Transaction, _ state.UserAccountHandler) error {
					return nil
				},
			}

			session, err := NewSelectionSession(ArgsSelectionSession{
				AccountsAdapter:       accounts,
				TransactionsProcessor: processor,
				TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{
					IsGuardedTransactionCalled: func(_ *transaction.Transaction) bool { return true },
				},
			})
			require.NoError(t, err)

			result := session.IsIncorrectlyGuarded(&transaction.Transaction{SndAddr: []byte("alice")})
			require.False(t, result)
		})

		t.Run("guardian mismatch (ErrTransactionNotExecutable)", func(t *testing.T) {
			t.Parallel()

			processor := &testscommon.TxProcessorStub{
				VerifyGuardianCalled: func(_ *transaction.Transaction, _ state.UserAccountHandler) error {
					return process.ErrTransactionNotExecutable
				},
			}

			session, err := NewSelectionSession(ArgsSelectionSession{
				AccountsAdapter:       accounts,
				TransactionsProcessor: processor,
				TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{
					IsGuardedTransactionCalled: func(_ *transaction.Transaction) bool { return true },
				},
			})
			require.NoError(t, err)

			result := session.IsIncorrectlyGuarded(&transaction.Transaction{SndAddr: []byte("alice")})
			require.True(t, result)
		})

		t.Run("arbitrary processing error", func(t *testing.T) {
			t.Parallel()

			processor := &testscommon.TxProcessorStub{
				VerifyGuardianCalled: func(_ *transaction.Transaction, _ state.UserAccountHandler) error {
					return fmt.Errorf("arbitrary processing error")
				},
			}

			session, err := NewSelectionSession(ArgsSelectionSession{
				AccountsAdapter:       accounts,
				TransactionsProcessor: processor,
				TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{
					IsGuardedTransactionCalled: func(_ *transaction.Transaction) bool { return true },
				},
			})
			require.NoError(t, err)

			result := session.IsIncorrectlyGuarded(&transaction.Transaction{SndAddr: []byte("alice")})
			require.False(t, result)
		})
	})

	t.Run("account fetch error should return false", func(t *testing.T) {
		t.Parallel()

		accounts := &stateMock.AccountsStub{
			GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				return nil, fmt.Errorf("unexpected trie error")
			},
		}

		session, err := NewSelectionSession(ArgsSelectionSession{
			AccountsAdapter:         accounts,
			TransactionsProcessor:   &testscommon.TxProcessorStub{},
			TxVersionCheckerHandler: &testscommon.TxVersionCheckerStub{},
		})
		require.NoError(t, err)

		result := session.IsIncorrectlyGuarded(&transaction.Transaction{SndAddr: []byte("alice")})
		require.False(t, result)
	})
}
