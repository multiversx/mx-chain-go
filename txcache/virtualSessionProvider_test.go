package txcache

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/state"
	testscommonState "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_handleAccountBreadcrumb(t *testing.T) {
	t.Parallel()

	userAccountMock := testscommonState.StateUserAccountHandlerStub{
		GetBalanceCalled: func() *big.Int {
			return big.NewInt(2)
		},
	}
	address := "bob"

	breadcrumbBob := accountBreadcrumb{
		initialNonce: core.OptionalUint64{
			Value:    1,
			HasValue: true,
		},
		lastNonce: core.OptionalUint64{
			Value:    2,
			HasValue: true,
		},
		consumedBalance: big.NewInt(3),
	}

	expectedVirtualRecord := &virtualAccountRecord{
		initialNonce: core.OptionalUint64{
			Value:    3,
			HasValue: true,
		},
		initialBalance:  big.NewInt(2),
		consumedBalance: big.NewInt(3),
	}

	sessionMock := txcachemocks.SelectionSessionMock{
		GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
			return &testscommonState.StateUserAccountHandlerStub{
				GetBalanceCalled: func() *big.Int {
					return big.NewInt(2)
				},
				GetNonceCalled: func() uint64 {
					return 2
				},
			}, nil
		},
	}
	provider := newVirtualSessionProvider(&sessionMock)
	provider.handleAccountBreadcrumb(&breadcrumbBob, &userAccountMock, address)

	actualVirtualRecord, ok := provider.virtualAccountsByAddress[address]
	require.True(t, ok)
	require.Equal(t, expectedVirtualRecord, actualVirtualRecord)
}

func Test_createVirtualSelectionSession(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
				return &testscommonState.StateUserAccountHandlerStub{
					GetBalanceCalled: func() *big.Int {
						return big.NewInt(2)
					},
					GetNonceCalled: func() uint64 {
						return 2
					},
				}, nil
			},
		}

		trackedBlocks := []*trackedBlock{
			{
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					"alice": {
						initialNonce: core.OptionalUint64{
							Value:    2,
							HasValue: true,
						},
						lastNonce: core.OptionalUint64{
							Value:    2,
							HasValue: true,
						},
						consumedBalance: big.NewInt(2),
					},
					"bob": {
						initialNonce: core.OptionalUint64{
							Value:    2,
							HasValue: true,
						},
						lastNonce: core.OptionalUint64{
							Value:    3,
							HasValue: true,
						},
						consumedBalance: big.NewInt(3),
					},
				},
			},
			{
				breadcrumbsByAddress: map[string]*accountBreadcrumb{
					// carol's virtual record will not be saved because the initialNonce is != session nonce
					"carol": {
						initialNonce: core.OptionalUint64{
							Value:    10,
							HasValue: true,
						},
						lastNonce: core.OptionalUint64{
							Value:    11,
							HasValue: true,
						},
						consumedBalance: big.NewInt(2),
					},
					"bob": {
						initialNonce: core.OptionalUint64{
							Value:    4,
							HasValue: true,
						},
						lastNonce: core.OptionalUint64{
							Value:    5,
							HasValue: true,
						},
						consumedBalance: big.NewInt(3),
					},
				},
			},
		}

		expectedVirtualAccounts := map[string]*virtualAccountRecord{
			"alice": {
				initialNonce: core.OptionalUint64{
					Value:    3,
					HasValue: true,
				},
				initialBalance:  big.NewInt(2),
				consumedBalance: big.NewInt(2),
			},
			"bob": {
				initialNonce: core.OptionalUint64{
					Value:    6,
					HasValue: true,
				},
				initialBalance:  big.NewInt(2),
				consumedBalance: big.NewInt(6),
			},
		}

		provider := newVirtualSessionProvider(&sessionMock)
		virtualSession, err := provider.createVirtualSelectionSession(trackedBlocks)
		require.Nil(t, err)
		require.Equal(t, expectedVirtualAccounts, virtualSession.virtualAccountsByAddress)
	})

}

