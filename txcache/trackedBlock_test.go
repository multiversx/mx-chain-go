package txcache

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
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

func TestTrackedBlock_getBreadcrumb(t *testing.T) {
	t.Parallel()

	t.Run("should return new breadcrumb", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"addr1": {
				initialNonce:    0,
				lastNonce:       0,
				consumedBalance: nil,
			},
		}

		nonce := uint64(1)
		expectedBreadcrumb := accountBreadcrumb{
			initialNonce:    nonce,
			lastNonce:       nonce,
			consumedBalance: big.NewInt(0),
		}

		breadcrumb := block.getBreadcrumb("addr2", nonce)
		require.Equal(t, &expectedBreadcrumb, breadcrumb)
	})

	t.Run("should return breadcrumb", func(t *testing.T) {
		t.Parallel()

		nonce := uint64(1)
		expectedBreadcrumb := accountBreadcrumb{
			initialNonce:    nonce,
			lastNonce:       nonce,
			consumedBalance: big.NewInt(1),
		}

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"addr1": &expectedBreadcrumb,
		}

		breadcrumb := block.getBreadcrumb("addr1", nonce)
		require.Equal(t, &expectedBreadcrumb, breadcrumb)
	})
}

func equalBreadcrumbs(t *testing.T, breadCrumb1 *accountBreadcrumb, breadCrumb2 *accountBreadcrumb) {
	require.Equal(t, breadCrumb1.initialNonce, breadCrumb2.initialNonce)
	require.Equal(t, breadCrumb1.lastNonce, breadCrumb2.lastNonce)
	require.Equal(t, breadCrumb1.consumedBalance, breadCrumb2.consumedBalance)
}

func TestTrackedBlock_compileBreadcrumb(t *testing.T) {
	t.Parallel()

	t.Run("sender does not exist in map", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))

		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("sender1"),
					Nonce:   1,
				},
				TransferredValue: big.NewInt(5),
			},
		}

		block.compileBreadcrumbs(txs)
		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"sender1": {
				initialNonce:    1,
				lastNonce:       1,
				consumedBalance: big.NewInt(5),
			},
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			equalBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})

	t.Run("sender exists in map, nil fee payer", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"sender1": {
				initialNonce:    1,
				lastNonce:       1,
				consumedBalance: big.NewInt(5),
			},
		}

		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("sender1"),
					Nonce:   4,
				},
				TransferredValue: big.NewInt(5),
			},
		}

		block.compileBreadcrumbs(txs)
		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"sender1": {
				initialNonce:    1,
				lastNonce:       4,
				consumedBalance: big.NewInt(10),
			},
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			equalBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})

	t.Run("sender exists in map, fee payer does not", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"sender1": {
				initialNonce:    1,
				lastNonce:       1,
				consumedBalance: big.NewInt(5),
			},
		}
		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("sender1"),
					Nonce:   4,
				},
				TransferredValue: big.NewInt(5),
				FeePayer:         []byte("payer1"),
			},
		}

		block.compileBreadcrumbs(txs)
		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"sender1": {
				initialNonce:    1,
				lastNonce:       4,
				consumedBalance: big.NewInt(10),
			},
			"payer1": {
				initialNonce:    4,
				lastNonce:       4,
				consumedBalance: big.NewInt(0),
			},
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			equalBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})

	t.Run("sender and fee payer existing in map", func(t *testing.T) {
		t.Parallel()

		block := newTrackedBlock(0, []byte("blockHash1"), []byte("blockRootHash1"), []byte("blockPrevHash1"))
		block.breadcrumbsByAddress = map[string]*accountBreadcrumb{
			"sender1": {
				initialNonce:    1,
				lastNonce:       1,
				consumedBalance: big.NewInt(5),
			},
			"payer1": {
				initialNonce:    2,
				lastNonce:       2,
				consumedBalance: big.NewInt(3),
			},
		}
		txs := []*WrappedTransaction{
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("sender1"),
					Nonce:   3,
				},
				TransferredValue: big.NewInt(5),
				FeePayer:         []byte("payer1"),
			},
			{
				Tx: &transaction.Transaction{
					SndAddr: []byte("sender1"),
					Nonce:   4,
				},
				TransferredValue: big.NewInt(5),
				FeePayer:         []byte("payer1"),
				Fee:              big.NewInt(3),
			},
		}

		block.compileBreadcrumbs(txs)
		expectedBreadcrumbs := map[string]*accountBreadcrumb{
			"sender1": {
				initialNonce:    1,
				lastNonce:       4,
				consumedBalance: big.NewInt(15),
			},
			"payer1": {
				initialNonce:    2,
				lastNonce:       4,
				consumedBalance: big.NewInt(6),
			},
		}

		for key := range expectedBreadcrumbs {
			_, ok := block.breadcrumbsByAddress[key]
			require.True(t, ok)
			equalBreadcrumbs(t, expectedBreadcrumbs[key], block.breadcrumbsByAddress[key])
		}
	})
}
