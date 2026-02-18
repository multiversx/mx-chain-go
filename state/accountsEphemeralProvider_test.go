package state_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/state"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewAccountsEphemeralProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil accounts adapter should error", func(t *testing.T) {
		t.Parallel()

		provider, err := state.NewAccountsEphemeralProvider(nil)
		require.Error(t, err, state.ErrNilAccountsAdapter)
		require.Nil(t, provider)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		provider, err := state.NewAccountsEphemeralProvider(&stateMock.AccountsStub{})
		require.NoError(t, err)
		require.NotNil(t, provider)
	})
}

func TestAccountsEphemeralProvider_GetRootHash(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return []byte("abba"), nil
		},
	}

	provider, err := state.NewAccountsEphemeralProvider(accounts)
	require.NoError(t, err)
	require.NotNil(t, provider)

	rootHash, err := provider.GetRootHash()
	require.Nil(t, err)
	require.Equal(t, []byte("abba"), rootHash)
}

func TestAccountsEphemeralProvider_GetAccountNonce(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}

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
			}, nil
		}

		if bytes.Equal(address, []byte("carol")) {
			return nil, state.ErrAccNotFound
		}

		return nil, errors.New("arbitrary error")
	}

	provider, err := state.NewAccountsEphemeralProvider(accounts)
	require.NoError(t, err)
	require.NotNil(t, provider)

	nonce, exists, err := provider.GetAccountNonce([]byte("alice"))
	require.NoError(t, err)
	require.Equal(t, uint64(42), nonce)
	require.True(t, exists)

	nonce, exists, err = provider.GetAccountNonce([]byte("bob"))
	require.NoError(t, err)
	require.Equal(t, uint64(7), nonce)
	require.True(t, exists)

	// If account is not found, no error is returned.
	nonce, exists, err = provider.GetAccountNonce([]byte("carol"))
	require.NoError(t, err)
	require.Equal(t, uint64(0), nonce)
	require.False(t, exists)

	nonce, exists, err = provider.GetAccountNonce([]byte("judy"))
	require.ErrorContains(t, err, "arbitrary error")
	require.Equal(t, uint64(0), nonce)
	require.False(t, exists)
}

func TestAccountsEphemeralProvider_GetAccountNonceAndBalance(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}

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

		if bytes.Equal(address, []byte("carol")) {
			return nil, state.ErrAccNotFound
		}

		return nil, errors.New("arbitrary error")
	}

	provider, err := state.NewAccountsEphemeralProvider(accounts)
	require.NoError(t, err)
	require.NotNil(t, provider)

	nonce, balance, exists, err := provider.GetAccountNonceAndBalance([]byte("alice"))
	require.NoError(t, err)
	require.Equal(t, uint64(42), nonce)
	require.Equal(t, "3000000000000000000", balance.String())
	require.True(t, exists)

	nonce, balance, exists, err = provider.GetAccountNonceAndBalance([]byte("bob"))
	require.NoError(t, err)
	require.Equal(t, uint64(7), nonce)
	require.Equal(t, "1000000000000000000", balance.String())
	require.True(t, exists)

	// If account is not found, no error is returned.
	nonce, balance, exists, err = provider.GetAccountNonceAndBalance([]byte("carol"))
	require.NoError(t, err)
	require.Equal(t, uint64(0), nonce)
	require.Equal(t, "0", balance.String())
	require.False(t, exists)

	nonce, balance, exists, err = provider.GetAccountNonceAndBalance([]byte("judy"))
	require.ErrorContains(t, err, "arbitrary error")
	require.Equal(t, uint64(0), nonce)
	require.Nil(t, balance)
	require.False(t, exists)
}

