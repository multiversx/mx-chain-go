package txcache

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListForSender_AddTx_Sorts(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("a"), ".", 1), txCache.tracker)
	list.AddTx(createTx([]byte("c"), ".", 3), txCache.tracker)
	list.AddTx(createTx([]byte("d"), ".", 4), txCache.tracker)
	list.AddTx(createTx([]byte("b"), ".", 2), txCache.tracker)

	require.Equal(t, []string{"a", "b", "c", "d"}, list.getTxHashesAsStrings())
}

func TestListForSender_AddTx_GivesPriorityToHigherGas(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("a"), ".", 1), txCache.tracker)
	list.AddTx(createTx([]byte("b"), ".", 3).withGasPrice(1.2*oneBillion), txCache.tracker)
	list.AddTx(createTx([]byte("c"), ".", 3).withGasPrice(1.1*oneBillion), txCache.tracker)
	list.AddTx(createTx([]byte("d"), ".", 2), txCache.tracker)
	list.AddTx(createTx([]byte("e"), ".", 3).withGasPrice(1.3*oneBillion), txCache.tracker)

	require.Equal(t, []string{"a", "d", "e", "b", "c"}, list.getTxHashesAsStrings())
}

func TestListForSender_AddTx_SortsCorrectlyWhenSameNonceSamePrice(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("a"), ".", 1).withGasPrice(oneBillion), txCache.tracker)
	list.AddTx(createTx([]byte("b"), ".", 3).withGasPrice(3*oneBillion), txCache.tracker)
	list.AddTx(createTx([]byte("c"), ".", 3).withGasPrice(3*oneBillion), txCache.tracker)
	list.AddTx(createTx([]byte("d"), ".", 3).withGasPrice(2*oneBillion), txCache.tracker)
	list.AddTx(createTx([]byte("e"), ".", 3).withGasPrice(3.5*oneBillion), txCache.tracker)
	list.AddTx(createTx([]byte("f"), ".", 2).withGasPrice(oneBillion), txCache.tracker)
	list.AddTx(createTx([]byte("g"), ".", 3).withGasPrice(2.5*oneBillion), txCache.tracker)

	// In case of same-nonce, same-price transactions, the newer one has priority
	require.Equal(t, []string{"a", "f", "e", "b", "c", "g", "d"}, list.getTxHashesAsStrings())
}

func TestListForSender_AddTx_IgnoresDuplicates(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	added, _ := list.AddTx(createTx([]byte("tx1"), ".", 1), txCache.tracker)
	require.True(t, added)
	added, _ = list.AddTx(createTx([]byte("tx2"), ".", 2), txCache.tracker)
	require.True(t, added)
	added, _ = list.AddTx(createTx([]byte("tx3"), ".", 3), txCache.tracker)
	require.True(t, added)
	added, _ = list.AddTx(createTx([]byte("tx2"), ".", 2), txCache.tracker)
	require.False(t, added)
}

