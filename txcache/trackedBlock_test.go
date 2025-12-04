package txcache

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func requireEqualBreadcrumbs(t *testing.T, breadCrumb1 *accountBreadcrumb, breadCrumb2 *accountBreadcrumb) {
	require.Equal(t, breadCrumb1.firstNonce, breadCrumb2.firstNonce)
	require.Equal(t, breadCrumb1.lastNonce, breadCrumb2.lastNonce)
	require.Equal(t, breadCrumb1.consumedBalance, breadCrumb2.consumedBalance)
}

func TestTrackedBlock_sameNonceOrBelow(t *testing.T) {
	t.Parallel()

	t.Run("same nonce and same prev hash", func(t *testing.T) {
		t.Parallel()

		trackedBlock1 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		trackedBlock2 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash2"), []byte("blockPrevHash1"))

		shouldRemoveBlock := trackedBlock1.hasSameNonceOrLower(trackedBlock2)
		require.True(t, shouldRemoveBlock)
	})

	t.Run("lower nonce", func(t *testing.T) {
		t.Parallel()

		trackedBlock1 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		trackedBlock2 := newTrackedBlock(1, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))

		shouldRemoveBlock := trackedBlock1.hasSameNonceOrLower(trackedBlock2)
		require.True(t, shouldRemoveBlock)
	})

	t.Run("greater nonce", func(t *testing.T) {
		t.Parallel()

		trackedBlock1 := newTrackedBlock(2, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		trackedBlock2 := newTrackedBlock(1, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))

		shouldRemoveBlock := trackedBlock1.hasSameNonceOrLower(trackedBlock2)
		require.False(t, shouldRemoveBlock)
	})
}

func TestTrackedBlock_sameNonceOrHigher(t *testing.T) {
	t.Parallel()

	trackedBlock0 := newTrackedBlock(1, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
	trackedBlock1 := newTrackedBlock(2, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
	trackedBlock2 := newTrackedBlock(3, []byte("blockHash1"), []byte("blockRootHash2"), []byte("blockPrevHash1"))

	newProposedBlock := newTrackedBlock(2, []byte("blockHash1"), []byte("blockRootHash2"), []byte("blockPrevHash1"))

	shouldRemoveBlock := trackedBlock1.hasSameNonceOrHigher(newProposedBlock)
	require.True(t, shouldRemoveBlock)

	shouldRemoveBlock = trackedBlock2.hasSameNonceOrHigher(newProposedBlock)
	require.True(t, shouldRemoveBlock)

	shouldRemoveBlock = trackedBlock0.hasSameNonceOrHigher(newProposedBlock)
	require.False(t, shouldRemoveBlock)
}

func TestTrackedBlock_getBreadcrumb(t *testing.T) {
	t.Parallel()

	t.Run("should return new breadcrumb", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    0,
				HasValue: true,
			}),
		}

		nonce := core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}
		expectedBreadcrumb := newAccountBreadcrumb(nonce)

		breadcrumb := block.getOrCreateBreadcrumbWithNonce("bob", nonce)
		require.Equal(t, expectedBreadcrumb, breadcrumb)
	})

	t.Run("should return existing breadcrumb", func(t *testing.T) {
		t.Parallel()

		nonce := core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}
		expectedBreadcrumb := newAccountBreadcrumb(nonce)
		expectedBreadcrumb.accumulateConsumedBalance(big.NewInt(1))

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))

		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": expectedBreadcrumb,
		}

		breadcrumb := block.getOrCreateBreadcrumbWithNonce("alice", nonce)
		require.Equal(t, expectedBreadcrumb, breadcrumb)
	})
}

