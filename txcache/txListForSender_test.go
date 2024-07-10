package txcache

import (
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestListForSender_AddTx_Sorts(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	list.AddTx(createTx([]byte("a"), ".", 1), txGasHandler)
	list.AddTx(createTx([]byte("c"), ".", 3), txGasHandler)
	list.AddTx(createTx([]byte("d"), ".", 4), txGasHandler)
	list.AddTx(createTx([]byte("b"), ".", 2), txGasHandler)

	require.Equal(t, []string{"a", "b", "c", "d"}, list.getTxHashesAsStrings())
}

func TestListForSender_AddTx_GivesPriorityToHigherGas(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	list.AddTx(createTx([]byte("a"), ".", 1), txGasHandler)
	list.AddTx(createTx([]byte("b"), ".", 3).withGasPrice(1.2*oneBillion), txGasHandler)
	list.AddTx(createTx([]byte("c"), ".", 3).withGasPrice(1.1*oneBillion), txGasHandler)
	list.AddTx(createTx([]byte("d"), ".", 2), txGasHandler)
	list.AddTx(createTx([]byte("e"), ".", 3).withGasPrice(1.3*oneBillion), txGasHandler)

	require.Equal(t, []string{"a", "d", "e", "b", "c"}, list.getTxHashesAsStrings())
}

func TestListForSender_AddTx_SortsCorrectlyWhenSameNonceSamePrice(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	list.AddTx(createTx([]byte("a"), ".", 1).withGasPrice(oneBillion), txGasHandler)
	list.AddTx(createTx([]byte("b"), ".", 3).withGasPrice(3*oneBillion), txGasHandler)
	list.AddTx(createTx([]byte("c"), ".", 3).withGasPrice(3*oneBillion), txGasHandler)
	list.AddTx(createTx([]byte("d"), ".", 3).withGasPrice(2*oneBillion), txGasHandler)
	list.AddTx(createTx([]byte("e"), ".", 3).withGasPrice(3.5*oneBillion), txGasHandler)
	list.AddTx(createTx([]byte("f"), ".", 2).withGasPrice(oneBillion), txGasHandler)
	list.AddTx(createTx([]byte("g"), ".", 3).withGasPrice(2.5*oneBillion), txGasHandler)

	// In case of same-nonce, same-price transactions, the newer one has priority
	require.Equal(t, []string{"a", "f", "e", "b", "c", "g", "d"}, list.getTxHashesAsStrings())
}

func TestListForSender_AddTx_IgnoresDuplicates(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	added, _ := list.AddTx(createTx([]byte("tx1"), ".", 1), txGasHandler)
	require.True(t, added)
	added, _ = list.AddTx(createTx([]byte("tx2"), ".", 2), txGasHandler)
	require.True(t, added)
	added, _ = list.AddTx(createTx([]byte("tx3"), ".", 3), txGasHandler)
	require.True(t, added)
	added, _ = list.AddTx(createTx([]byte("tx2"), ".", 2), txGasHandler)
	require.False(t, added)
}

func TestListForSender_AddTx_AppliesSizeConstraintsForNumTransactions(t *testing.T) {
	list := newListToTest(math.MaxUint32, 3)
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	list.AddTx(createTx([]byte("tx1"), ".", 1), txGasHandler)
	list.AddTx(createTx([]byte("tx5"), ".", 5), txGasHandler)
	list.AddTx(createTx([]byte("tx4"), ".", 4), txGasHandler)
	list.AddTx(createTx([]byte("tx2"), ".", 2), txGasHandler)
	require.Equal(t, []string{"tx1", "tx2", "tx4"}, list.getTxHashesAsStrings())

	_, evicted := list.AddTx(createTx([]byte("tx3"), ".", 3), txGasHandler)
	require.Equal(t, []string{"tx1", "tx2", "tx3"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx4"}, hashesAsStrings(evicted))

	// Gives priority to higher gas - though undesirably to some extent, "tx3" is evicted
	_, evicted = list.AddTx(createTx([]byte("tx2++"), ".", 2).withGasPrice(1.5*oneBillion), txGasHandler)
	require.Equal(t, []string{"tx1", "tx2++", "tx2"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx3"}, hashesAsStrings(evicted))

	// Though Undesirably to some extent, "tx3++"" is added, then evicted
	_, evicted = list.AddTx(createTx([]byte("tx3++"), ".", 3).withGasPrice(1.5*oneBillion), txGasHandler)
	require.Equal(t, []string{"tx1", "tx2++", "tx2"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx3++"}, hashesAsStrings(evicted))
}

func TestListForSender_AddTx_AppliesSizeConstraintsForNumBytes(t *testing.T) {
	list := newListToTest(1024, math.MaxUint32)
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	list.AddTx(createTx([]byte("tx1"), ".", 1).withSize(128).withGasLimit(50000), txGasHandler)
	list.AddTx(createTx([]byte("tx2"), ".", 2).withSize(512).withGasLimit(1500000), txGasHandler)
	list.AddTx(createTx([]byte("tx3"), ".", 3).withSize(256).withGasLimit(1500000), txGasHandler)
	_, evicted := list.AddTx(createTx([]byte("tx5"), ".", 4).withSize(256).withGasLimit(1500000), txGasHandler)
	require.Equal(t, []string{"tx1", "tx2", "tx3"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx5"}, hashesAsStrings(evicted))

	_, evicted = list.AddTx(createTx([]byte("tx5--"), ".", 4).withSize(128).withGasLimit(50000), txGasHandler)
	require.Equal(t, []string{"tx1", "tx2", "tx3", "tx5--"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{}, hashesAsStrings(evicted))

	_, evicted = list.AddTx(createTx([]byte("tx4"), ".", 4).withSize(128).withGasLimit(50000), txGasHandler)
	require.Equal(t, []string{"tx1", "tx2", "tx3", "tx4"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx5--"}, hashesAsStrings(evicted))

	// Gives priority to higher gas - though undesirably to some extent, "tx4" is evicted
	_, evicted = list.AddTx(createTx([]byte("tx3++"), ".", 3).withSize(256).withGasLimit(1500000).withGasPrice(1.5*oneBillion), txGasHandler)
	require.Equal(t, []string{"tx1", "tx2", "tx3++", "tx3"}, list.getTxHashesAsStrings())
	require.Equal(t, []string{"tx4"}, hashesAsStrings(evicted))
}

func TestListForSender_findTx(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	txA := createTx([]byte("A"), ".", 41)
	txANewer := createTx([]byte("ANewer"), ".", 41)
	txB := createTx([]byte("B"), ".", 42)
	txD := createTx([]byte("none"), ".", 43)
	list.AddTx(txA, txGasHandler)
	list.AddTx(txANewer, txGasHandler)
	list.AddTx(txB, txGasHandler)

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
	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	list.AddTx(createTx([]byte("A"), ".", 42), txGasHandler)

	// Find one with a lower nonce, not added to cache
	noElement := list.findListElementWithTx(createTx(nil, ".", 41))
	require.Nil(t, noElement)
}

func TestListForSender_RemoveTransaction(t *testing.T) {
	list := newUnconstrainedListToTest()
	tx := createTx([]byte("a"), ".", 1)
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	list.AddTx(tx, txGasHandler)
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

func TestListForSender_SelectBatchTo(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	for index := 0; index < 100; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)), txGasHandler)
	}

	destination := make([]*WrappedTransaction, 1000)

	// First batch
	journal := list.selectBatchTo(true, destination, 50, math.MaxUint64)
	require.Equal(t, 50, journal.selectedNum)
	require.NotNil(t, destination[49])
	require.Nil(t, destination[50])

	// Second batch
	journal = list.selectBatchTo(false, destination[50:], 50, math.MaxUint64)
	require.Equal(t, 50, journal.selectedNum)
	require.NotNil(t, destination[99])

	// No third batch
	journal = list.selectBatchTo(false, destination, 50, math.MaxUint64)
	require.Equal(t, 0, journal.selectedNum)

	// Restart copy
	journal = list.selectBatchTo(true, destination, 12345, math.MaxUint64)
	require.Equal(t, 100, journal.selectedNum)
}

func TestListForSender_SelectBatchToWithLimitedGasBandwidth(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	for index := 0; index < 40; index++ {
		wtx := createTx([]byte{byte(index)}, ".", uint64(index))
		tx, _ := wtx.Tx.(*transaction.Transaction)
		tx.GasLimit = 1000000
		list.AddTx(wtx, txGasHandler)
	}

	destination := make([]*WrappedTransaction, 1000)

	// First batch
	journal := list.selectBatchTo(true, destination, 50, 500000)
	require.Equal(t, 1, journal.selectedNum)
	require.NotNil(t, destination[0])
	require.Nil(t, destination[1])

	// Second batch
	journal = list.selectBatchTo(false, destination[1:], 50, 20000000)
	require.Equal(t, 20, journal.selectedNum)
	require.NotNil(t, destination[20])
	require.Nil(t, destination[21])

	// third batch
	journal = list.selectBatchTo(false, destination[21:], 20, math.MaxUint64)
	require.Equal(t, 19, journal.selectedNum)

	// Restart copy
	journal = list.selectBatchTo(true, destination[41:], 12345, math.MaxUint64)
	require.Equal(t, 40, journal.selectedNum)
}

func TestListForSender_SelectBatchTo_NoPanicWhenCornerCases(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	for index := 0; index < 100; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)), txGasHandler)
	}

	// When empty destination
	destination := make([]*WrappedTransaction, 0)
	journal := list.selectBatchTo(true, destination, 10, math.MaxUint64)
	require.Equal(t, 0, journal.selectedNum)

	// When small destination
	destination = make([]*WrappedTransaction, 5)
	journal = list.selectBatchTo(false, destination, 10, math.MaxUint64)
	require.Equal(t, 5, journal.selectedNum)
}

func TestListForSender_SelectBatchTo_WhenInitialGap(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	list.notifyAccountNonce(1)

	for index := 10; index < 20; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)), txGasHandler)
	}

	destination := make([]*WrappedTransaction, 1000)

	// First batch of selection, first failure
	journal := list.selectBatchTo(true, destination, 50, math.MaxUint64)
	require.Equal(t, 0, journal.selectedNum)
	require.Nil(t, destination[0])
	require.Equal(t, int64(1), list.numFailedSelections.Get())

	// Second batch of selection, don't count failure again
	journal = list.selectBatchTo(false, destination, 50, math.MaxUint64)
	require.Equal(t, 0, journal.selectedNum)
	require.Nil(t, destination[0])
	require.Equal(t, int64(1), list.numFailedSelections.Get())

	// First batch of another selection, second failure, enters grace period
	journal = list.selectBatchTo(true, destination, 50, math.MaxUint64)
	require.Equal(t, 1, journal.selectedNum)
	require.NotNil(t, destination[0])
	require.Nil(t, destination[1])
	require.Equal(t, int64(2), list.numFailedSelections.Get())
}

func TestListForSender_SelectBatchTo_WhenGracePeriodWithGapResolve(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	list.notifyAccountNonce(1)

	for index := 2; index < 20; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)), txGasHandler)
	}

	destination := make([]*WrappedTransaction, 1000)

	// Try a number of selections with failure, reach close to grace period
	for i := 1; i < senderGracePeriodLowerBound; i++ {
		journal := list.selectBatchTo(true, destination, math.MaxInt32, math.MaxUint64)
		require.Equal(t, 0, journal.selectedNum)
		require.Equal(t, int64(i), list.numFailedSelections.Get())
	}

	// Try selection again. Failure will move the sender to grace period and return 1 transaction
	journal := list.selectBatchTo(true, destination, math.MaxInt32, math.MaxUint64)
	require.Equal(t, 1, journal.selectedNum)
	require.Equal(t, int64(senderGracePeriodLowerBound), list.numFailedSelections.Get())
	require.False(t, list.sweepable.IsSet())

	// Now resolve the gap
	list.AddTx(createTx([]byte("resolving-tx"), ".", 1), txGasHandler)
	// Selection will be successful
	journal = list.selectBatchTo(true, destination, math.MaxInt32, math.MaxUint64)
	require.Equal(t, 19, journal.selectedNum)
	require.Equal(t, int64(0), list.numFailedSelections.Get())
	require.False(t, list.sweepable.IsSet())
}

