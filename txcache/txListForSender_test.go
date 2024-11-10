package txcache

import (
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListForSender_AddTx_Sorts(t *testing.T) {
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("a"), ".", 1))
	list.AddTx(createTx([]byte("c"), ".", 3))
	list.AddTx(createTx([]byte("d"), ".", 4))
	list.AddTx(createTx([]byte("b"), ".", 2))

	require.Equal(t, []string{"a", "b", "c", "d"}, list.getTxHashesAsStrings())
}

func TestListForSender_AddTx_GivesPriorityToHigherGas(t *testing.T) {
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("a"), ".", 1))
	list.AddTx(createTx([]byte("b"), ".", 3).withGasPrice(1.2 * oneBillion))
	list.AddTx(createTx([]byte("c"), ".", 3).withGasPrice(1.1 * oneBillion))
	list.AddTx(createTx([]byte("d"), ".", 2))
	list.AddTx(createTx([]byte("e"), ".", 3).withGasPrice(1.3 * oneBillion))

	require.Equal(t, []string{"a", "d", "e", "b", "c"}, list.getTxHashesAsStrings())
}

func TestListForSender_AddTx_SortsCorrectlyWhenSameNonceSamePrice(t *testing.T) {
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("a"), ".", 1).withGasPrice(oneBillion))
	list.AddTx(createTx([]byte("b"), ".", 3).withGasPrice(3 * oneBillion))
	list.AddTx(createTx([]byte("c"), ".", 3).withGasPrice(3 * oneBillion))
	list.AddTx(createTx([]byte("d"), ".", 3).withGasPrice(2 * oneBillion))
	list.AddTx(createTx([]byte("e"), ".", 3).withGasPrice(3.5 * oneBillion))
	list.AddTx(createTx([]byte("f"), ".", 2).withGasPrice(oneBillion))
	list.AddTx(createTx([]byte("g"), ".", 3).withGasPrice(2.5 * oneBillion))

	// In case of same-nonce, same-price transactions, the newer one has priority
	require.Equal(t, []string{"a", "f", "e", "b", "c", "g", "d"}, list.getTxHashesAsStrings())
}

func TestListForSender_AddTx_IgnoresDuplicates(t *testing.T) {
	list := newUnconstrainedListToTest()

	added, _ := list.AddTx(createTx([]byte("tx1"), ".", 1))
	require.True(t, added)
	added, _ = list.AddTx(createTx([]byte("tx2"), ".", 2))
	require.True(t, added)
	added, _ = list.AddTx(createTx([]byte("tx3"), ".", 3))
	require.True(t, added)
	added, _ = list.AddTx(createTx([]byte("tx2"), ".", 2))
	require.False(t, added)
}

