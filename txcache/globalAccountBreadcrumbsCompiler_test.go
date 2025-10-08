package txcache

import (
	"fmt"
	"math"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func requireEqualGlobalAccountsBreadcrumbs(
	t *testing.T,
	expected map[string]*globalAccountBreadcrumb,
	actual map[string]*globalAccountBreadcrumb,
) {
	require.Equal(t, len(expected), len(actual))
	for account, globalBreadcrumb := range expected {
		_, ok := actual[account]
		require.True(t, ok)
		require.Equal(t, globalBreadcrumb.lastNonce, actual[account].lastNonce)
		require.Equal(t, globalBreadcrumb.firstNonce, actual[account].firstNonce)
		require.Equal(t, 0, globalBreadcrumb.consumedBalance.Cmp(actual[account].consumedBalance))
	}
}

func Test_newGlobalAccountBreadcrumbsCompiler(t *testing.T) {
	t.Parallel()

	gabc := newGlobalAccountBreadcrumbsCompiler()
	require.NotNil(t, gabc)
}

func Test_shouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should not panic")
		}
	}()

	gabc := newGlobalAccountBreadcrumbsCompiler()
	numOfBlocks := 10

	wg := sync.WaitGroup{}
	wg.Add(2 * numOfBlocks)

	for i := 1; i <= numOfBlocks; i++ {
		tb := newTrackedBlock(uint64(i), []byte(fmt.Sprintf("hash%d", i-1)), []byte("rootHash0"), []byte(fmt.Sprintf("prevHash%d", i-1)))

		go func(tb *trackedBlock) {
			defer wg.Done()

			gabc.updateOnAddedBlock(tb)
		}(tb)

		go func(tb *trackedBlock) {
			defer wg.Done()

			err := gabc.updateAfterRemovedBlockWithSameNonceOrBelow(tb)
			require.NoError(t, err)
		}(tb)
	}

	wg.Wait()
}