func TestTrackedBlock_compileBreadcrumbs(t *testing.T) {
	t.Parallel()

	t.Run("sender does not exist in map", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))

		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   1,
				},
				TransferredValue: big.NewInt(5),
			},
		}

		err := block.compileBreadcrumbs(txs)
		require.NoError(t, err)

		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{Value: 1, HasValue: true})
		aliceBreadcrumb.accumulateConsumedBalance(big.NewInt(5))
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

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{
			Value:    1,
			HasValue: true})

		aliceBreadcrumb.accumulateConsumedBalance(big.NewInt(5))
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": aliceBreadcrumb,
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

		err := block.compileBreadcrumbs(txs)
		require.NoError(t, err)

		aliceBreadcrumb = newAccountBreadcrumb(core.OptionalUint64{Value: 1, HasValue: true})
		aliceBreadcrumb.accumulateConsumedBalance(big.NewInt(10))
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

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{
			Value:    1,
			HasValue: true,
		})
		aliceBreadcrumb.accumulateConsumedBalance(big.NewInt(5))
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": aliceBreadcrumb,
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

		err := block.compileBreadcrumbs(txs)
		require.NoError(t, err)

		aliceBreadcrumb = newAccountBreadcrumb(core.OptionalUint64{Value: 1, HasValue: true})
		aliceBreadcrumb.accumulateConsumedBalance(big.NewInt(10))
		aliceBreadcrumb.lastNonce = core.OptionalUint64{Value: 4, HasValue: true}

		bobBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{Value: 0, HasValue: false})
		bobBreadcrumb.accumulateConsumedBalance(big.NewInt(5))

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

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{
			Value:    1,
			HasValue: true,
		})
		aliceBreadcrumb.accumulateConsumedBalance(big.NewInt(5))

		bobBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{
			Value:    0,
			HasValue: true,
		})
		bobBreadcrumb.accumulateConsumedBalance(big.NewInt(3))

		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": aliceBreadcrumb,
			"bob":   bobBreadcrumb,
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

		err := block.compileBreadcrumbs(txs)
		require.NoError(t, err)

		aliceBreadcrumb = newAccountBreadcrumb(core.OptionalUint64{Value: 1, HasValue: true})
		aliceBreadcrumb.accumulateConsumedBalance(big.NewInt(15))
		aliceBreadcrumb.lastNonce = core.OptionalUint64{Value: 4, HasValue: true}

		bobBreadcrumb = newAccountBreadcrumb(core.OptionalUint64{Value: 0, HasValue: true})
		bobBreadcrumb.accumulateConsumedBalance(big.NewInt(6))

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

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{
			Value:    1,
			HasValue: true,
		})
		aliceBreadcrumb.accumulateConsumedBalance(big.NewInt(5))
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": aliceBreadcrumb,
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

		err := block.compileBreadcrumbs(txs)
		require.NoError(t, err)

		aliceBreadcrumb = newAccountBreadcrumb(core.OptionalUint64{
			Value:    1,
			HasValue: true,
		})
		aliceBreadcrumb.accumulateConsumedBalance(big.NewInt(12)) // initial value in breadcrumb + transferredValue + fee
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

	t.Run("sender becomes fee payer, fee payer becomes sender", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))

		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   1,
				},
				FeePayer:         []byte("bob"),
				Fee:              big.NewInt(2),
				TransferredValue: big.NewInt(5),
			},
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("bob"),
					Nonce:   4,
				},
				FeePayer:         []byte("alice"),
				Fee:              big.NewInt(3),
				TransferredValue: big.NewInt(10),
			},
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   2,
				},
				TransferredValue: big.NewInt(10),
			},
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("bob"),
					Nonce:   5,
				},
				TransferredValue: big.NewInt(10),
			},
		}

		err := block.compileBreadcrumbs(txs)
		require.NoError(t, err)

		aliceBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{Value: 1, HasValue: true})
		aliceBreadcrumb.accumulateConsumedBalance(big.NewInt(18))
		aliceBreadcrumb.lastNonce = core.OptionalUint64{Value: 2, HasValue: true}

		bobBreadcrumb := newAccountBreadcrumb(core.OptionalUint64{Value: 4, HasValue: true})
		bobBreadcrumb.accumulateConsumedBalance(big.NewInt(22))
		bobBreadcrumb.lastNonce = core.OptionalUint64{Value: 5, HasValue: true}

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
}