func Test_handleTrackedBlock(t *testing.T) {
	t.Parallel()

	t.Run("should err", func(t *testing.T) {
		tb := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		tb.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": {
				initialNonce: core.OptionalUint64{
					Value:    1,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    1,
					HasValue: true,
				},
				consumedBalance: big.NewInt(2),
			},
			"bob": {
				initialNonce: core.OptionalUint64{
					Value:    2,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    3,
					HasValue: true,
				},
				consumedBalance: big.NewInt(3),
			},
		}

		expErr := errors.New("error")
		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
				return nil, expErr
			},
		}

		provider := newVirtualSessionProvider(&sessionMock)
		err := provider.handleTrackedBlock(tb)
		require.Equal(t, expErr, err)
	})

	t.Run("should skip alice", func(t *testing.T) {
		t.Parallel()

		tb := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		tb.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": {
				initialNonce: core.OptionalUint64{
					Value:    1,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    1,
					HasValue: true,
				},
				consumedBalance: big.NewInt(2),
			},
			"bob": {
				initialNonce: core.OptionalUint64{
					Value:    2,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    3,
					HasValue: true,
				},
				consumedBalance: big.NewInt(3),
			},
		}

		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
				return &testscommonState.StateUserAccountHandlerStub{
					GetBalanceCalled: func() *big.Int {
						return big.NewInt(2)
					},
					GetNonceCalled: func() uint64 {
						return 2
					},
				}, nil
			},
		}

		skippedSenders := map[string]struct{}{
			"alice": {},
		}

		provider := newVirtualSessionProvider(&sessionMock)
		provider.skippedSenders = skippedSenders

		err := provider.handleTrackedBlock(tb)
		require.Nil(t, err)
		require.Equal(t, 1, len(provider.sendersInContinuityWithSessionNonce))
		require.Equal(t, 1, len(provider.accountPreviousBreadcrumb))

		virtualRecord, ok := provider.virtualAccountsByAddress["bob"]
		require.True(t, ok)
		require.Equal(t, core.OptionalUint64{Value: 4, HasValue: true}, virtualRecord.initialNonce)
		require.Equal(t, big.NewInt(2), virtualRecord.initialBalance)
		require.Equal(t, big.NewInt(3), virtualRecord.consumedBalance)

		_, ok = provider.virtualAccountsByAddress["alice"]
		require.False(t, ok)
	})

	t.Run("should delete bob and add it to skipped senders", func(t *testing.T) {
		t.Parallel()

		tb := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		breadcrumb1 := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			consumedBalance: big.NewInt(3),
		}

		breadcrumb2 := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    5,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    5,
				HasValue: true,
			},
			consumedBalance: big.NewInt(3),
		}

		tb.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"bob": &breadcrumb2,
		}

		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
				return &testscommonState.StateUserAccountHandlerStub{
					GetBalanceCalled: func() *big.Int {
						return big.NewInt(2)
					},
					GetNonceCalled: func() uint64 {
						return 2
					},
				}, nil
			},
		}

		accountPreviousBreadcrumb := map[string]*accountBreadcrumb{
			"bob": &breadcrumb1,
		}
		virtualAccountsByAddress := map[string]*virtualAccountRecord{
			"bob": {
				initialNonce: core.OptionalUint64{
					Value:    6,
					HasValue: true,
				},
				initialBalance: big.NewInt(5),
			},
		}

		provider := newVirtualSessionProvider(&sessionMock)
		provider.accountPreviousBreadcrumb = accountPreviousBreadcrumb
		provider.virtualAccountsByAddress = virtualAccountsByAddress

		_, ok := provider.virtualAccountsByAddress["bob"]
		require.True(t, ok)

		_, ok = provider.skippedSenders["bob"]
		require.False(t, ok)

		err := provider.handleTrackedBlock(tb)
		require.Nil(t, err)

		_, ok = provider.virtualAccountsByAddress["bob"]
		require.False(t, ok)

		_, ok = provider.skippedSenders["bob"]
		require.True(t, ok)
	})
}

func Test_continousBreadcrumbs(t *testing.T) {
	t.Parallel()

	// when breadcrumb is relayer
	t.Run("relayer should be continuous", func(t *testing.T) {
		t.Parallel()

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: false,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: false,
			},
			consumedBalance: nil,
		}

		sessionMock := txcachemocks.SelectionSessionMock{}

		provider := newVirtualSessionProvider(&sessionMock)
		actualRes := provider.continuousBreadcrumb("bob", &breadcrumb, 3)
		require.True(t, actualRes)
	})

	// when certain account is sender for the first time in the chain of tracked blocks
	t.Run("sender not continuous with session nonce", func(t *testing.T) {
		t.Parallel()

		breadcrumbAlice := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		sessionMock := txcachemocks.SelectionSessionMock{}

		provider := newVirtualSessionProvider(&sessionMock)
		actualRes := provider.continuousBreadcrumb("alice", &breadcrumbAlice, 3)
		require.False(t, actualRes)
	})

	t.Run("sender continuous with session nonce", func(t *testing.T) {
		t.Parallel()

		breadcrumbAlice := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		sessionMock := txcachemocks.SelectionSessionMock{}
		provider := newVirtualSessionProvider(&sessionMock)

		actualRes := provider.continuousBreadcrumb("alice", &breadcrumbAlice, 1)
		require.True(t, actualRes)

		_, ok := provider.sendersInContinuityWithSessionNonce["alice"]
		require.True(t, ok)

		actualBreadcrumb, ok := provider.accountPreviousBreadcrumb["alice"]
		require.True(t, ok)
		require.Equal(t, &breadcrumbAlice, actualBreadcrumb)
	})

	// when address was already a sender in the chain of tracked blocks
	t.Run("sender continuous with previous account breadcrumb ", func(t *testing.T) {
		t.Parallel()

		breadcrumbAlice1 := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		breadcrumbAlice2 := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    4,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		sessionMock := txcachemocks.SelectionSessionMock{}
		provider := newVirtualSessionProvider(&sessionMock)

		provider.accountPreviousBreadcrumb = map[string]*accountBreadcrumb{
			"alice": &breadcrumbAlice1,
		}

		actualRes := provider.continuousBreadcrumb("alice", &breadcrumbAlice2, 3)
		require.True(t, actualRes)

		actualBreadcrumb, ok := provider.accountPreviousBreadcrumb["alice"]
		require.True(t, ok)
		require.Equal(t, &breadcrumbAlice2, actualBreadcrumb)
	})

	t.Run("sender is not continuous with previous account breadcrumb ", func(t *testing.T) {
		t.Parallel()

		breadcrumbAlice1 := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		breadcrumbAlice2 := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    4,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    4,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		sessionMock := txcachemocks.SelectionSessionMock{}
		provider := newVirtualSessionProvider(&sessionMock)

		provider.accountPreviousBreadcrumb = map[string]*accountBreadcrumb{
			"alice": &breadcrumbAlice1,
		}
		provider.sendersInContinuityWithSessionNonce = map[string]struct{}{
			"alice": {},
		}

		actualRes := provider.continuousBreadcrumb("alice", &breadcrumbAlice2, 1)
		require.False(t, actualRes)
	})
}