func Test_shouldWorkOnDifferentScenarios(t *testing.T) {
	t.Parallel()
	t.Run("should work on most common scenario", func(t *testing.T) {
		t.Parallel()

		gabc := newGlobalAccountBreadcrumbsCompiler()

		// create a first proposed block
		tb1 := newTrackedBlock(0, []byte("hash0"), []byte("rootHash"), []byte("prevHash"))

		tx1 := createTx([]byte("hash1"), "alice", 0).withRelayer([]byte("bob")).withTransferredValue(big.NewInt(1)).withFee(big.NewInt(1))
		tx2 := createTx([]byte("hash2"), "carol", 0).withTransferredValue(big.NewInt(2)).withFee(big.NewInt(1))

		txs := []*WrappedTransaction{
			tx1, tx2,
		}

		// compile its breadcrumbs
		err := tb1.compileBreadcrumbs(txs)
		require.NoError(t, err)

		// update the global state
		gabc.updateOnAddedBlock(tb1)

		expectedGlobalBreadcrumbs := map[string]*globalAccountBreadcrumb{
			"alice": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    0,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    0,
				},
				consumedBalance: big.NewInt(1),
			},
			"carol": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    0,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    0,
				},
				consumedBalance: big.NewInt(2),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					HasValue: false,
					Value:    math.MaxUint64,
				},
				lastNonce: core.OptionalUint64{
					HasValue: false,
					Value:    0,
				},
				consumedBalance: big.NewInt(1),
			},
		}

		requireEqualGlobalAccountsBreadcrumbs(t, expectedGlobalBreadcrumbs, gabc.globalAccountBreadcrumbs)

		// create the second proposed block
		tb2 := newTrackedBlock(0, []byte("hash0"), []byte("rootHash"), []byte("prevHash"))

		tx3 := createTx([]byte("hash3"), "alice", 1).withTransferredValue(big.NewInt(1)).withFee(big.NewInt(1))
		tx4 := createTx([]byte("hash4"), "carol", 1).withTransferredValue(big.NewInt(2)).withFee(big.NewInt(1))

		txs = []*WrappedTransaction{
			tx3, tx4,
		}

		// compile its breadcrumbs
		err = tb2.compileBreadcrumbs(txs)
		require.NoError(t, err)

		// update the global state
		gabc.updateOnAddedBlock(tb2)

		expectedGlobalBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"alice": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    0,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    1,
				},
				consumedBalance: big.NewInt(2),
			},
			"carol": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    0,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    1,
				},
				consumedBalance: big.NewInt(4),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					HasValue: false,
					Value:    math.MaxUint64,
				},
				lastNonce: core.OptionalUint64{
					HasValue: false,
					Value:    0,
				},
				consumedBalance: big.NewInt(1),
			},
		}

		requireEqualGlobalAccountsBreadcrumbs(t, expectedGlobalBreadcrumbs, gabc.globalAccountBreadcrumbs)

		// remove the first proposed block and update the global state
		err = gabc.updateAfterRemovedBlockWithSameNonceOrBelow(tb1)
		require.NoError(t, err)

		expectedGlobalBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"alice": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    1,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    1,
				},
				consumedBalance: big.NewInt(1),
			},
			"carol": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    1,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    1,
				},
				consumedBalance: big.NewInt(2),
			},
		}

		requireEqualGlobalAccountsBreadcrumbs(t, expectedGlobalBreadcrumbs, gabc.globalAccountBreadcrumbs)

		// remove the second proposed block and update the global state
		err = gabc.updateAfterRemovedBlockWithSameNonceOrBelow(tb2)
		require.NoError(t, err)

		expectedGlobalBreadcrumbs = map[string]*globalAccountBreadcrumb{}
		requireEqualGlobalAccountsBreadcrumbs(t, expectedGlobalBreadcrumbs, gabc.globalAccountBreadcrumbs)

	})

	t.Run("should work for forks scenarios", func(t *testing.T) {
		t.Parallel()

		gabc := newGlobalAccountBreadcrumbsCompiler()
		// propose another block
		tb3 := newTrackedBlock(0, []byte("hash0"), []byte("rootHash"), []byte("prevHash"))

		tx5 := createTx([]byte("hash5"), "eve", 5).withRelayer([]byte("bob")).withFee(big.NewInt(1))
		tx6 := createTx([]byte("hash6"), "dave", 10).withRelayer([]byte("bob")).withFee(big.NewInt(1))
		tx7 := createTx([]byte("hash7"), "eve", 6).withRelayer([]byte("bob")).withFee(big.NewInt(1))
		tx8 := createTx([]byte("hash8"), "dave", 11).withRelayer([]byte("bob")).withFee(big.NewInt(1))
		tx9 := createTx([]byte("hash9"), "eve", 7).withRelayer([]byte("bob")).withFee(big.NewInt(1))

		txs := []*WrappedTransaction{
			tx5, tx6, tx7, tx8, tx9,
		}

		// compile its breadcrumbs
		err := tb3.compileBreadcrumbs(txs)
		require.NoError(t, err)

		// update the global state
		gabc.updateOnAddedBlock(tb3)

		expectedGlobalBreadcrumbs := map[string]*globalAccountBreadcrumb{
			"eve": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    5,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    7,
				},
				consumedBalance: big.NewInt(0),
			},
			"dave": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    10,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    11,
				},
				consumedBalance: big.NewInt(0),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					HasValue: false,
					Value:    math.MaxUint64,
				},
				lastNonce: core.OptionalUint64{
					HasValue: false,
					Value:    0,
				},
				consumedBalance: big.NewInt(5),
			},
		}

		requireEqualGlobalAccountsBreadcrumbs(t, expectedGlobalBreadcrumbs, gabc.globalAccountBreadcrumbs)

		// start a non-canonical chain
		// propose another block
		tb4 := newTrackedBlock(0, []byte("hash0"), []byte("rootHash"), []byte("prevHash"))

		tx10 := createTx([]byte("hash10"), "dave", 12).withRelayer([]byte("eve")).withFee(big.NewInt(1))
		tx11 := createTx([]byte("hash11"), "dave", 13).withRelayer([]byte("eve")).withFee(big.NewInt(1))

		txs = []*WrappedTransaction{
			tx10, tx11,
		}

		// compile its breadcrumbs
		err = tb4.compileBreadcrumbs(txs)
		require.NoError(t, err)

		// update the global state
		gabc.updateOnAddedBlock(tb4)

		expectedGlobalBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"eve": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    5,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    7,
				},
				consumedBalance: big.NewInt(2),
			},
			"dave": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    10,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    13,
				},
				consumedBalance: big.NewInt(0),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					HasValue: false,
					Value:    math.MaxUint64,
				},
				lastNonce: core.OptionalUint64{
					HasValue: false,
					Value:    0,
				},
				consumedBalance: big.NewInt(5),
			},
		}

		requireEqualGlobalAccountsBreadcrumbs(t, expectedGlobalBreadcrumbs, gabc.globalAccountBreadcrumbs)

		// propose another block to the non-canonical chain
		tb5 := newTrackedBlock(0, []byte("hash0"), []byte("rootHash"), []byte("prevHash"))

		tx12 := createTx([]byte("hash12"), "bob", 5).withTransferredValue(big.NewInt(1)).withFee(big.NewInt(1))
		tx13 := createTx([]byte("hash13"), "alice", 20).withTransferredValue(big.NewInt(3)).withFee(big.NewInt(1))

		txs = []*WrappedTransaction{
			tx12, tx13,
		}

		// compile its breadcrumbs
		err = tb5.compileBreadcrumbs(txs)
		require.NoError(t, err)

		// propose
		gabc.updateOnAddedBlock(tb5)

		expectedGlobalBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"eve": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    5,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    7,
				},
				consumedBalance: big.NewInt(2),
			},
			"dave": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    10,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    13,
				},
				consumedBalance: big.NewInt(0),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    5,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    5,
				},
				consumedBalance: big.NewInt(6),
			},
			"alice": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    20,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    20,
				},
				consumedBalance: big.NewInt(3),
			},
		}

		requireEqualGlobalAccountsBreadcrumbs(t, expectedGlobalBreadcrumbs, gabc.globalAccountBreadcrumbs)

		// now, replace the first block in the non-canonical chain
		// first, remove all the once greater or equal to its nonce

		err = gabc.updateAfterRemovedBlockWithSameNonceOrAbove(tb4)
		require.NoError(t, err)

		expectedGlobalBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"eve": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    5,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    7,
				},
				consumedBalance: big.NewInt(0),
			},
			"dave": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    10,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    11,
				},
				consumedBalance: big.NewInt(0),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    5,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    5,
				},
				consumedBalance: big.NewInt(6),
			},
			"alice": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    20,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    20,
				},
				consumedBalance: big.NewInt(3),
			},
		}

		requireEqualGlobalAccountsBreadcrumbs(t, expectedGlobalBreadcrumbs, gabc.globalAccountBreadcrumbs)

		err = gabc.updateAfterRemovedBlockWithSameNonceOrAbove(tb5)
		require.NoError(t, err)

		expectedGlobalBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"eve": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    5,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    7,
				},
				consumedBalance: big.NewInt(0),
			},
			"dave": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    10,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    11,
				},
				consumedBalance: big.NewInt(0),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					HasValue: false,
					Value:    math.MaxUint64,
				},
				lastNonce: core.OptionalUint64{
					HasValue: false,
					Value:    0,
				},
				consumedBalance: big.NewInt(5),
			},
		}

		requireEqualGlobalAccountsBreadcrumbs(t, expectedGlobalBreadcrumbs, gabc.globalAccountBreadcrumbs)

		// now, add the new block, the one that replaces the first of the non-canonical chain

		tb6 := newTrackedBlock(0, []byte("hash0"), []byte("rootHash"), []byte("prevHash"))

		tx14 := createTx([]byte("hash14"), "frank", 0).withRelayer([]byte("eve")).withFee(big.NewInt(1))
		tx15 := createTx([]byte("hash15"), "frank", 1).withRelayer([]byte("eve")).withFee(big.NewInt(1))

		txs = []*WrappedTransaction{
			tx14, tx15,
		}

		// compile its breadcrumbs
		err = tb6.compileBreadcrumbs(txs)
		require.NoError(t, err)

		// update the global state
		gabc.updateOnAddedBlock(tb6)

		expectedGlobalBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"eve": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    5,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    7,
				},
				consumedBalance: big.NewInt(2),
			},
			"dave": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    10,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    11,
				},
				consumedBalance: big.NewInt(0),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					HasValue: false,
					Value:    math.MaxUint64,
				},
				lastNonce: core.OptionalUint64{
					HasValue: false,
					Value:    0,
				},
				consumedBalance: big.NewInt(5),
			},
			"frank": {
				firstNonce: core.OptionalUint64{
					HasValue: true,
					Value:    0,
				},
				lastNonce: core.OptionalUint64{
					HasValue: true,
					Value:    1,
				},
				consumedBalance: big.NewInt(0),
			},
		}

		requireEqualGlobalAccountsBreadcrumbs(t, expectedGlobalBreadcrumbs, gabc.globalAccountBreadcrumbs)
	})
}

