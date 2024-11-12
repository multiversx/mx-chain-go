package state

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestAccountNonceProvider_SetAccountsAdapter(t *testing.T) {
	t.Parallel()

	t.Run("with a nil the accounts adapter", func(t *testing.T) {
		t.Parallel()

		provider, err := NewAccountNonceProvider(nil)
		require.NoError(t, err)
		require.NotNil(t, provider)

		err = provider.SetAccountsAdapter(nil)
		require.ErrorIs(t, err, errors.ErrNilAccountsAdapter)
	})

	t.Run("with a non-nil accounts adapter", func(t *testing.T) {
		t.Parallel()

		provider, err := NewAccountNonceProvider(nil)
		require.NoError(t, err)
		require.NotNil(t, provider)

		err = provider.SetAccountsAdapter(&state.AccountsStub{})
		require.NoError(t, err)
	})
}

func TestAccountNonceProvider_GetAccountNonce(t *testing.T) {
	t.Parallel()

	t.Run("without a backing the accounts adapter", func(t *testing.T) {
		t.Parallel()

		provider, err := NewAccountNonceProvider(nil)
		require.NoError(t, err)
		require.NotNil(t, provider)

		nonce, err := provider.GetAccountNonce(nil)
		require.ErrorIs(t, err, errors.ErrNilAccountsAdapter)
		require.Equal(t, uint64(0), nonce)
	})

	t.Run("with a backing accounts adapter (provided in constructor)", func(t *testing.T) {
		t.Parallel()

		userAddress := []byte("alice")
		accounts := &state.AccountsStub{}
		accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
			if bytes.Equal(address, userAddress) {
				return nil, fmt.Errorf("account not found: %s", address)
			}

			return &state.UserAccountStub{
				Nonce: 42,
			}, nil
		}

		provider, err := NewAccountNonceProvider(accounts)
		require.NoError(t, err)
		require.NotNil(t, provider)

		nonce, err := provider.GetAccountNonce(userAddress)
		require.NoError(t, err)
		require.Equal(t, uint64(42), nonce)

		nonce, err = provider.GetAccountNonce([]byte("bob"))
		require.ErrorContains(t, err, "account not found: bob")
		require.Equal(t, uint64(0), nonce)
	})

	t.Run("with a backing accounts adapter (provided using setter)", func(t *testing.T) {
		t.Parallel()

		userAddress := []byte("alice")
		accounts := &state.AccountsStub{}
		accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
			if bytes.Equal(address, userAddress) {
				return nil, fmt.Errorf("account not found: %s", address)
			}

			return &state.UserAccountStub{
				Nonce: 42,
			}, nil
		}

		provider, err := NewAccountNonceProvider(nil)
		require.NoError(t, err)
		require.NotNil(t, provider)

		err = provider.SetAccountsAdapter(accounts)
		require.NoError(t, err)

		nonce, err := provider.GetAccountNonce(userAddress)
		require.NoError(t, err)
		require.Equal(t, uint64(42), nonce)

		nonce, err = provider.GetAccountNonce([]byte("bob"))
		require.ErrorContains(t, err, "account not found: bob")
		require.Equal(t, uint64(0), nonce)
	})
}
