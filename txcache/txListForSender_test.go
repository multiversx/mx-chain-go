package txcache

import (
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
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

	require.Equal(t, 4, len(list.items))

	_ = list.removeTransactionsWithLowerOrEqualNonceReturnHashes(43)
	require.Equal(t, 2, len(list.items))

	_ = list.removeTransactionsWithLowerOrEqualNonceReturnHashes(44)
	require.Equal(t, 1, len(list.items))

	_ = list.removeTransactionsWithLowerOrEqualNonceReturnHashes(99)
	require.Equal(t, 0, len(list.items))
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

func TestBenchmarkListForSender_getTxs(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numTransactions = 5_000 getTxs in order", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		numTransactions := 5_000

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		txs := list.getTxs()
		sw.Stop(t.Name())
		require.Len(t, txs, 5000)
	})

	t.Run("numTransactions = 5_000 getTxs reversed", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		numTransactions := 5_000

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		txs := list.getTxsReversed()
		sw.Stop(t.Name())
		require.Len(t, txs, 5000)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}

func TestBenchmarkListForSender_removeTransactionsWithLowerOrEqualNonceReturnHashes(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numTransactions = 1_000", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		numTransactions := 1_000

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		list.removeTransactionsWithLowerOrEqualNonceReturnHashes(1_000)
		sw.Stop(t.Name())
	})

	t.Run("numTransactions = 5_000", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		numTransactions := 5_000

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		list.removeTransactionsWithLowerOrEqualNonceReturnHashes(5000)
		sw.Stop(t.Name())
	})

	t.Run("numTransactions = 10_000", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		numTransactions := 10_000

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		list.removeTransactionsWithLowerOrEqualNonceReturnHashes(10_000)
		sw.Stop(t.Name())
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}

func TestBenchmarkListForSender_removeTransactionsWithHigherOrEqualNonce(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numTransactions = 1_000", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		numTransactions := 1_000

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		list.removeTransactionsWithHigherOrEqualNonce(0)
		sw.Stop(t.Name())
	})

	t.Run("numTransactions = 5_000", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		numTransactions := 5_000

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		list.removeTransactionsWithHigherOrEqualNonce(0)
		sw.Stop(t.Name())
	})

	t.Run("numTransactions = 10_000", func(t *testing.T) {
		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newUnconstrainedListToTest()

		numTransactions := 10_000

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		list.removeTransactionsWithHigherOrEqualNonce(0)
		sw.Stop(t.Name())
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}

func TestBenchmarkListForSender_applySizeConstraints(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numTransactions = 1_000", func(t *testing.T) {
		numTransactions := 1_000

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newListToTest(maxNumBytesPerSenderUpperBoundTest, uint32(numTransactions))

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		list.constraints.maxNumTxs = uint32(numTransactions / 2)
		list.applySizeConstraints(txCache.tracker)
		sw.Stop(t.Name())
	})

	t.Run("numTransactions = 5_000", func(t *testing.T) {
		numTransactions := 5_000

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newListToTest(maxNumBytesPerSenderUpperBoundTest, uint32(numTransactions))

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		list.constraints.maxNumTxs = uint32(numTransactions / 2)
		list.applySizeConstraints(txCache.tracker)
		sw.Stop(t.Name())
	})

	t.Run("numTransactions = 10_000", func(t *testing.T) {
		numTransactions := 10_000

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, math.MaxUint32)
		list := newListToTest(maxNumBytesPerSenderUpperBoundTest, uint32(numTransactions))

		for i := numTransactions - 1; i >= 0; i-- {
			list.AddTx(createTx([]byte(fmt.Sprintf("txHash%d", i)), "alice", uint64(i)), txCache.tracker)
		}

		sw.Start(t.Name())
		list.constraints.maxNumTxs = uint32(numTransactions / 2)
		list.applySizeConstraints(txCache.tracker)
		sw.Stop(t.Name())
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
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

func benchmarkTxList(b *testing.B, opFunc testTxListFunc) {
	numNoncesPerAccount := []int{10, 100, 1_000}
	numAccounts := 1_000

	for _, numNonces := range numNoncesPerAccount {
		b.Run(fmt.Sprintf("noncesPerAccount=%d", numNonces), func(b *testing.B) {
			txCache := newCacheToTest(
				maxNumBytesPerSenderUpperBoundTest,
				math.MaxUint32,
			)

			txs := make([]*WrappedTransaction, 0, numNonces*numAccounts)
			for accIdx := 0; accIdx < numAccounts; accIdx++ {
				txs = append(txs, createTxs(numNonces, accIdx)...)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				list := newUnconstrainedListToTest()

				for _, tx := range txs {
					list.AddTx(tx, txCache.tracker)
				}

				opFunc(txCache, list, uint64(numNonces))
			}
		})
	}
}

func createTxs(size int, senderIdx int) []*WrappedTransaction {
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

type testTxListFunc func(txCache *TxCache, list *txListForSender, numNoncesPerAccount uint64)

func funcRemoveTransactionsWithHigherOrEqualNonce(_ *TxCache, list *txListForSender, _ uint64) {
	list.removeTransactionsWithHigherOrEqualNonce(0)
}

func funcRemoveTransactionsWithLowerOrEqualNonceReturnHashes(_ *TxCache, list *txListForSender, numNoncesPerAccount uint64) {
	list.removeTransactionsWithLowerOrEqualNonceReturnHashes(numNoncesPerAccount)
}

func funcApplySizeConstraints(txCache *TxCache, list *txListForSender, numNoncesPerAccount uint64) {
	list.constraints.maxNumTxs = uint32(numNoncesPerAccount / 2)
	list.applySizeConstraints(txCache.tracker)
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