func Test_updateGlobalBreadcrumbsOnRemovedBlockOnProposed(t *testing.T) {
	t.Parallel()

	t.Run("should return errGlobalBreadcrumbDoesNotExist error", func(t *testing.T) {
		t.Parallel()

		gabc := newGlobalAccountBreadcrumbsCompiler()

		tb1 := newTrackedBlock(0, []byte("hash0"), []byte("rootHash"), []byte("prevHash"))
		tx1 := createTx([]byte("hash14"), "frank", 0).withRelayer([]byte("eve")).withFee(big.NewInt(1))
		tx2 := createTx([]byte("hash15"), "frank", 1).withRelayer([]byte("eve")).withFee(big.NewInt(1))

		txs := []*WrappedTransaction{
			tx1, tx2,
		}

		err := tb1.compileBreadcrumbs(txs)
		require.NoError(t, err)

		err = gabc.updateAfterRemovedBlockWithSameNonceOrAbove(tb1)
		require.Equal(t, errGlobalBreadcrumbDoesNotExist, err)
	})

	t.Run("should return errNegativeBalanceForBreadcrumb error", func(t *testing.T) {
		t.Parallel()

		gabc := newGlobalAccountBreadcrumbsCompiler()
		gabc.globalAccountBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"frank": {
				firstNonce:      core.OptionalUint64{},
				lastNonce:       core.OptionalUint64{},
				consumedBalance: big.NewInt(0),
			},
			"eve": {
				firstNonce:      core.OptionalUint64{},
				lastNonce:       core.OptionalUint64{},
				consumedBalance: big.NewInt(0),
			},
		}
		tb1 := newTrackedBlock(0, []byte("hash0"), []byte("rootHash"), []byte("prevHash"))
		tx1 := createTx([]byte("hash14"), "frank", 0).withRelayer([]byte("eve")).withFee(big.NewInt(1))
		tx2 := createTx([]byte("hash15"), "frank", 1).withRelayer([]byte("eve")).withFee(big.NewInt(1))

		txs := []*WrappedTransaction{
			tx1, tx2,
		}

		err := tb1.compileBreadcrumbs(txs)
		require.NoError(t, err)

		err = gabc.updateAfterRemovedBlockWithSameNonceOrAbove(tb1)
		require.Equal(t, errNegativeBalanceForBreadcrumb, err)
	})
}

