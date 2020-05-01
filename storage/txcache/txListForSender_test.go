package txcache

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListForSender_AddTx_Sorts(t *testing.T) {
	list := newListToTest()

	list.AddTx(createTx([]byte("a"), ".", 1))
	list.AddTx(createTx([]byte("c"), ".", 3))
	list.AddTx(createTx([]byte("d"), ".", 4))
	list.AddTx(createTx([]byte("b"), ".", 2))

	txHashes := list.getTxHashes()

	require.Equal(t, 4, list.items.Len())
	require.Equal(t, 4, len(txHashes))

	require.Equal(t, []byte("a"), txHashes[0])
	require.Equal(t, []byte("b"), txHashes[1])
	require.Equal(t, []byte("c"), txHashes[2])
	require.Equal(t, []byte("d"), txHashes[3])
}

func TestListForSender_findTx(t *testing.T) {
	list := newListToTest()

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

	require.Equal(t, txA, elementWithA.Value.(*WrappedTransaction))
	require.Equal(t, txANewer, elementWithANewer.Value.(*WrappedTransaction))
	require.Equal(t, txB, elementWithB.Value.(*WrappedTransaction))
	require.Nil(t, noElementWithD)
}

func TestListForSender_findTx_CoverNonceComparisonOptimization(t *testing.T) {
	list := newListToTest()
	list.AddTx(createTx([]byte("A"), ".", 42))

	// Find one with a lower nonce, not added to cache
	noElement := list.findListElementWithTx(createTx(nil, ".", 41))
	require.Nil(t, noElement)
}

func TestListForSender_RemoveTransaction(t *testing.T) {
	list := newListToTest()
	tx := createTx([]byte("a"), ".", 1)

	list.AddTx(tx)
	require.Equal(t, 1, list.items.Len())

	list.RemoveTx(tx)
	require.Equal(t, 0, list.items.Len())
}

func TestListForSender_RemoveTransaction_NoPanicWhenTxMissing(t *testing.T) {
	list := newListToTest()
	tx := createTx([]byte(""), ".", 1)

	list.RemoveTx(tx)
	require.Equal(t, 0, list.items.Len())
}

func TestListForSender_RemoveHighNonceTransactions(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)))
	}

	list.RemoveHighNonceTxs(50)
	require.Equal(t, 50, list.items.Len())

	list.RemoveHighNonceTxs(20)
	require.Equal(t, 30, list.items.Len())

	list.RemoveHighNonceTxs(30)
	require.Equal(t, 0, list.items.Len())
}

func TestListForSender_RemoveHighNonceTransactions_NoPanicWhenCornerCases(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)))
	}

	list.RemoveHighNonceTxs(0)
	require.Equal(t, 100, list.items.Len())

	list.RemoveHighNonceTxs(500)
	require.Equal(t, 0, list.items.Len())
}

func TestListForSender_SelectBatchTo(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)))
	}

	destination := make([]*WrappedTransaction, 1000)

	// First batch
	copied := list.selectBatchTo(true, destination, 50)
	require.Equal(t, 50, copied)
	require.NotNil(t, destination[49])
	require.Nil(t, destination[50])

	// Second batch
	copied = list.selectBatchTo(false, destination[50:], 50)
	require.Equal(t, 50, copied)
	require.NotNil(t, destination[99])

	// No third batch
	copied = list.selectBatchTo(false, destination, 50)
	require.Equal(t, 0, copied)

	// Restart copy
	copied = list.selectBatchTo(true, destination, 12345)
	require.Equal(t, 100, copied)
}

func TestListForSender_SelectBatchTo_NoPanicWhenCornerCases(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)))
	}

	// When empty destination
	destination := make([]*WrappedTransaction, 0)
	copied := list.selectBatchTo(true, destination, 10)
	require.Equal(t, 0, copied)

	// When small destination
	destination = make([]*WrappedTransaction, 5)
	copied = list.selectBatchTo(false, destination, 10)
	require.Equal(t, 5, copied)
}

func TestListForSender_SelectBatchTo_WhenInitialGap(t *testing.T) {
	list := newListToTest()

	list.notifyAccountNonce(1)

	for index := 10; index < 20; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)))
	}

	destination := make([]*WrappedTransaction, 1000)

	// First batch of selection, first failure
	copied := list.selectBatchTo(true, destination, 50)
	require.Equal(t, 0, copied)
	require.Nil(t, destination[0])
	require.Equal(t, int64(1), list.numFailedSelections.Get())

	// Second batch of selection, don't count failure again
	copied = list.selectBatchTo(false, destination, 50)
	require.Equal(t, 0, copied)
	require.Nil(t, destination[0])
	require.Equal(t, int64(1), list.numFailedSelections.Get())

	// First batch of another selection, second failure, enters grace period
	copied = list.selectBatchTo(true, destination, 50)
	require.Equal(t, 1, copied)
	require.NotNil(t, destination[0])
	require.Nil(t, destination[1])
	require.Equal(t, int64(2), list.numFailedSelections.Get())
}

