package preprocess

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewAccountStateProvider(t *testing.T) {
	t.Parallel()

	provider, err := newAccountStateProvider(nil, &guardianMocks.GuardedAccountHandlerStub{})
	require.Nil(t, provider)
	require.ErrorIs(t, err, process.ErrNilAccountsAdapter)

	provider, err = newAccountStateProvider(&state.AccountsStub{}, nil)
	require.Nil(t, provider)
	require.ErrorIs(t, err, process.ErrNilGuardianChecker)

	provider, err = newAccountStateProvider(&state.AccountsStub{}, &guardianMocks.GuardedAccountHandlerStub{})
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestAccountStateProvider_GetAccountState(t *testing.T) {
	t.Parallel()

	accounts := &state.AccountsStub{}
	accounts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		// Alice has no guardian
		if bytes.Equal(address, []byte("alice")) {
			return &state.UserAccountStub{
				Address: []byte("alice"),
				Nonce:   42,
			}, nil
		}

		// Bob has Heidi as guardian
		if bytes.Equal(address, []byte("bob")) {
			return &state.UserAccountStub{
				Address: []byte("bob"),
				Nonce:   7,
				IsGuardedCalled: func() bool {
					return true
				},
			}, nil
		}

		return nil, fmt.Errorf("account not found: %s", address)
	}

	guardianChecker := &guardianMocks.GuardedAccountHandlerStub{}
	guardianChecker.GetActiveGuardianCalled = func(userAccount vmcommon.UserAccountHandler) ([]byte, error) {
		if bytes.Equal(userAccount.AddressBytes(), []byte("bob")) {
			return []byte("heidi"), nil
		}

		return nil, nil
	}

	provider, err := newAccountStateProvider(accounts, guardianChecker)
	require.NoError(t, err)
	require.NotNil(t, provider)

	state, err := provider.GetAccountState([]byte("alice"))
	require.NoError(t, err)
	require.Equal(t, uint64(42), state.Nonce)
	require.Nil(t, state.Guardian)

	state, err = provider.GetAccountState([]byte("bob"))
	require.NoError(t, err)
	require.Equal(t, uint64(7), state.Nonce)
	require.Equal(t, []byte("heidi"), state.Guardian)

	state, err = provider.GetAccountState([]byte("carol"))
	require.ErrorContains(t, err, "account not found: carol")
	require.Nil(t, state)
}
