package txcache

import (
	"math"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/stretchr/testify/require"
)

func TestListForSender_AddTx_Sorts(t *testing.T) {
	list := newListToTest()

	list.AddTx([]byte("a"), createTx(".", 1))
	list.AddTx([]byte("c"), createTx(".", 3))
	list.AddTx([]byte("d"), createTx(".", 4))
	list.AddTx([]byte("b"), createTx(".", 2))

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

	txA := createTx(".", 41)
	txANewer := createTx(".", 41)
	txB := createTx(".", 42)
	txD := createTx(".", 43)
	list.AddTx([]byte("A"), txA)
	list.AddTx([]byte("ANewer"), txANewer)
	list.AddTx([]byte("B"), txB)

	elementWithA := list.findListElementWithTx(txA)
	elementWithANewer := list.findListElementWithTx(txANewer)
	elementWithB := list.findListElementWithTx(txB)
	noElementWithD := list.findListElementWithTx(txD)

	require.Equal(t, txA, elementWithA.Value.(txListForSenderNode).tx)
	require.Equal(t, txANewer, elementWithANewer.Value.(txListForSenderNode).tx)
	require.Equal(t, txB, elementWithB.Value.(txListForSenderNode).tx)
	require.Nil(t, noElementWithD)
}

func TestListForSender_findTx_CoverNonceComparisonOptimization(t *testing.T) {
	list := newListToTest()
	list.AddTx([]byte("A"), createTx(".", 42))

	// Find one with a lower nonce, not added to cache
	noElement := list.findListElementWithTx(createTx(".", 41))
	require.Nil(t, noElement)
}

func TestListForSender_RemoveTransaction(t *testing.T) {
	list := newListToTest()
	tx := createTx(".", 1)

	list.AddTx([]byte("a"), tx)
	require.Equal(t, 1, list.items.Len())

	list.RemoveTx(tx)
	require.Equal(t, 0, list.items.Len())
}

func TestListForSender_RemoveTransaction_NoPanicWhenTxMissing(t *testing.T) {
	list := newListToTest()
	tx := createTx(".", 1)

	list.RemoveTx(tx)
	require.Equal(t, 0, list.items.Len())
}

func TestListForSender_RemoveHighNonceTransactions(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
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
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	list.RemoveHighNonceTxs(0)
	require.Equal(t, 100, list.items.Len())

	list.RemoveHighNonceTxs(500)
	require.Equal(t, 0, list.items.Len())
}

func TestListForSender_SelectBatchTo(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	destination := make([]data.TransactionHandler, 1000)
	destinationHashes := make([][]byte, 1000)

	// First batch
	copied := list.selectBatchTo(true, destination, destinationHashes, 50)
	require.Equal(t, 50, copied)
	require.NotNil(t, destination[49])
	require.Nil(t, destination[50])

	// Second batch
	copied = list.selectBatchTo(false, destination[50:], destinationHashes[50:], 50)
	require.Equal(t, 50, copied)
	require.NotNil(t, destination[99])

	// No third batch
	copied = list.selectBatchTo(false, destination, destinationHashes, 50)
	require.Equal(t, 0, copied)

	// Restart copy
	copied = list.selectBatchTo(true, destination, destinationHashes, 12345)
	require.Equal(t, 100, copied)
}

func TestListForSender_SelectBatchTo_NoPanicWhenCornerCases(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	// When empty destination
	destination := make([]data.TransactionHandler, 0)
	destinationHashes := make([][]byte, 0)
	copied := list.selectBatchTo(true, destination, destinationHashes, 10)
	require.Equal(t, 0, copied)

	// When small destination
	destination = make([]data.TransactionHandler, 5)
	destinationHashes = make([][]byte, 5)
	copied = list.selectBatchTo(false, destination, destinationHashes, 10)
	require.Equal(t, 5, copied)
}

func TestListForSender_SelectBatchTo_WhenInitialGap(t *testing.T) {
	list := newListToTest()

	list.notifyAccountNonce(1)

	for index := 10; index < 20; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	destination := make([]data.TransactionHandler, 1000)
	destinationHashes := make([][]byte, 1000)

	// First batch, first failure
	copied := list.selectBatchTo(true, destination, destinationHashes, 50)
	require.Equal(t, 0, copied)
	require.Nil(t, destination[0])
	require.Equal(t, int64(1), list.numFailedSelections.Get())

	// Second batch, don't count failure again
	copied = list.selectBatchTo(false, destination, destinationHashes, 50)
	require.Equal(t, 0, copied)
	require.Nil(t, destination[0])
	require.Equal(t, int64(1), list.numFailedSelections.Get())

	// First batch again, second failure
	copied = list.selectBatchTo(true, destination, destinationHashes, 50)
	require.Equal(t, 0, copied)
	require.Nil(t, destination[0])
	require.Equal(t, int64(2), list.numFailedSelections.Get())
}