func Test_updateGlobalBreadcrumbsOnRemovedBlockOnExecuted(t *testing.T) {
	t.Parallel()

	t.Run("should return errGlobalBreadcrumbDoesNotExist error", func(t *testing.T) {
		t.Parallel()

		gabc := newGlobalAccountBreadcrumbsCompiler()

		tb1 := newTrackedBlock(0, []byte("hash0"), []byte("rootHash"), []byte("prevHash"))
		tx1 := createTx([]byte("hash14"), "frank", 0).withRelayer([]byte("eve")).withFee(big.NewInt(1))
		tx2 := createTx([]byte("hash15"), "frank", 1).withRelayer([]byte("eve")).withFee(big.NewInt(1))

		txs := []*WrappedTransaction{
			tx1, tx2,
		}

		err := tb1.compileBreadcrumbs(txs)
		require.NoError(t, err)

		err = gabc.updateAfterRemovedBlockWithSameNonceOrBelow(tb1)
		require.Equal(t, errGlobalBreadcrumbDoesNotExist, err)
	})

	t.Run("should return errNegativeBalanceForBreadcrumb error", func(t *testing.T) {
		t.Parallel()

		gabc := newGlobalAccountBreadcrumbsCompiler()
		gabc.globalAccountBreadcrumbs = map[string]*globalAccountBreadcrumb{
			"frank": {
				firstNonce:      core.OptionalUint64{},
				lastNonce:       core.OptionalUint64{},
				consumedBalance: big.NewInt(0),
			},
			"eve": {
				firstNonce:      core.OptionalUint64{},
				lastNonce:       core.OptionalUint64{},
				consumedBalance: big.NewInt(0),
			},
		}
		tb1 := newTrackedBlock(0, []byte("hash0"), []byte("rootHash"), []byte("prevHash"))
		tx1 := createTx([]byte("hash14"), "frank", 0).withRelayer([]byte("eve")).withFee(big.NewInt(1))
		tx2 := createTx([]byte("hash15"), "frank", 1).withRelayer([]byte("eve")).withFee(big.NewInt(1))

		txs := []*WrappedTransaction{
			tx1, tx2,
		}

		err := tb1.compileBreadcrumbs(txs)
		require.NoError(t, err)

		err = gabc.updateAfterRemovedBlockWithSameNonceOrBelow(tb1)
		require.Equal(t, errNegativeBalanceForBreadcrumb, err)
	})
}

func Test_cleanGlobalBreadcrumbs(t *testing.T) {
	t.Parallel()

	gabc := newGlobalAccountBreadcrumbsCompiler()
	gabc.globalAccountBreadcrumbs = map[string]*globalAccountBreadcrumb{
		"alice": {},
		"bob":   {},
	}

	require.Equal(t, 2, len(gabc.globalAccountBreadcrumbs))

	gabc.cleanGlobalBreadcrumbs()
	require.Equal(t, 0, len(gabc.globalAccountBreadcrumbs))
}
