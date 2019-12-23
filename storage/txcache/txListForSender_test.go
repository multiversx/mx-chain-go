package txcache

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/stretchr/testify/assert"
)

func Test_AddTx_Sorts(t *testing.T) {
	list := newTxListForSender(".", 0)

	list.AddTx([]byte("a"), createTx(".", 1))
	list.AddTx([]byte("c"), createTx(".", 3))
	list.AddTx([]byte("d"), createTx(".", 4))
	list.AddTx([]byte("b"), createTx(".", 2))

	txHashes := list.getTxHashes()

	assert.Equal(t, 4, list.items.Len())
	assert.Equal(t, 4, len(txHashes))

	assert.Equal(t, []byte("a"), txHashes[0])
	assert.Equal(t, []byte("b"), txHashes[1])
	assert.Equal(t, []byte("c"), txHashes[2])
	assert.Equal(t, []byte("d"), txHashes[3])
}

func Test_findTx(t *testing.T) {
	list := newTxListForSender(".", 0)

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

	assert.Equal(t, txA, elementWithA.Value.(txListForSenderNode).tx)
	assert.Equal(t, txANewer, elementWithANewer.Value.(txListForSenderNode).tx)
	assert.Equal(t, txB, elementWithB.Value.(txListForSenderNode).tx)
	assert.Nil(t, noElementWithD)
}

func Test_RemoveTransaction(t *testing.T) {
	list := newTxListForSender(".", 0)
	tx := createTx(".", 1)

	list.AddTx([]byte("a"), tx)
	assert.Equal(t, 1, list.items.Len())

	list.RemoveTx(tx)
	assert.Equal(t, 0, list.items.Len())
}

func Test_RemoveTransaction_NoPanicWhenTxMissing(t *testing.T) {
	list := newTxListForSender(".", 0)
	tx := createTx(".", 1)

	list.RemoveTx(tx)
	assert.Equal(t, 0, list.items.Len())
}

func Test_RemoveHighNonceTransactions(t *testing.T) {
	list := newTxListForSender(".", 0)

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	list.RemoveHighNonceTxs(50)
	assert.Equal(t, 50, list.items.Len())
	assert.Equal(t, uint64(49), list.getHighestNonceTx().GetNonce())

	list.RemoveHighNonceTxs(20)
	assert.Equal(t, 30, list.items.Len())
	assert.Equal(t, uint64(29), list.getHighestNonceTx().GetNonce())

	list.RemoveHighNonceTxs(30)
	assert.Equal(t, 0, list.items.Len())
	assert.Nil(t, list.getHighestNonceTx())
}

func Test_RemoveHighNonceTransactions_NoPanicWhenCornerCases(t *testing.T) {
	list := newTxListForSender(".", 0)

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	list.RemoveHighNonceTxs(0)
	assert.Equal(t, 100, list.items.Len())

	list.RemoveHighNonceTxs(500)
	assert.Equal(t, 0, list.items.Len())
}

func Test_CopyBatchTo(t *testing.T) {
	list := newTxListForSender(".", 0)

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	destination := make([]data.TransactionHandler, 1000)

	// First batch
	copied := list.copyBatchTo(true, destination, 50)
	assert.Equal(t, 50, copied)
	assert.NotNil(t, destination[49])
	assert.Nil(t, destination[50])

	// Second batch
	copied = list.copyBatchTo(false, destination[50:], 50)
	assert.Equal(t, 50, copied)
	assert.NotNil(t, destination[99])

	// No third batch
	copied = list.copyBatchTo(false, destination, 50)
	assert.Equal(t, 0, copied)

	// Restart copy
	copied = list.copyBatchTo(true, destination, 12345)
	assert.Equal(t, 100, copied)
}

func Test_CopyBatchTo_NoPanicWhenCornerCases(t *testing.T) {
	list := newTxListForSender(".", 0)

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	// When empty destination
	destination := make([]data.TransactionHandler, 0)
	copied := list.copyBatchTo(true, destination, 10)
	assert.Equal(t, 0, copied)

	// When small destination
	destination = make([]data.TransactionHandler, 5)
	copied = list.copyBatchTo(false, destination, 10)
	assert.Equal(t, 5, copied)
}