func TestListForSender_SelectBatchTo_WhenGracePeriodWithNoGapResolve(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	list.notifyAccountNonce(1)

	for index := 2; index < 20; index++ {
		list.AddTx(createTx([]byte{byte(index)}, ".", uint64(index)), txGasHandler)
	}

	destination := make([]*WrappedTransaction, 1000)

	// Try a number of selections with failure, reach close to grace period
	for i := 1; i < senderGracePeriodLowerBound; i++ {
		journal := list.selectBatchTo(true, destination, math.MaxInt32, math.MaxUint64)
		require.Equal(t, 0, journal.selectedNum)
		require.Equal(t, int64(i), list.numFailedSelections.Get())
	}

	// Try a number of selections with failure, within the grace period
	for i := senderGracePeriodLowerBound; i <= senderGracePeriodUpperBound; i++ {
		journal := list.selectBatchTo(true, destination, math.MaxInt32, math.MaxUint64)
		require.Equal(t, 1, journal.selectedNum)
		require.Equal(t, int64(i), list.numFailedSelections.Get())
	}

	// Grace period exceeded now
	journal := list.selectBatchTo(true, destination, math.MaxInt32, math.MaxUint64)
	require.Equal(t, 0, journal.selectedNum)
	require.Equal(t, int64(senderGracePeriodUpperBound+1), list.numFailedSelections.Get())
	require.True(t, list.sweepable.IsSet())
}

