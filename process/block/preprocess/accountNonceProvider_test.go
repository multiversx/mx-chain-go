package preprocess

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewAccountNonceProvider(t *testing.T) {
	t.Parallel()

	provider, err := newAccountNonceProvider(nil)
	require.Nil(t, provider)
	require.ErrorIs(t, err, errors.ErrNilAccountsAdapter)

	provider, err = newAccountNonceProvider(&state.AccountsStub{})
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestAccountNonceProvider_GetAccountNonce(t *testing.T) {
	t.Parallel()

	userAddress := []byte("alice")
	accounts := &state.AccountsStub{}
	accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if !bytes.Equal(address, userAddress) {
			return nil, fmt.Errorf("account not found: %s", address)
		}

		return &state.UserAccountStub{
			Nonce: 42,
		}, nil
	}

	provider, err := newAccountNonceProvider(accounts)
	require.NoError(t, err)
	require.NotNil(t, provider)

	nonce, err := provider.GetAccountNonce(userAddress)
	require.NoError(t, err)
	require.Equal(t, uint64(42), nonce)

	nonce, err = provider.GetAccountNonce([]byte("bob"))
	require.ErrorContains(t, err, "account not found: bob")
	require.Equal(t, uint64(0), nonce)
}
