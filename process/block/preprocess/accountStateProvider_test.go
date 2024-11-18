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

func TestNewAccountStateProvider(t *testing.T) {
	t.Parallel()

	provider, err := newAccountStateProvider(nil)
	require.Nil(t, provider)
	require.ErrorIs(t, err, errors.ErrNilAccountsAdapter)

	provider, err = newAccountStateProvider(&state.AccountsStub{})
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestAccountStateProvider_GetAccountState(t *testing.T) {
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

	provider, err := newAccountStateProvider(accounts)
	require.NoError(t, err)
	require.NotNil(t, provider)

	state, err := provider.GetAccountState(userAddress)
	require.NoError(t, err)
	require.Equal(t, uint64(42), state.Nonce)

	state, err = provider.GetAccountState([]byte("bob"))
	require.ErrorContains(t, err, "account not found: bob")
	require.Nil(t, state)
}