func TestAccountsEphemeralProvider_GetUserAccount(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}

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

		if bytes.Equal(address, []byte("carol")) {
			return nil, state.ErrAccNotFound
		}

		return nil, errors.New("arbitrary error")
	}

	provider, err := state.NewAccountsEphemeralProvider(accounts)
	require.NoError(t, err)
	require.NotNil(t, provider)

	account, err := provider.GetUserAccount([]byte("alice"))
	require.NoError(t, err)
	require.Equal(t, uint64(42), account.GetNonce())
	require.Equal(t, "3000000000000000000", account.GetBalance().String())

	account, err = provider.GetUserAccount([]byte("bob"))
	require.NoError(t, err)
	require.Equal(t, uint64(7), account.GetNonce())
	require.Equal(t, "1000000000000000000", account.GetBalance().String())

	// If account is not found, no error is returned.
	account, err = provider.GetUserAccount([]byte("carol"))
	require.NoError(t, err)
	require.Nil(t, account)

	account, err = provider.GetUserAccount([]byte("judy"))
	require.ErrorContains(t, err, "arbitrary error")
	require.Nil(t, account)
}

func TestAccountsEphemeralProvider_GetUserAccount_cacheIsSharedAmongCalls(t *testing.T) {
	t.Parallel()

	accounts := &stateMock.AccountsStub{}

	numCallsGetExistingAccount := 0

	accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		numCallsGetExistingAccount++

		if bytes.Equal(address, []byte("carol")) {
			// Missing (not found) accounts should be cached, as well.
			return nil, state.ErrAccNotFound
		}

		return &stateMock.UserAccountStub{Nonce: 7, Balance: big.NewInt(42)}, nil
	}

	provider, err := state.NewAccountsEphemeralProvider(accounts)
	require.NoError(t, err)
	require.NotNil(t, provider)

	account, err := provider.GetUserAccount([]byte("alice"))
	require.NotNil(t, account)
	require.Nil(t, err)
	require.Equal(t, 1, numCallsGetExistingAccount)

	account, err = provider.GetUserAccount([]byte("alice"))
	require.NotNil(t, account)
	require.Nil(t, err)
	require.Equal(t, 1, numCallsGetExistingAccount)

	nonce, balance, exists, err := provider.GetAccountNonceAndBalance([]byte("alice"))
	require.Equal(t, uint64(7), nonce)
	require.Equal(t, uint64(42), balance.Uint64())
	require.True(t, exists)
	require.Nil(t, err)
	require.Equal(t, 1, numCallsGetExistingAccount)

	account, err = provider.GetUserAccount([]byte("bob"))
	require.NotNil(t, account)
	require.Nil(t, err)
	require.Equal(t, 2, numCallsGetExistingAccount)

	account, err = provider.GetUserAccount([]byte("bob"))
	require.NotNil(t, account)
	require.Nil(t, err)
	require.Equal(t, 2, numCallsGetExistingAccount)

	nonce, balance, exists, err = provider.GetAccountNonceAndBalance([]byte("bob"))
	require.Equal(t, uint64(7), nonce)
	require.Equal(t, uint64(42), balance.Uint64())
	require.True(t, exists)
	require.Nil(t, err)
	require.Equal(t, 2, numCallsGetExistingAccount)

	// Missing (not found) accounts are cached, as well.
	account, err = provider.GetUserAccount([]byte("carol"))
	require.Nil(t, account)
	require.Nil(t, err)
	require.Equal(t, 3, numCallsGetExistingAccount)

	account, err = provider.GetUserAccount([]byte("carol"))
	require.Nil(t, account)
	require.Nil(t, err)
	require.Equal(t, 3, numCallsGetExistingAccount)

	nonce, balance, exists, err = provider.GetAccountNonceAndBalance([]byte("carol"))
	require.Equal(t, uint64(0), nonce)
	require.Equal(t, uint64(0), balance.Uint64())
	require.False(t, exists)
	require.Nil(t, err)
	require.Equal(t, 3, numCallsGetExistingAccount)
}