func TestListForSender_NotifyAccountNonce(t *testing.T) {
	list := newUnconstrainedListToTest()

	require.Equal(t, uint64(0), list.accountNonce.Get())
	require.False(t, list.accountNonceKnown.IsSet())

	list.notifyAccountNonce(42)

	require.Equal(t, uint64(42), list.accountNonce.Get())
	require.True(t, list.accountNonceKnown.IsSet())
}

func TestListForSender_hasInitialGap(t *testing.T) {
	list := newUnconstrainedListToTest()
	list.notifyAccountNonce(42)
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	// No transaction, no gap
	require.False(t, list.hasInitialGap())
	// One gap
	list.AddTx(createTx([]byte("tx-43"), ".", 43), txGasHandler)
	require.True(t, list.hasInitialGap())
	// Resolve gap
	list.AddTx(createTx([]byte("tx-42"), ".", 42), txGasHandler)
	require.False(t, list.hasInitialGap())
}

func TestListForSender_getTxHashes(t *testing.T) {
	list := newUnconstrainedListToTest()
	require.Len(t, list.getTxHashes(), 0)
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	list.AddTx(createTx([]byte("A"), ".", 1), txGasHandler)
	require.Len(t, list.getTxHashes(), 1)

	list.AddTx(createTx([]byte("B"), ".", 2), txGasHandler)
	list.AddTx(createTx([]byte("C"), ".", 3), txGasHandler)
	require.Len(t, list.getTxHashes(), 3)
}

