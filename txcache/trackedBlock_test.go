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

func TestTrackedBlock_sameNonce(t *testing.T) {
	t.Parallel()

	t.Run("same nonce and same prev hash", func(t *testing.T) {
		t.Parallel()

		trackedBlock1 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		trackedBlock2 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash2"), []byte("blockPrevHash1"))
		equalBlocks := trackedBlock1.sameNonce(trackedBlock2)
		require.True(t, equalBlocks)
	})

	t.Run("different nonce", func(t *testing.T) {
		t.Parallel()

		trackedBlock1 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		trackedBlock2 := newTrackedBlock(1, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		equalBlocks := trackedBlock1.sameNonce(trackedBlock2)
		require.False(t, equalBlocks)
	})
}

func TestTrackedBlock_createOrUpdateVirtualRecords(t *testing.T) {
	t.Parallel()

	t.Run("should err", func(t *testing.T) {
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
			GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
				return nil, expErr
			},
		}

		err := tb.createOrUpdateVirtualRecords(&sessionMock, nil, nil, nil, nil)
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

		skippedSenders := map[string]struct{}{
			"alice": {},
		}
		sendersInContinuityWithSessionNonce := map[string]struct{}{}
		accountPreviousBreadcrumb := map[string]*accountBreadcrumb{}
		virtualAccountsByAddress := map[string]*virtualAccountRecord{}

		err := tb.createOrUpdateVirtualRecords(&sessionMock, skippedSenders,
			sendersInContinuityWithSessionNonce, accountPreviousBreadcrumb, virtualAccountsByAddress)
		require.Nil(t, err)
		require.Equal(t, 1, len(sendersInContinuityWithSessionNonce))
		require.Equal(t, 1, len(accountPreviousBreadcrumb))

		virtualRecord, ok := virtualAccountsByAddress["bob"]
		require.True(t, ok)
		require.Equal(t, core.OptionalUint64{Value: 2, HasValue: true}, virtualRecord.initialNonce)
		require.Equal(t, big.NewInt(5), virtualRecord.initialBalance)

		_, ok = virtualAccountsByAddress["alice"]
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

		skippedSenders := map[string]struct{}{}
		sendersInContinuityWithSessionNonce := map[string]struct{}{}
		accountPreviousBreadcrumb := map[string]*accountBreadcrumb{
			"bob": &breadcrumb1,
		}
		virtualAccountsByAddress := map[string]*virtualAccountRecord{
			"bob": {
				initialNonce: core.OptionalUint64{
					Value:    2,
					HasValue: true,
				},
				initialBalance: big.NewInt(5),
			},
		}

		_, ok := virtualAccountsByAddress["bob"]
		require.True(t, ok)

		_, ok = skippedSenders["bob"]
		require.False(t, ok)

		err := tb.createOrUpdateVirtualRecords(&sessionMock, skippedSenders,
			sendersInContinuityWithSessionNonce, accountPreviousBreadcrumb, virtualAccountsByAddress)
		require.Nil(t, err)

		_, ok = virtualAccountsByAddress["bob"]
		require.False(t, ok)

		_, ok = skippedSenders["bob"]
		require.True(t, ok)
	})
}