func TestAccountsEphemeralProvider_GetUserAccount_UpgradesLightAccount(t *testing.T) {
	t.Parallel()

	t.Run("light account in cache triggers upgrade via GetUserAccount", func(t *testing.T) {
		t.Parallel()

		getExistingCalls := 0

		// First call returns a UserAccountStub (simulating fetchAndCacheAccount for the light upgrade fallback).
		accounts := &stateMock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				getExistingCalls++
				return &stateMock.UserAccountStub{
					Address: address,
					Nonce:   10,
					Balance: big.NewInt(500),
				}, nil
			},
		}

		provider, err := state.NewAccountsEphemeralProvider(accounts)
		require.NoError(t, err)

		// Warm cache with GetAccountNonceAndBalance (uses getAccountForNonceAndBalance → fetchAndCacheAccount)
		nonce, balance, exists, err := provider.GetAccountNonceAndBalance([]byte("alice"))
		require.NoError(t, err)
		require.True(t, exists)
		require.Equal(t, uint64(10), nonce)
		require.Equal(t, big.NewInt(500), balance)
		require.Equal(t, 1, getExistingCalls) // One trie read

		// GetUserAccount should return the cached full account (no upgrade needed since it's already full)
		account, err := provider.GetUserAccount([]byte("alice"))
		require.NoError(t, err)
		require.NotNil(t, account)
		require.Equal(t, 1, getExistingCalls) // No additional trie read — cache hit
	})
}

func TestAccountsEphemeralProvider_IsAccountGuarded(t *testing.T) {
	t.Parallel()

	t.Run("non-existent account returns false", func(t *testing.T) {
		t.Parallel()

		accounts := &stateMock.AccountsStub{
			GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				return nil, state.ErrAccNotFound
			},
		}

		provider, err := state.NewAccountsEphemeralProvider(accounts)
		require.NoError(t, err)

		isGuarded, err := provider.IsAccountGuarded([]byte("unknown"))
		require.NoError(t, err)
		require.False(t, isGuarded)
	})

	t.Run("light account returns correct guarded status without upgrade", func(t *testing.T) {
		t.Parallel()

		numGetExistingAccountCalls := 0

		accounts := &stateMock.AccountsStub{
			GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				numGetExistingAccountCalls++
				return &stateMock.UserAccountStub{
					IsGuardedCalled: func() bool { return false },
				}, nil
			},
		}

		provider, err := state.NewAccountsEphemeralProvider(accounts)
		require.NoError(t, err)

		// Warm cache via GetAccountNonceAndBalance (simulates proposal flow)
		_, _, _, err = provider.GetAccountNonceAndBalance([]byte("alice"))
		require.NoError(t, err)
		require.Equal(t, 1, numGetExistingAccountCalls)

		// IsAccountGuarded should use the cached account, no new trie read
		isGuarded, err := provider.IsAccountGuarded([]byte("alice"))
		require.NoError(t, err)
		require.False(t, isGuarded)
		require.Equal(t, 1, numGetExistingAccountCalls)
	})

	t.Run("guarded full account returns true", func(t *testing.T) {
		t.Parallel()

		accounts := &stateMock.AccountsStub{
			GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.UserAccountStub{
					IsGuardedCalled: func() bool { return true },
				}, nil
			},
		}

		provider, err := state.NewAccountsEphemeralProvider(accounts)
		require.NoError(t, err)

		isGuarded, err := provider.IsAccountGuarded([]byte("bob"))
		require.NoError(t, err)
		require.True(t, isGuarded)
	})

	t.Run("error from underlying adapter is propagated", func(t *testing.T) {
		t.Parallel()

		accounts := &stateMock.AccountsStub{
			GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				return nil, errors.New("trie error")
			},
		}

		provider, err := state.NewAccountsEphemeralProvider(accounts)
		require.NoError(t, err)

		isGuarded, err := provider.IsAccountGuarded([]byte("fail"))
		require.Error(t, err)
		require.False(t, isGuarded)
	})
}
