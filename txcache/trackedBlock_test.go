package txcache

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func requireEqualBreadcrumbs(t *testing.T, breadCrumb1 *accountBreadcrumb, breadCrumb2 *accountBreadcrumb) {
	require.Equal(t, breadCrumb1.initialNonce, breadCrumb2.initialNonce)
	require.Equal(t, breadCrumb1.lastNonce, breadCrumb2.lastNonce)
	require.Equal(t, breadCrumb1.consumedBalance, breadCrumb2.consumedBalance)
}

func TestTrackedBlock_sameNonce(t *testing.T) {
	t.Parallel()

	t.Run("same nonce and same prev hash", func(t *testing.T) {
		t.Parallel()

		trackedBlock1, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil, nil)
		require.NoError(t, err)

		trackedBlock2, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash2"), []byte("blockPrevHash1"), nil, nil)
		require.NoError(t, err)

		equalBlocks := trackedBlock1.sameNonce(trackedBlock2)
		require.True(t, equalBlocks)
	})

	t.Run("different nonce", func(t *testing.T) {
		t.Parallel()

		trackedBlock1, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil, nil)
		require.NoError(t, err)

		trackedBlock2, err := newTrackedBlock(1, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil, nil)
		require.NoError(t, err)

		equalBlocks := trackedBlock1.sameNonce(trackedBlock2)
		require.False(t, equalBlocks)
	})
}

func TestTrackedBlock_getBreadcrumb(t *testing.T) {
	t.Parallel()

	t.Run("should return new breadcrumb", func(t *testing.T) {
		t.Parallel()

		block, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil, nil)
		require.NoError(t, err)

		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    0,
				HasValue: true,
			}, big.NewInt(10), big.NewInt(0)),
		}

		nonce := core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}
		expectedBreadcrumb := newAccountBreadcrumb(nonce, big.NewInt(10), big.NewInt(0))

		breadcrumb := block.getOrCreateBreadcrumbWithNonce("bob", nonce, big.NewInt(10))
		require.Equal(t, expectedBreadcrumb, breadcrumb)
	})

	t.Run("should return existing breadcrumb", func(t *testing.T) {
		t.Parallel()

		nonce := core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}
		expectedBreadcrumb := newAccountBreadcrumb(nonce, big.NewInt(10), big.NewInt(1))

		block, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil, nil)
		require.NoError(t, err)

		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": expectedBreadcrumb,
		}

		breadcrumb := block.getOrCreateBreadcrumbWithNonce("alice", nonce, big.NewInt(10))
		require.Equal(t, expectedBreadcrumb, breadcrumb)
	})
}

func TestTrackedBlock_compileBreadcrumb(t *testing.T) {
	t.Parallel()

	t.Run("sender does not exist in map", func(t *testing.T) {
		t.Parallel()

		block, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil, nil)
		require.NoError(t, err)

		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   1,
				},
				TransferredValue: big.NewInt(5),
			},
		}

		err = block.compileBreadcrumbs(txs, &defaultSelectionSessionMock)
		require.NoError(t, err)

		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{Value: 1, HasValue: true}, big.NewInt(20), big.NewInt(5))
		aliceBreadcrumb.lastNonce = core.OptionalUint64{Value: 1, HasValue: true}
		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"alice": aliceBreadcrumb,
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			requireEqualBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})

	t.Run("sender exists in map, nil fee payer", func(t *testing.T) {
		t.Parallel()

		block, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil, nil)
		require.NoError(t, err)

		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, big.NewInt(20), big.NewInt(5)),
		}

		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   4,
				},
				TransferredValue: big.NewInt(5),
			},
		}

		err = block.compileBreadcrumbs(txs, &defaultSelectionSessionMock)
		require.NoError(t, err)
		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{Value: 1, HasValue: true}, big.NewInt(20), big.NewInt(10))
		aliceBreadcrumb.lastNonce = core.OptionalUint64{Value: 4, HasValue: true}
		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"alice": aliceBreadcrumb,
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			requireEqualBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})

	t.Run("sender exists in map, fee payer does not", func(t *testing.T) {
		t.Parallel()

		block, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil, nil)
		require.NoError(t, err)

		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, big.NewInt(20), big.NewInt(5)),
		}
		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   4,
				},
				TransferredValue: big.NewInt(5),
				Fee:              big.NewInt(5),
				FeePayer:         []byte("bob"),
			},
		}

		err = block.compileBreadcrumbs(txs, &defaultSelectionSessionMock)
		require.NoError(t, err)

		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{Value: 1, HasValue: true}, big.NewInt(20), big.NewInt(10))
		aliceBreadcrumb.lastNonce = core.OptionalUint64{Value: 4, HasValue: true}

		bobBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{Value: 0, HasValue: false}, big.NewInt(20), big.NewInt(5))

		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"alice": aliceBreadcrumb,
			"bob":   bobBreadcrumb,
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			requireEqualBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})

	t.Run("sender and fee payer existing in map", func(t *testing.T) {
		t.Parallel()

		block, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil, nil)
		require.NoError(t, err)

		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, big.NewInt(20), big.NewInt(5)),
			"bob": newAccountBreadcrumb(core.OptionalUint64{
				Value:    0,
				HasValue: true,
			}, big.NewInt(20), big.NewInt(3)),
		}
		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   3,
				},
				TransferredValue: big.NewInt(5),
				FeePayer:         []byte("bob"),
			},
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   4,
				},
				TransferredValue: big.NewInt(5),
				FeePayer:         []byte("bob"),
				Fee:              big.NewInt(3),
			},
		}

		err = block.compileBreadcrumbs(txs, &defaultSelectionSessionMock)
		require.NoError(t, err)

		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{Value: 1, HasValue: true}, big.NewInt(20), big.NewInt(15))
		aliceBreadcrumb.lastNonce = core.OptionalUint64{Value: 4, HasValue: true}

		bobBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{Value: 0, HasValue: true}, big.NewInt(20), big.NewInt(6))

		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"alice": aliceBreadcrumb,
			"bob":   bobBreadcrumb,
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			requireEqualBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})

	t.Run("sender and fee payer are equal", func(t *testing.T) {
		t.Parallel()

		block, err := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		require.NoError(t, err)
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, big.NewInt(5)),
		}
		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   3,
				},
				TransferredValue: big.NewInt(5),
				FeePayer:         []byte("alice"),
				Fee:              big.NewInt(2),
			},
		}

		err = block.compileBreadcrumbs(txs)
		require.NoError(t, err)

		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}, big.NewInt(12)) // initial value in breadcrumb + transferredValue + fee
		aliceBreadcrumb.lastNonce = core.OptionalUint64{Value: 3, HasValue: true}

		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"alice": aliceBreadcrumb,
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			requireEqualBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}

	})
}
