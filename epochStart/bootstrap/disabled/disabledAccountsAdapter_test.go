package disabled

import (
	"testing"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/state"
	testState "github.com/multiversx/mx-chain-go/testscommon/state"
)

func TestNewAccountsAdapter(t *testing.T) {
	t.Parallel()

	t.Run("nil input", func(t *testing.T) {
		accAdapter, err := NewAccountsAdapter(nil)
		require.Nil(t, accAdapter)
		require.Equal(t, state.ErrNilAccountFactory, err)
	})
	t.Run("should work", func(t *testing.T) {
		accAdapter, err := NewAccountsAdapter(&testState.AccountsFactoryStub{})
		require.Nil(t, err)
		require.False(t, accAdapter.IsInterfaceNil())
	})
}

func TestAccountsAdapter_LoadAccount_GetExistingAccount_GetAccountFromBytes(t *testing.T) {
	t.Parallel()

	accMock := &testState.AccountWrapMock{
		Address: []byte("addr"),
	}
	accFactory := &testState.AccountsFactoryStub{
		CreateAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			accMock.Address = address
			return accMock, nil
		},
	}
	accAdapter, _ := NewAccountsAdapter(accFactory)

	loadedAcc, err := accAdapter.LoadAccount([]byte("addr1"))
	require.Nil(t, err)
	require.Equal(t, accMock, loadedAcc)

	loadedAcc, err = accAdapter.GetExistingAccount([]byte("addr2"))
	require.Nil(t, err)
	require.Equal(t, accMock, loadedAcc)

	loadedAcc, err = accAdapter.GetAccountFromBytes([]byte("addr3"), nil)
	require.Nil(t, err)
	require.Equal(t, accMock, loadedAcc)
}
