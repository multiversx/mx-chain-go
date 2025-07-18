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

		trackedBlock1 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		trackedBlock2 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash2"), []byte("blockPrevHash1"), nil)
		equalBlocks := trackedBlock1.sameNonce(trackedBlock2)
		require.True(t, equalBlocks)
	})

	t.Run("different nonce", func(t *testing.T) {
		t.Parallel()

		trackedBlock1 := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		trackedBlock2 := newTrackedBlock(1, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		equalBlocks := trackedBlock1.sameNonce(trackedBlock2)
		require.False(t, equalBlocks)
	})
}

func TestTrackedBlock_getBreadcrumb(t *testing.T) {
	t.Parallel()

	t.Run("should return new breadcrumb", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    0,
				HasValue: true,
			}, core.OptionalUint64{
				Value:    0,
				HasValue: true,
			}, nil),
		}

		nonce := core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}
		expectedBreadcrumb := newAccountBreadcrumb(nonce, nonce, big.NewInt(0))

		breadcrumb := block.getOrCreateBreadcrumbWithNonce("bob", nonce)
		require.Equal(t, expectedBreadcrumb, breadcrumb)
	})

	t.Run("should return existing breadcrumb", func(t *testing.T) {
		t.Parallel()

		nonce := core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}
		expectedBreadcrumb := newAccountBreadcrumb(nonce, nonce, big.NewInt(1))

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": expectedBreadcrumb,
		}

		breadcrumb := block.getOrCreateBreadcrumbWithNonce("alice", nonce)
		require.Equal(t, expectedBreadcrumb, breadcrumb)
	})
}

func TestTrackedBlock_compileBreadcrumb(t *testing.T) {
	t.Parallel()

	t.Run("sender does not exist in map", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)

		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   1,
				},
				TransferredValue: big.NewInt(5),
			},
		}

		block.compileBreadcrumbs(txs)
		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(
				core.OptionalUint64{
					Value:    1,
					HasValue: true,
				},
				core.OptionalUint64{
					Value:    1,
					HasValue: true,
				},
				big.NewInt(5),
			),
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			requireEqualBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})

	t.Run("sender exists in map, nil fee payer", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, big.NewInt(5)),
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

		block.compileBreadcrumbs(txs)
		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(
				core.OptionalUint64{
					Value:    1,
					HasValue: true,
				},
				core.OptionalUint64{
					Value:    4,
					HasValue: true,
				},
				big.NewInt(10),
			),
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			requireEqualBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})

	t.Run("sender exists in map, fee payer does not", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, big.NewInt(5)),
		}
		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("alice"),
					Nonce:   4,
				},
				TransferredValue: big.NewInt(5),
				FeePayer:         []byte("bob"),
			},
		}

		block.compileBreadcrumbs(txs)
		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, core.OptionalUint64{
				Value:    4,
				HasValue: true,
			}, big.NewInt(10)),
			"bob": newAccountBreadcrumb(core.OptionalUint64{
				Value:    0,
				HasValue: false,
			}, core.OptionalUint64{
				Value:    0,
				HasValue: false,
			}, big.NewInt(0)),
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			requireEqualBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})

	t.Run("sender and fee payer existing in map", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"), nil)
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, big.NewInt(5)),
			"bob": newAccountBreadcrumb(core.OptionalUint64{
				Value:    0,
				HasValue: true,
			}, core.OptionalUint64{
				Value:    0,
				HasValue: true,
			}, big.NewInt(3)),
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

		block.compileBreadcrumbs(txs)
		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"alice": newAccountBreadcrumb(core.OptionalUint64{
				Value:    1,
				HasValue: true,
			}, core.OptionalUint64{
				Value:    4,
				HasValue: true,
			}, big.NewInt(15)),
			"bob": newAccountBreadcrumb(core.OptionalUint64{
				Value:    0,
				HasValue: true,
			}, core.OptionalUint64{
				Value:    0,
				HasValue: true,
			}, big.NewInt(6)),
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			requireEqualBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})
}
