package txcache

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/stretchr/testify/require"
)

func Test_AddTx_Sorts(t *testing.T) {
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

func Test_findTx(t *testing.T) {
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

func Test_findTx_CoverNonceComparisonOptimization(t *testing.T) {
	list := newListToTest()
	list.AddTx([]byte("A"), createTx(".", 42))

	// Find one with a lower nonce, not added to cache
	noElement := list.findListElementWithTx(createTx(".", 41))
	require.Nil(t, noElement)
}

func Test_RemoveTransaction(t *testing.T) {
	list := newListToTest()
	tx := createTx(".", 1)

	list.AddTx([]byte("a"), tx)
	require.Equal(t, 1, list.items.Len())

	list.RemoveTx(tx)
	require.Equal(t, 0, list.items.Len())
}

func Test_RemoveTransaction_NoPanicWhenTxMissing(t *testing.T) {
	list := newListToTest()
	tx := createTx(".", 1)

	list.RemoveTx(tx)
	require.Equal(t, 0, list.items.Len())
}

func Test_RemoveHighNonceTransactions(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	list.RemoveHighNonceTxs(50)
	require.Equal(t, 50, list.items.Len())
	require.Equal(t, uint64(49), list.getHighestNonceTx().GetNonce())

	list.RemoveHighNonceTxs(20)
	require.Equal(t, 30, list.items.Len())
	require.Equal(t, uint64(29), list.getHighestNonceTx().GetNonce())

	list.RemoveHighNonceTxs(30)
	require.Equal(t, 0, list.items.Len())
	require.Nil(t, list.getHighestNonceTx())
}

func Test_RemoveHighNonceTransactions_NoPanicWhenCornerCases(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	list.RemoveHighNonceTxs(0)
	require.Equal(t, 100, list.items.Len())

	list.RemoveHighNonceTxs(500)
	require.Equal(t, 0, list.items.Len())
}

func Test_CopyBatchTo(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	destination := make([]data.TransactionHandler, 1000)
	destinationHashes := make([][]byte, 1000)

	// First batch
	copied := list.copyBatchTo(true, destination, destinationHashes, 50)
	require.Equal(t, 50, copied)
	require.NotNil(t, destination[49])
	require.Nil(t, destination[50])

	// Second batch
	copied = list.copyBatchTo(false, destination[50:], destinationHashes[50:], 50)
	require.Equal(t, 50, copied)
	require.NotNil(t, destination[99])

	// No third batch
	copied = list.copyBatchTo(false, destination, destinationHashes, 50)
	require.Equal(t, 0, copied)

	// Restart copy
	copied = list.copyBatchTo(true, destination, destinationHashes, 12345)
	require.Equal(t, 100, copied)
}

func Test_CopyBatchTo_NoPanicWhenCornerCases(t *testing.T) {
	list := newListToTest()

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	// When empty destination
	destination := make([]data.TransactionHandler, 0)
	destinationHashes := make([][]byte, 0)
	copied := list.copyBatchTo(true, destination, destinationHashes, 10)
	require.Equal(t, 0, copied)

	// When small destination
	destination = make([]data.TransactionHandler, 5)
	destinationHashes = make([][]byte, 5)
	copied = list.copyBatchTo(false, destination, destinationHashes, 10)
	require.Equal(t, 5, copied)
}

func Test_getTxHashes(t *testing.T) {
	list := newListToTest()
	require.Len(t, list.getTxHashes(), 0)

	list.AddTx([]byte("A"), createTx(".", 1))
	require.Len(t, list.getTxHashes(), 1)

	list.AddTx([]byte("B"), createTx(".", 2))
	list.AddTx([]byte("C"), createTx(".", 3))
	require.Len(t, list.getTxHashes(), 3)
}

func newListToTest() *txListForSender {
	return newTxListForSender(".", &CacheConfig{MinGasPriceMicroErd: 100})
}
