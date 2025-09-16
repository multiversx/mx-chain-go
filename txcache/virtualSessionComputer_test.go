package txcache

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_fromBreadcrumbToVirtualRecord(t *testing.T) {
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

	computer := newVirtualSessionComputer(nil)
	computer.fromBreadcrumbToVirtualRecord(address, accountBalance, &breadcrumbBob)

	actualVirtualRecord, ok := computer.virtualAccountsByAddress[address]
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

		computer := newVirtualSessionComputer(&sessionMock)
		virtualSession, err := computer.createVirtualSelectionSession(trackedBlocks)
		require.Nil(t, err)
		require.Equal(t, expectedVirtualAccounts, virtualSession.virtualAccountsByAddress)
	})

}

func Test_handleTrackedBlock(t *testing.T) {
	t.Parallel()

	t.Run("should err", func(t *testing.T) {
		t.Parallel()

		tb := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
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

		computer := newVirtualSessionComputer(&sessionMock)
		err := computer.handleTrackedBlock(tb)
		require.Equal(t, expErr, err)
	})

	t.Run("should skip alice", func(t *testing.T) {
		t.Parallel()

		tb := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
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

		computer := newVirtualSessionComputer(&sessionMock)
		computer.validator.skippedSenders = skippedSenders

		err := computer.handleTrackedBlock(tb)
		require.Nil(t, err)
		require.Equal(t, 1, len(computer.validator.sendersInContinuityWithSessionNonce))
		require.Equal(t, 1, len(computer.validator.accountPreviousBreadcrumb))

		virtualRecord, ok := computer.virtualAccountsByAddress["bob"]
		require.True(t, ok)
		require.Equal(t, core.OptionalUint64{Value: 4, HasValue: true}, virtualRecord.initialNonce)
		require.Equal(t, big.NewInt(2), virtualRecord.getInitialBalance())
		require.Equal(t, big.NewInt(3), virtualRecord.getConsumedBalance())

		_, ok = computer.virtualAccountsByAddress["alice"]
		require.False(t, ok)
	})

	t.Run("should delete bob and add it to skipped senders", func(t *testing.T) {
		t.Parallel()

		tb := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))

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

		bobVirtualBalance, err := newVirtualAccountBalance(big.NewInt(5))
		require.Nil(t, err)

		virtualAccountsByAddress := map[string]*virtualAccountRecord{
			"bob": {
				initialNonce: core.OptionalUint64{
					Value:    6,
					HasValue: true,
				},
				virtualBalance: bobVirtualBalance,
			},
		}

		computer := newVirtualSessionComputer(&sessionMock)
		computer.validator.accountPreviousBreadcrumb = accountPreviousBreadcrumb
		computer.virtualAccountsByAddress = virtualAccountsByAddress

		_, ok := computer.virtualAccountsByAddress["bob"]
		require.True(t, ok)

		_, ok = computer.validator.skippedSenders["bob"]
		require.False(t, ok)

		err = computer.handleTrackedBlock(tb)
		require.Nil(t, err)

		_, ok = computer.virtualAccountsByAddress["bob"]
		require.False(t, ok)

		_, ok = computer.validator.skippedSenders["bob"]
		require.True(t, ok)
	})
}
