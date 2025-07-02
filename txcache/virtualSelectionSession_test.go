package txcache

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/state"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_newVirtualSelectionSession(t *testing.T) {
	t.Parallel()

	session := txcachemocks.NewSelectionSessionMock()
	virtualSession := newVirtualSelectionSession(session)
	require.NotNil(t, virtualSession)
}

func Test_getVirtualRecord(t *testing.T) {
	t.Parallel()

	t.Run("should return virtual record", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{}
		virtualSession := newVirtualSelectionSession(&sessionMock)

		expectedRecord := virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			initialBalance:  big.NewInt(2),
			consumedBalance: big.NewInt(3),
		}
		virtualSession.virtualAccountsByAddress = map[string]*virtualAccountRecord{
			"alice": &expectedRecord,
		}

		actualRecord, err := virtualSession.getVirtualRecord([]byte("alice"))
		require.NoError(t, err)
		require.Equal(t, &expectedRecord, actualRecord)
	})

	t.Run("should return account from real session", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
				return &stateMock.StateUserAccountHandlerStub{
					GetBalanceCalled: func() *big.Int {
						return big.NewInt(2)
					},
					GetNonceCalled: func() uint64 {
						return 2
					},
				}, nil
			},
		}
		virtualSession := newVirtualSelectionSession(&sessionMock)

		expectedRecord := virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			initialBalance:  big.NewInt(2),
			consumedBalance: big.NewInt(0),
		}
		actualRecord, err := virtualSession.getVirtualRecord([]byte("alice"))

		require.NoError(t, err)
		require.Equal(t, expectedRecord.initialNonce, actualRecord.initialNonce)
		require.Equal(t, expectedRecord.initialBalance, actualRecord.initialBalance)
		require.Equal(t, expectedRecord.consumedBalance, actualRecord.consumedBalance)
	})

	t.Run("should err", func(t *testing.T) {
		t.Parallel()

		expErr := errors.New("error")
		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
				return nil, expErr
			},
		}
		virtualSession := newVirtualSelectionSession(&sessionMock)

		actualRecord, err := virtualSession.getVirtualRecord([]byte("alice"))
		require.Nil(t, actualRecord)
		require.Equal(t, expErr, err)

	})
}

func Test_getNonce(t *testing.T) {
	t.Parallel()

	t.Run("should return nonce from real session", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
				return &stateMock.StateUserAccountHandlerStub{
					GetBalanceCalled: func() *big.Int {
						return big.NewInt(2)
					},
					GetNonceCalled: func() uint64 {
						return 2
					},
				}, nil
			},
		}
		virtualSession := newVirtualSelectionSession(&sessionMock)
		actualNonce, err := virtualSession.getNonce([]byte("alice"))

		require.NoError(t, err)
		require.Equal(t, uint64(2), actualNonce)
	})

	t.Run("should return nonce from account record", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
				return &stateMock.StateUserAccountHandlerStub{
					GetBalanceCalled: func() *big.Int {
						return big.NewInt(2)
					},
					GetNonceCalled: func() uint64 {
						return 2
					},
				}, nil
			},
		}

		virtualSession := newVirtualSelectionSession(&sessionMock)

		expectedRecord := virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			initialBalance:  big.NewInt(2),
			consumedBalance: big.NewInt(3),
		}
		virtualSession.virtualAccountsByAddress = map[string]*virtualAccountRecord{
			"alice": &expectedRecord,
		}

		actualNonce, err := virtualSession.getNonce([]byte("alice"))

		require.NoError(t, err)
		require.Equal(t, uint64(3), actualNonce)
	})

	t.Run("should err", func(t *testing.T) {
		t.Parallel()

		expErr := errors.New("error")
		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
				return nil, expErr
			},
		}
		virtualSession := newVirtualSelectionSession(&sessionMock)

		_, err := virtualSession.getNonce([]byte("alice"))
		require.Equal(t, expErr, err)
	})
}
