package txcache

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

// TODO: check if this test makes sense.
func Test_handleAccountBreadcrumb(t *testing.T) {
	t.Parallel()

	address := "bob"
	accountBalance := big.NewInt(2)

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
		virtualBalance: &virtualAccountBalance{
			initialBalance:  big.NewInt(2),
			consumedBalance: big.NewInt(3),
		},
	}

	sessionMock := txcachemocks.SelectionSessionMock{
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			return 2, big.NewInt(2), true, nil
		},
	}

	provider := newVirtualSessionProvider(&sessionMock)
	provider.handleAccountBreadcrumb(&breadcrumbBob, accountBalance, address)

	actualVirtualRecord, ok := provider.virtualAccountsByAddress[address]
	require.True(t, ok)
	require.Equal(t, expectedVirtualRecord, actualVirtualRecord)
}

func Test_createVirtualSelectionSession(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 2, big.NewInt(2), true, nil
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
				virtualBalance: &virtualAccountBalance{
					initialBalance:  big.NewInt(2),
					consumedBalance: big.NewInt(2),
				},
			},
			"bob": {
				initialNonce: core.OptionalUint64{
					Value:    6,
					HasValue: true,
				},
				virtualBalance: &virtualAccountBalance{
					initialBalance:  big.NewInt(2),
					consumedBalance: big.NewInt(6),
				},
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
		tb, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		require.NoError(t, err)
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
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, nil, false, expErr
			},
		}

		provider := newVirtualSessionProvider(&sessionMock)
		err = provider.handleTrackedBlock(tb)
		require.Equal(t, expErr, err)
	})

	t.Run("should skip alice", func(t *testing.T) {
		t.Parallel()

		tb, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		require.NoError(t, err)
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
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 2, big.NewInt(2), true, nil
			},
		}

		skippedSenders := map[string]struct{}{
			"alice": {},
		}

		provider := newVirtualSessionProvider(&sessionMock)
		provider.validator.skippedSenders = skippedSenders

		err = provider.handleTrackedBlock(tb)
		require.Nil(t, err)
		require.Equal(t, 1, len(provider.validator.sendersInContinuityWithSessionNonce))
		require.Equal(t, 1, len(provider.validator.accountPreviousBreadcrumb))

		virtualRecord, ok := provider.virtualAccountsByAddress["bob"]
		require.True(t, ok)
		require.Equal(t, core.OptionalUint64{Value: 4, HasValue: true}, virtualRecord.initialNonce)
		require.Equal(t, big.NewInt(2), virtualRecord.getInitialBalance())
		require.Equal(t, big.NewInt(3), virtualRecord.getConsumedBalance())

		_, ok = provider.virtualAccountsByAddress["alice"]
		require.False(t, ok)
	})

	t.Run("should delete bob and add it to skipped senders", func(t *testing.T) {
		t.Parallel()

		tb, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		require.NoError(t, err)
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
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 2, big.NewInt(2), true, nil
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
				virtualBalance: newVirtualAccountBalance(big.NewInt(5)),
			},
		}

		provider := newVirtualSessionProvider(&sessionMock)
		provider.validator.accountPreviousBreadcrumb = accountPreviousBreadcrumb
		provider.virtualAccountsByAddress = virtualAccountsByAddress

		_, ok := provider.virtualAccountsByAddress["bob"]
		require.True(t, ok)

		_, ok = provider.validator.skippedSenders["bob"]
		require.False(t, ok)

		err = provider.handleTrackedBlock(tb)
		require.Nil(t, err)

		_, ok = provider.virtualAccountsByAddress["bob"]
		require.False(t, ok)

		_, ok = provider.validator.skippedSenders["bob"]
		require.True(t, ok)
	})
}