func TestListForSender_SelectBatchTo_WhenGracePeriodWithGapResolve(t *testing.T) {
	list := newListToTest()

	list.notifyAccountNonce(1)

	for index := 2; index < 20; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	destination := make([]data.TransactionHandler, 1000)
	destinationHashes := make([][]byte, 1000)

	// Try a number of selections with failure, reach close to grace period
	for i := 1; i < gracePeriodLowerBound; i++ {
		copied := list.selectBatchTo(true, destination, destinationHashes, math.MaxInt32)
		require.Equal(t, 0, copied)
		require.Equal(t, int64(i), list.numFailedSelections.Get())
	}

	// Try selection again. Failure will move the sender to grace period and return 1 transaction
	copied := list.selectBatchTo(true, destination, destinationHashes, math.MaxInt32)
	require.Equal(t, 1, copied)
	require.Equal(t, int64(gracePeriodLowerBound), list.numFailedSelections.Get())

	// Try a new selection during the grace period
	copied = list.selectBatchTo(true, destination, destinationHashes, math.MaxInt32)
	require.Equal(t, 1, copied)
	require.Equal(t, int64(gracePeriodLowerBound+1), list.numFailedSelections.Get())

	// Now resolve the gap
	list.AddTx([]byte("resolving-tx"), createTx(".", 1))
	// Selection will be successful
	copied = list.selectBatchTo(true, destination, destinationHashes, math.MaxInt32)
	require.Equal(t, 19, copied)
	require.Equal(t, int64(0), list.numFailedSelections.Get())
}

func TestListForSender_SelectBatchTo_WhenGracePeriodWithNoGapResolve(t *testing.T) {
	list := newListToTest()

	list.notifyAccountNonce(1)

	for index := 2; index < 20; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	destination := make([]data.TransactionHandler, 1000)
	destinationHashes := make([][]byte, 1000)

	// Try a number of selections with failure, reach close to grace period
	for i := 1; i < gracePeriodLowerBound; i++ {
		copied := list.selectBatchTo(true, destination, destinationHashes, math.MaxInt32)
		require.Equal(t, 0, copied)
		require.Equal(t, int64(i), list.numFailedSelections.Get())
	}

	// Try a number of selections with failure, within the grace period
	for i := gracePeriodLowerBound; i <= gracePeriodUpperBound; i++ {
		copied := list.selectBatchTo(true, destination, destinationHashes, math.MaxInt32)
		require.Equal(t, 1, copied)
		require.Equal(t, int64(i), list.numFailedSelections.Get())
	}

	// Grace period exceeded now
	copied := list.selectBatchTo(true, destination, destinationHashes, math.MaxInt32)
	require.Equal(t, 0, copied)
	require.Equal(t, int64(gracePeriodUpperBound+1), list.numFailedSelections.Get())
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
	list.AddTx([]byte("tx-44"), createTx(".", 43))
	require.True(t, list.hasInitialGap())
	// Resolve gap
	list.AddTx([]byte("tx-44"), createTx(".", 42))
	require.False(t, list.hasInitialGap())
}

func TestListForSender_getTxHashes(t *testing.T) {
	list := newListToTest()
	require.Len(t, list.getTxHashes(), 0)

	list.AddTx([]byte("A"), createTx(".", 1))
	require.Len(t, list.getTxHashes(), 1)

	list.AddTx([]byte("B"), createTx(".", 2))
	list.AddTx([]byte("C"), createTx(".", 3))
	require.Len(t, list.getTxHashes(), 3)
}

func TestListForSender_DetectRaceConditions(t *testing.T) {
	list := newListToTest()

	go func() {
		// This is called during eviction
		approximatelyCountTxInLists([]*txListForSender{list})
	}()

	go func() {
		list.AddTx([]byte("test"), createTx(".", 42))
	}()
}

func newListToTest() *txListForSender {
	return newTxListForSender(".", &CacheConfig{MinGasPriceMicroErd: 100})
}