func TestListForSender_SelectBatchTo_WhenGracePeriodWithGapResolve(t *testing.T) {
	list := newListToTest()

	list.notifyAccountNonce(1)

	for index := 2; index < 20; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)))
	}

	destination := make([]*WrappedTransaction, 1000)

	// Try a number of selections with failure, reach close to grace period
	for i := 1; i < senderGracePeriodLowerBound; i++ {
		copied := list.selectBatchTo(true, destination, math.MaxInt32)
		require.Equal(t, 0, copied)
		require.Equal(t, int64(i), list.numFailedSelections.Get())
	}

	// Try selection again. Failure will move the sender to grace period and return 1 transaction
	copied := list.selectBatchTo(true, destination, math.MaxInt32)
	require.Equal(t, 1, copied)
	require.Equal(t, int64(senderGracePeriodLowerBound), list.numFailedSelections.Get())
	require.False(t, list.sweepable.IsSet())

	// Now resolve the gap
	list.AddTx(createTx([]byte("resolving-tx"), ".", 1))
	// Selection will be successful
	copied = list.selectBatchTo(true, destination, math.MaxInt32)
	require.Equal(t, 19, copied)
	require.Equal(t, int64(0), list.numFailedSelections.Get())
	require.False(t, list.sweepable.IsSet())
}

func TestListForSender_SelectBatchTo_WhenGracePeriodWithNoGapResolve(t *testing.T) {
	list := newListToTest()

	list.notifyAccountNonce(1)

	for index := 2; index < 20; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)))
	}

	destination := make([]*WrappedTransaction, 1000)

	// Try a number of selections with failure, reach close to grace period
	for i := 1; i < senderGracePeriodLowerBound; i++ {
		copied := list.selectBatchTo(true, destination, math.MaxInt32)
		require.Equal(t, 0, copied)
		require.Equal(t, int64(i), list.numFailedSelections.Get())
	}

	// Try a number of selections with failure, within the grace period
	for i := senderGracePeriodLowerBound; i <= senderGracePeriodUpperBound; i++ {
		copied := list.selectBatchTo(true, destination, math.MaxInt32)
		require.Equal(t, 1, copied)
		require.Equal(t, int64(i), list.numFailedSelections.Get())
	}

	// Grace period exceeded now
	copied := list.selectBatchTo(true, destination, math.MaxInt32)
	require.Equal(t, 0, copied)
	require.Equal(t, int64(senderGracePeriodUpperBound+1), list.numFailedSelections.Get())
	require.True(t, list.sweepable.IsSet())
}

func TestListForSender_NotifyAccountNonce(t *testing.T) {
	list := newListToTest()

	require.Equal(t, uint64(0), list.accountNonce.Get())
	require.False(t, list.accountNonceKnown.IsSet())

	list.notifyAccountNonce(42)

	require.Equal(t, uint64(42), list.accountNonce.Get())
	require.True(t, list.accountNonceKnown.IsSet())
}

func TestListForSender_hasInitialGap(t *testing.T) {
	list := newListToTest()
	list.notifyAccountNonce(42)

	// No transaction, no gap
	require.False(t, list.hasInitialGap())
	// One gap
	list.AddTx(createTx([]byte("tx-44"), ".", 43))
	require.True(t, list.hasInitialGap())
	// Resolve gap
	list.AddTx(createTx([]byte("tx-44"), ".", 42))
	require.False(t, list.hasInitialGap())
}

func TestListForSender_getTxHashes(t *testing.T) {
	list := newListToTest()
	require.Len(t, list.getTxHashes(), 0)

	list.AddTx(createTx([]byte("A"), ".", 1))
	require.Len(t, list.getTxHashes(), 1)

	list.AddTx(createTx([]byte("B"), ".", 2))
	list.AddTx(createTx([]byte("C"), ".", 3))
	require.Len(t, list.getTxHashes(), 3)
}

func TestListForSender_DetectRaceConditions(t *testing.T) {
	list := newListToTest()

	go func() {
		// These are called concurrently with addition: during eviction, during removal etc.
		approximatelyCountTxInLists([]*txListForSender{list})
		list.HasMoreThan(42)
		list.IsEmpty()
	}()

	go func() {
		list.AddTx(createTx([]byte("test"), ".", 42))
	}()
}

func newListToTest() *txListForSender {
	return newTxListForSender(".", &CacheConfig{MinGasPriceMicroErd: 100})
}