func TestListForSender_DetectRaceConditions(t *testing.T) {
	list := newUnconstrainedListToTest()
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	go func() {
		// These are called concurrently with addition: during eviction, during removal etc.
		approximatelyCountTxInLists([]*txListForSender{list})
		list.IsEmpty()
	}()

	go func() {
		list.AddTx(createTx([]byte("test"), ".", 42), txGasHandler)
	}()
}

func TestListForSender_transactionAddAndRemove_updateScore(t *testing.T) {
	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	alice := newUnconstrainedListToTest()
	bob := newUnconstrainedListToTest()

	a := createTx([]byte("a"), ".", 1)
	b := createTx([]byte("b"), ".", 1)
	c := createTx([]byte("c"), ".", 2).withDataLength(42).withGasLimit(50000 + 1500*42)
	d := createTx([]byte("d"), ".", 2).withDataLength(84).withGasLimit(50000 + 1500*84)
	e := createTx([]byte("e"), ".", 3).withDataLength(1).withGasLimit(50000000).withGasPrice(oneBillion)
	f := createTx([]byte("f"), ".", 3).withDataLength(1).withGasLimit(150000000).withGasPrice(oneBillion)
	g := createTx([]byte("g"), ".", 4).withDataLength(7).withGasLimit(5000000).withGasPrice(oneBillion)
	h := createTx([]byte("h"), ".", 4).withDataLength(7).withGasLimit(5000000).withGasPrice(oneBillion)
	i := createTx([]byte("i"), ".", 5).withDataLength(42).withGasLimit(5000000).withGasPrice(2 * oneBillion)
	j := createTx([]byte("j"), ".", 5).withDataLength(42).withGasLimit(5000000).withGasPrice(3 * oneBillion)
	k := createTx([]byte("k"), ".", 5).withDataLength(42).withGasLimit(5000000).withGasPrice(2 * oneBillion)
	l := createTx([]byte("l"), ".", 8)

	alice.AddTx(a, txGasHandler)
	bob.AddTx(b, txGasHandler)

	require.Equal(t, 74, alice.getScore())
	require.Equal(t, 74, bob.getScore())

	alice.AddTx(c, txGasHandler)
	bob.AddTx(d, txGasHandler)

	require.Equal(t, 74, alice.getScore())
	require.Equal(t, 74, bob.getScore())

	alice.AddTx(e, txGasHandler)
	bob.AddTx(f, txGasHandler)

	require.Equal(t, 5, alice.getScore())
	require.Equal(t, 2, bob.getScore())

	alice.AddTx(g, txGasHandler)
	bob.AddTx(h, txGasHandler)

	require.Equal(t, 6, alice.getScore())
	require.Equal(t, 3, bob.getScore())

	alice.AddTx(i, txGasHandler)
	bob.AddTx(j, txGasHandler)

	require.Equal(t, 10, alice.getScore())
	require.Equal(t, 6, bob.getScore())

	// Bob adds a transaction with duplicated nonce
	bob.AddTx(k, txGasHandler)

	require.Equal(t, 10, alice.getScore())
	require.Equal(t, 0, bob.getScore())

	require.True(t, alice.RemoveTx(a))
	require.True(t, alice.RemoveTx(c))

	require.Equal(t, 7, alice.getScore())
	require.Equal(t, 0, bob.getScore())

	// Alice comes with a nonce gap
	alice.AddTx(l, txGasHandler)

	require.Equal(t, 0, alice.getScore())
	require.Equal(t, 0, bob.getScore())
}

func newUnconstrainedListToTest() *txListForSender {
	return newListToTest(math.MaxUint32, math.MaxUint32)
}

func newListToTest(maxNumBytes uint32, maxNumTxs uint32) *txListForSender {
	senderConstraints := &senderConstraints{
		maxNumBytes: maxNumBytes,
		maxNumTxs:   maxNumTxs,
	}

	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	scoreComputer := newDefaultScoreComputer(txGasHandler)

	return newTxListForSender(".", senderConstraints, scoreComputer)
}