func TestListForSender_AddTx_AppliesSizeConstraintsForNumTransactions(t *testing.T) {
	list := newListToTest(math.MaxUint32, 3)

	list.AddTx(createTx([]byte("tx1"), ".", 1))
	list.AddTx(createTx([]byte("tx5"), ".", 5))
	list.AddTx(createTx([]byte("tx4"), ".", 4))
	list.AddTx(createTx([]byte("tx2"), ".", 2))
	require.Equal(t, []string{"tx1", "tx2", "tx4"}, list.getTxHashesAsStrings())

	_, evicted := list.AddTx(createTx([]byte("tx3"), ".", 3))
	require.Equal(t, []string{"tx1", "tx2", "tx3"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx4"}, hashesAsStrings(evicted))

	// Gives priority to higher gas - though undesirable to some extent, "tx3" is evicted
	_, evicted = list.AddTx(createTx([]byte("tx2++"), ".", 2).withGasPrice(1.5 * oneBillion))
	require.Equal(t, []string{"tx1", "tx2++", "tx2"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx3"}, hashesAsStrings(evicted))

	// Though undesirable to some extent, "tx3++"" is added, then evicted
	_, evicted = list.AddTx(createTx([]byte("tx3++"), ".", 3).withGasPrice(1.5 * oneBillion))
	require.Equal(t, []string{"tx1", "tx2++", "tx2"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx3++"}, hashesAsStrings(evicted))
}

func TestListForSender_AddTx_AppliesSizeConstraintsForNumBytes(t *testing.T) {
	list := newListToTest(1024, math.MaxUint32)

	list.AddTx(createTx([]byte("tx1"), ".", 1).withSize(128).withGasLimit(50000))
	list.AddTx(createTx([]byte("tx2"), ".", 2).withSize(512).withGasLimit(1500000))
	list.AddTx(createTx([]byte("tx3"), ".", 3).withSize(256).withGasLimit(1500000))
	_, evicted := list.AddTx(createTx([]byte("tx5"), ".", 4).withSize(256).withGasLimit(1500000))
	require.Equal(t, []string{"tx1", "tx2", "tx3"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx5"}, hashesAsStrings(evicted))

	_, evicted = list.AddTx(createTx([]byte("tx5--"), ".", 4).withSize(128).withGasLimit(50000))
	require.Equal(t, []string{"tx1", "tx2", "tx3", "tx5--"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{}, hashesAsStrings(evicted))

	_, evicted = list.AddTx(createTx([]byte("tx4"), ".", 4).withSize(128).withGasLimit(50000))
	require.Equal(t, []string{"tx1", "tx2", "tx3", "tx4"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx5--"}, hashesAsStrings(evicted))

	// Gives priority to higher gas - though undesirably to some extent, "tx4" is evicted
	_, evicted = list.AddTx(createTx([]byte("tx3++"), ".", 3).withSize(256).withGasLimit(1500000).withGasPrice(1.5 * oneBillion))
	require.Equal(t, []string{"tx1", "tx2", "tx3++", "tx3"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx4"}, hashesAsStrings(evicted))
}

func TestListForSender_findTx(t *testing.T) {
	list := newUnconstrainedListToTest()

	txA := createTx([]byte("A"), ".", 41)
	txANewer := createTx([]byte("ANewer"), ".", 41)
	txB := createTx([]byte("B"), ".", 42)
	txD := createTx([]byte("none"), ".", 43)
	list.AddTx(txA)
	list.AddTx(txANewer)
	list.AddTx(txB)

	elementWithA := list.findListElementWithTx(txA)
	elementWithANewer := list.findListElementWithTx(txANewer)
	elementWithB := list.findListElementWithTx(txB)
	noElementWithD := list.findListElementWithTx(txD)

	require.NotNil(t, elementWithA)
	require.NotNil(t, elementWithANewer)
	require.NotNil(t, elementWithB)

	require.Equal(t, txA, elementWithA.Value.(*WrappedTransaction))
	require.Equal(t, txANewer, elementWithANewer.Value.(*WrappedTransaction))
	require.Equal(t, txB, elementWithB.Value.(*WrappedTransaction))
	require.Nil(t, noElementWithD)
}

func TestListForSender_findTx_CoverNonceComparisonOptimization(t *testing.T) {
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("A"), ".", 42))

	// Find one with a lower nonce, not added to cache
	noElement := list.findListElementWithTx(createTx(nil, ".", 41))
	require.Nil(t, noElement)
}

func TestListForSender_RemoveTransaction(t *testing.T) {
	list := newUnconstrainedListToTest()
	tx := createTx([]byte("a"), ".", 1)

	list.AddTx(tx)
	require.Equal(t, 1, list.items.Len())

	list.RemoveTx(tx)
	require.Equal(t, 0, list.items.Len())
}

func TestListForSender_RemoveTransaction_NoPanicWhenTxMissing(t *testing.T) {
	list := newUnconstrainedListToTest()
	tx := createTx([]byte(""), ".", 1)

	list.RemoveTx(tx)
	require.Equal(t, 0, list.items.Len())
}

func TestListForSender_NotifyAccountNonce(t *testing.T) {
	list := newUnconstrainedListToTest()

	require.Equal(t, uint64(0), list.accountNonce.Get())
	require.False(t, list.accountNonceKnown.IsSet())

	list.notifyAccountNonceReturnEvictedTransactions(42)

	require.Equal(t, uint64(42), list.accountNonce.Get())
	require.True(t, list.accountNonceKnown.IsSet())
}

func TestListForSender_evictTransactionsWithLowerNoncesNoLock(t *testing.T) {
	list := newUnconstrainedListToTest()

	list.AddTx(createTx([]byte("tx-42"), ".", 42))
	list.AddTx(createTx([]byte("tx-43"), ".", 43))
	list.AddTx(createTx([]byte("tx-44"), ".", 44))
	list.AddTx(createTx([]byte("tx-45"), ".", 45))

	require.Equal(t, 4, list.items.Len())

	list.evictTransactionsWithLowerNoncesNoLockReturnEvicted(43)
	require.Equal(t, 3, list.items.Len())

	list.evictTransactionsWithLowerNoncesNoLockReturnEvicted(44)
	require.Equal(t, 2, list.items.Len())

	list.evictTransactionsWithLowerNoncesNoLockReturnEvicted(99)
	require.Equal(t, 0, list.items.Len())
}

func TestListForSender_getTxs(t *testing.T) {
	list := newUnconstrainedListToTest()
	list.notifyAccountNonceReturnEvictedTransactions(42)

	// No transaction, no gap
	require.Len(t, list.getTxs(), 0)
	require.Len(t, list.getTxsReversed(), 0)
	require.Len(t, list.getSequentialTxs(), 0)

	// One gap
	list.AddTx(createTx([]byte("tx-43"), ".", 43))
	require.Len(t, list.getTxs(), 1)
	require.Len(t, list.getTxsReversed(), 1)
	require.Len(t, list.getSequentialTxs(), 0)

	// Resolve gap
	list.AddTx(createTx([]byte("tx-42"), ".", 42))
	require.Len(t, list.getTxs(), 2)
	require.Len(t, list.getTxsReversed(), 2)
	require.Len(t, list.getSequentialTxs(), 2)

	require.Equal(t, []byte("tx-42"), list.getTxs()[0].TxHash)
	require.Equal(t, []byte("tx-43"), list.getTxs()[1].TxHash)
	require.Equal(t, list.getTxs(), list.getSequentialTxs())

	require.Equal(t, []byte("tx-43"), list.getTxsReversed()[0].TxHash)
	require.Equal(t, []byte("tx-42"), list.getTxsReversed()[1].TxHash)

	// With nonce duplicates
	list.AddTx(createTx([]byte("tx-42++"), ".", 42).withGasPrice(1.1 * oneBillion))
	list.AddTx(createTx([]byte("tx-43++"), ".", 43).withGasPrice(1.1 * oneBillion))
	require.Len(t, list.getTxs(), 4)
	require.Len(t, list.getTxsReversed(), 4)
	require.Len(t, list.getSequentialTxs(), 2)

	require.Equal(t, []byte("tx-42++"), list.getSequentialTxs()[0].TxHash)
	require.Equal(t, []byte("tx-43++"), list.getSequentialTxs()[1].TxHash)

	require.Equal(t, []byte("tx-42++"), list.getTxs()[0].TxHash)
	require.Equal(t, []byte("tx-42"), list.getTxs()[1].TxHash)
	require.Equal(t, []byte("tx-43++"), list.getTxs()[2].TxHash)
	require.Equal(t, []byte("tx-43"), list.getTxs()[3].TxHash)

	require.Equal(t, []byte("tx-43"), list.getTxsReversed()[0].TxHash)
	require.Equal(t, []byte("tx-43++"), list.getTxsReversed()[1].TxHash)
	require.Equal(t, []byte("tx-42"), list.getTxsReversed()[2].TxHash)
	require.Equal(t, []byte("tx-42++"), list.getTxsReversed()[3].TxHash)
}

func TestListForSender_DetectRaceConditions(t *testing.T) {
	list := newUnconstrainedListToTest()

	wg := sync.WaitGroup{}

	doOperations := func() {
		// These might be called concurrently:
		_ = list.IsEmpty()
		_ = list.getTxs()
		_ = list.getTxsReversed()
		_ = list.getSequentialTxs()
		_ = list.countTxWithLock()
		_ = list.notifyAccountNonceReturnEvictedTransactions(42)
		_, _ = list.AddTx(createTx([]byte("test"), ".", 42))

		wg.Done()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go doOperations()
	}

	wg.Wait()
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