func TestListForSender_AddTx_AppliesSizeConstraintsForNumTransactions(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newListToTest(math.MaxUint32, 3)

	list.AddTx(createTx([]byte("tx1"), ".", 1), txCache.tracker)
	list.AddTx(createTx([]byte("tx5"), ".", 5), txCache.tracker)
	list.AddTx(createTx([]byte("tx4"), ".", 4), txCache.tracker)
	list.AddTx(createTx([]byte("tx2"), ".", 2), txCache.tracker)
	require.Equal(t, []string{"tx1", "tx2", "tx4"}, list.getTxHashesAsStrings())

	_, evicted := list.AddTx(createTx([]byte("tx3"), ".", 3), txCache.tracker)
	require.Equal(t, []string{"tx1", "tx2", "tx3"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx4"}, hashesAsStrings(evicted))

	// Gives priority to higher gas - though undesirable to some extent, "tx3" is evicted
	_, evicted = list.AddTx(createTx([]byte("tx2++"), ".", 2).withGasPrice(1.5*oneBillion), txCache.tracker)
	require.Equal(t, []string{"tx1", "tx2++", "tx2"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx3"}, hashesAsStrings(evicted))

	// Though undesirable to some extent, "tx3++"" is added, then evicted
	_, evicted = list.AddTx(createTx([]byte("tx3++"), ".", 3).withGasPrice(1.5*oneBillion), txCache.tracker)
	require.Equal(t, []string{"tx1", "tx2++", "tx2"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx3++"}, hashesAsStrings(evicted))
}

func TestListForSender_AddTx_AppliesSizeConstraintsForNumBytes(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newListToTest(1024, math.MaxUint32)

	list.AddTx(createTx([]byte("tx1"), ".", 1).withSize(128).withGasLimit(50000), txCache.tracker)
	list.AddTx(createTx([]byte("tx2"), ".", 2).withSize(512).withGasLimit(1500000), txCache.tracker)
	list.AddTx(createTx([]byte("tx3"), ".", 3).withSize(256).withGasLimit(1500000), txCache.tracker)
	_, evicted := list.AddTx(createTx([]byte("tx5"), ".", 4).withSize(256).withGasLimit(1500000), txCache.tracker)
	require.Equal(t, []string{"tx1", "tx2", "tx3"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx5"}, hashesAsStrings(evicted))

	_, evicted = list.AddTx(createTx([]byte("tx5--"), ".", 4).withSize(128).withGasLimit(50000), txCache.tracker)
	require.Equal(t, []string{"tx1", "tx2", "tx3", "tx5--"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{}, hashesAsStrings(evicted))

	_, evicted = list.AddTx(createTx([]byte("tx4"), ".", 4).withSize(128).withGasLimit(50000), txCache.tracker)
	require.Equal(t, []string{"tx1", "tx2", "tx3", "tx4"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx5--"}, hashesAsStrings(evicted))

	// Gives priority to higher gas - though undesirably to some extent, "tx4" is evicted
	_, evicted = list.AddTx(createTx([]byte("tx3++"), ".", 3).withSize(256).withGasLimit(1500000).withGasPrice(1.5*oneBillion), txCache.tracker)
	require.Equal(t, []string{"tx1", "tx2", "tx3++"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx4", "tx3"}, hashesAsStrings(evicted))
}

func TestListForSender_removeTransactionsWithLowerOrEqualNonceReturnHashes(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("tx-42"), ".", 42), txCache.tracker)
	list.AddTx(createTx([]byte("tx-43"), ".", 43), txCache.tracker)
	list.AddTx(createTx([]byte("tx-44"), ".", 44), txCache.tracker)
	list.AddTx(createTx([]byte("tx-45"), ".", 45), txCache.tracker)

	require.Equal(t, 4, list.list.len())

	_ = list.removeTransactionsWithLowerOrEqualNonceReturnHashes(43)
	require.Equal(t, 2, list.list.len())

	_ = list.removeTransactionsWithLowerOrEqualNonceReturnHashes(44)
	require.Equal(t, 1, list.list.len())

	_ = list.removeTransactionsWithLowerOrEqualNonceReturnHashes(99)
	require.Equal(t, 0, list.list.len())
}

func TestListForSender_getTxs(t *testing.T) {
	t.Run("without transactions", func(t *testing.T) {
		list := newUnconstrainedListToTest()

		require.Len(t, list.getTxs(), 0)
		require.Len(t, list.getTxsReversed(), 0)
	})

	t.Run("with transactions", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		list.AddTx(createTx([]byte("tx-42"), ".", 42), txCache.tracker)
		require.Len(t, list.getTxs(), 1)
		require.Len(t, list.getTxsReversed(), 1)

		list.AddTx(createTx([]byte("tx-44"), ".", 44), txCache.tracker)
		require.Len(t, list.getTxs(), 2)
		require.Len(t, list.getTxsReversed(), 2)

		list.AddTx(createTx([]byte("tx-43"), ".", 43), txCache.tracker)
		require.Len(t, list.getTxs(), 3)
		require.Len(t, list.getTxsReversed(), 3)

		require.Equal(t, []byte("tx-42"), list.getTxs()[0].TxHash)
		require.Equal(t, []byte("tx-43"), list.getTxs()[1].TxHash)
		require.Equal(t, []byte("tx-44"), list.getTxs()[2].TxHash)
		require.Equal(t, []byte("tx-44"), list.getTxsReversed()[0].TxHash)
		require.Equal(t, []byte("tx-43"), list.getTxsReversed()[1].TxHash)
		require.Equal(t, []byte("tx-42"), list.getTxsReversed()[2].TxHash)
	})
}

func TestListForSender_DetectRaceConditions(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	wg := sync.WaitGroup{}

	doOperations := func() {
		// These might be called concurrently:
		_ = list.IsEmpty()
		_ = list.getTxs()
		_ = list.getTxsReversed()
		_ = list.countTxWithLock()
		_, _ = list.AddTx(createTx([]byte("test"), ".", 42), txCache.tracker)

		wg.Done()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go doOperations()
	}

	wg.Wait()
}

const numAccountsForBenchmark = 1_000

type testTxListFunc func(txCache *TxCache, list *txListForSender, numNoncesPerAccount uint64)
type createTxsFunc func(size int, senderIdx int) []*WrappedTransaction

func funcRemoveTransactionsWithHigherOrEqualNonce(_ *TxCache, list *txListForSender, _ uint64) {
	list.removeTransactionsWithHigherOrEqualNonce(0)
}

func funcRemoveTransactionsWithLowerOrEqualNonceReturnHashes(_ *TxCache, list *txListForSender, numNoncesPerAccount uint64) {
	list.removeTransactionsWithLowerOrEqualNonceReturnHashes(numNoncesPerAccount)
}

func funcApplySizeConstraints(txCache *TxCache, list *txListForSender, numNoncesPerAccount uint64) {
	list.constraints.maxNumTxs = uint32(numNoncesPerAccount * numAccountsForBenchmark / 2)
	list.applySizeConstraints(txCache.tracker)
}

func createTxsOrdered(size int, senderIdx int) []*WrappedTransaction {
	txs := make([]*WrappedTransaction, 0, size)

	for i := 0; i < size; i++ {
		txs = append(txs,
			createTx(
				[]byte(fmt.Sprintf("txHash%d", i)),
				fmt.Sprintf("sender-%d", senderIdx),
				uint64(i),
			),
		)
	}

	return txs
}

func createTxsReversed(size int, senderIdx int) []*WrappedTransaction {
	txs := make([]*WrappedTransaction, 0, size)

	for i := size - 1; i >= 0; i-- {
		txs = append(txs,
			createTx(
				[]byte(fmt.Sprintf("txHash%d", i)),
				fmt.Sprintf("sender-%d", senderIdx),
				uint64(i),
			),
		)
	}

	return txs
}

func createTxsRandom(size int, senderIdx int) []*WrappedTransaction {
	txs := createTxsOrdered(size, senderIdx)
	rand.Shuffle(size, func(i, j int) { txs[i], txs[j] = txs[j], txs[i] })
	return txs
}

func BenchmarkTxList_removeTransactionsWithHigherOrEqualNonce(b *testing.B) {
	benchmarkTxList(b, funcRemoveTransactionsWithHigherOrEqualNonce)
}

func BenchmarkTxList_removeTransactionsWithLowerOrEqualNonceReturnHashes(b *testing.B) {
	benchmarkTxList(b, funcRemoveTransactionsWithLowerOrEqualNonceReturnHashes)
}

func BenchmarkTxList_applySizeConstraints(b *testing.B) {
	benchmarkTxList(b, funcApplySizeConstraints)
}

func benchmarkTxList(b *testing.B, testFunc testTxListFunc) {
	b.Run("ordered", func(b *testing.B) {
		benchmarkTxListForTxOrder(b, testFunc, createTxsOrdered)
	})
	b.Run("reversed", func(b *testing.B) {
		benchmarkTxListForTxOrder(b, testFunc, createTxsReversed)
	})
	b.Run("random", func(b *testing.B) {
		benchmarkTxListForTxOrder(b, testFunc, createTxsRandom)
	})
}

func benchmarkTxListForTxOrder(b *testing.B, testFunc testTxListFunc, createTxFunc createTxsFunc) {
	numNoncesPerAccount := []int{10, 100, 1_000}

	for _, numNonces := range numNoncesPerAccount {
		b.Run(fmt.Sprintf("noncesPerAccount=%d", numNonces), func(b *testing.B) {
			txCache := newCacheToTest(
				maxNumBytesPerSenderUpperBoundTest,
				math.MaxUint32,
			)

			txs := make([]*WrappedTransaction, 0, numNonces*numAccountsForBenchmark)
			for accIdx := 0; accIdx < numAccountsForBenchmark; accIdx++ {
				txs = append(txs, createTxFunc(numNonces, accIdx)...)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				list := newUnconstrainedListToTest()

				for _, tx := range txs {
					list.AddTx(tx, txCache.tracker)
				}

				testFunc(txCache, list, uint64(numNonces))
			}
		})
	}
}

func TestListForSender_getTxsForSelection(t *testing.T) {
	t.Run("without transactions", func(t *testing.T) {
		list := newUnconstrainedListToTest()
		require.Len(t, list.getTxsForSelection(), 0)
	})

	t.Run("with zero offset", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		list.AddTx(createTx([]byte("tx-42"), ".", 42), txCache.tracker)
		list.AddTx(createTx([]byte("tx-43"), ".", 43), txCache.tracker)
		list.AddTx(createTx([]byte("tx-44"), ".", 44), txCache.tracker)

		txs := list.getTxsForSelection()
		require.Len(t, txs, 3)
		require.Equal(t, []byte("tx-42"), txs[0].TxHash)
		require.Equal(t, []byte("tx-43"), txs[1].TxHash)
		require.Equal(t, []byte("tx-44"), txs[2].TxHash)
	})

	t.Run("with offset = 1", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		list.AddTx(createTx([]byte("tx-42"), ".", 42), txCache.tracker)
		list.AddTx(createTx([]byte("tx-43"), ".", 43), txCache.tracker)
		list.AddTx(createTx([]byte("tx-44"), ".", 44), txCache.tracker)

		list.mutex.Lock()
		list.incrementSelectionOffset(1)
		list.mutex.Unlock()

		txs := list.getTxsForSelection()
		require.Len(t, txs, 2)
		require.Equal(t, []byte("tx-43"), txs[0].TxHash)
		require.Equal(t, []byte("tx-44"), txs[1].TxHash)
	})

	t.Run("with offset = all transactions", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		list.AddTx(createTx([]byte("tx-42"), ".", 42), txCache.tracker)
		list.AddTx(createTx([]byte("tx-43"), ".", 43), txCache.tracker)

		list.mutex.Lock()
		list.incrementSelectionOffset(2)
		list.mutex.Unlock()

		txs := list.getTxsForSelection()
		require.Len(t, txs, 0)
	})

	t.Run("with offset exceeds list length", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		list.AddTx(createTx([]byte("tx-42"), ".", 42), txCache.tracker)

		list.mutex.Lock()
		list.incrementSelectionOffset(10)
		list.mutex.Unlock()

		// Offset should be clamped to list length
		require.Equal(t, 1, list.getSelectionOffset())
		txs := list.getTxsForSelection()
		require.Len(t, txs, 0)
	})
}

func TestListForSender_incrementSelectionOffset(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("tx-1"), ".", 1), txCache.tracker)
	list.AddTx(createTx([]byte("tx-2"), ".", 2), txCache.tracker)
	list.AddTx(createTx([]byte("tx-3"), ".", 3), txCache.tracker)

	require.Equal(t, 0, list.getSelectionOffset())

	list.mutex.Lock()
	list.incrementSelectionOffset(1)
	list.mutex.Unlock()
	require.Equal(t, 1, list.getSelectionOffset())

	list.mutex.Lock()
	list.incrementSelectionOffset(2)
	list.mutex.Unlock()
	require.Equal(t, 3, list.getSelectionOffset())

	// Should be clamped to list length
	list.mutex.Lock()
	list.incrementSelectionOffset(5)
	list.mutex.Unlock()
	require.Equal(t, 3, list.getSelectionOffset())
}

func TestListForSender_decrementSelectionOffset(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("tx-1"), ".", 1), txCache.tracker)
	list.AddTx(createTx([]byte("tx-2"), ".", 2), txCache.tracker)
	list.AddTx(createTx([]byte("tx-3"), ".", 3), txCache.tracker)

	list.mutex.Lock()
	list.incrementSelectionOffset(3)
	list.mutex.Unlock()
	require.Equal(t, 3, list.getSelectionOffset())

	list.mutex.Lock()
	list.decrementSelectionOffset(1)
	list.mutex.Unlock()
	require.Equal(t, 2, list.getSelectionOffset())

	list.mutex.Lock()
	list.decrementSelectionOffset(2)
	list.mutex.Unlock()
	require.Equal(t, 0, list.getSelectionOffset())

	// Should be clamped to 0
	list.mutex.Lock()
	list.decrementSelectionOffset(5)
	list.mutex.Unlock()
	require.Equal(t, 0, list.getSelectionOffset())
}

func TestListForSender_resetSelectionOffsetByNonce(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("tx-10"), ".", 10), txCache.tracker)
	list.AddTx(createTx([]byte("tx-20"), ".", 20), txCache.tracker)
	list.AddTx(createTx([]byte("tx-30"), ".", 30), txCache.tracker)
	list.AddTx(createTx([]byte("tx-40"), ".", 40), txCache.tracker)

	// Reset to first transaction with nonce >= 20
	list.resetSelectionOffsetByNonce(20)
	require.Equal(t, 1, list.getSelectionOffset())

	// Reset to first transaction with nonce >= 35 (should be tx-40 at index 3)
	list.resetSelectionOffsetByNonce(35)
	require.Equal(t, 3, list.getSelectionOffset())

	// Reset to first transaction with nonce >= 10 (should be tx-10 at index 0)
	list.resetSelectionOffsetByNonce(10)
	require.Equal(t, 0, list.getSelectionOffset())

	// Reset with nonce beyond all transactions
	list.resetSelectionOffsetByNonce(100)
	require.Equal(t, 4, list.getSelectionOffset())
}

func TestListForSender_OffsetAdjustsOnAddTx(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("tx-10"), ".", 10), txCache.tracker)
	list.AddTx(createTx([]byte("tx-20"), ".", 20), txCache.tracker)
	list.AddTx(createTx([]byte("tx-30"), ".", 30), txCache.tracker)

	// Set offset to 2 (pointing to tx-30)
	list.mutex.Lock()
	list.incrementSelectionOffset(2)
	list.mutex.Unlock()
	require.Equal(t, 2, list.getSelectionOffset())

	// Add transaction with nonce 15 (inserts before offset)
	list.AddTx(createTx([]byte("tx-15"), ".", 15), txCache.tracker)
	// Offset should be incremented to 3
	require.Equal(t, 3, list.getSelectionOffset())

	// Add transaction with nonce 35 (inserts after offset)
	list.AddTx(createTx([]byte("tx-35"), ".", 35), txCache.tracker)
	// Offset should remain 3
	require.Equal(t, 3, list.getSelectionOffset())
}

func TestListForSender_OffsetAdjustsOnRemove(t *testing.T) {
	txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("tx-10"), ".", 10), txCache.tracker)
	list.AddTx(createTx([]byte("tx-20"), ".", 20), txCache.tracker)
	list.AddTx(createTx([]byte("tx-30"), ".", 30), txCache.tracker)
	list.AddTx(createTx([]byte("tx-40"), ".", 40), txCache.tracker)

	// Set offset to 2
	list.mutex.Lock()
	list.incrementSelectionOffset(2)
	list.mutex.Unlock()
	require.Equal(t, 2, list.getSelectionOffset())

	// Remove transactions with nonce <= 10 (removes 1 tx before offset)
	list.removeTransactionsWithLowerOrEqualNonceReturnHashes(10)
	// Offset should be decremented to 1
	require.Equal(t, 1, list.getSelectionOffset())

	// Remove transactions with nonce <= 20 (removes 1 tx before offset)
	list.removeTransactionsWithLowerOrEqualNonceReturnHashes(20)
	// Offset should be decremented to 0
	require.Equal(t, 0, list.getSelectionOffset())
}

func newUnconstrainedListToTest() *txListForSender {
	return newListToTest(math.MaxUint32, math.MaxUint32)
}

func newListToTest(maxNumBytes uint32, maxNumTxs uint32) *txListForSender {
	senderConstraints := &senderConstraints{
		maxNumBytes: maxNumBytes,
		maxNumTxs:   maxNumTxs,
	}

	return newTxListForSender(".", senderConstraints)
}
