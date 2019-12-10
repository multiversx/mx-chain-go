package txcache

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/stretchr/testify/assert"
)

func Test_AddTx_Sorts(t *testing.T) {
	list := NewTxListForSender(".", 0)

	list.AddTx([]byte("a"), createTx(".", 1))
	list.AddTx([]byte("c"), createTx(".", 3))
	list.AddTx([]byte("d"), createTx(".", 4))
	list.AddTx([]byte("b"), createTx(".", 2))

	txHashes := list.GetTxHashes()

	assert.Equal(t, 4, list.Items.Len())
	assert.Equal(t, 4, len(txHashes))

	assert.Equal(t, []byte("a"), txHashes[0])
	assert.Equal(t, []byte("b"), txHashes[1])
	assert.Equal(t, []byte("c"), txHashes[2])
	assert.Equal(t, []byte("d"), txHashes[3])
}

func Test_RemoveTransaction(t *testing.T) {
	list := NewTxListForSender(".", 0)
	tx := createTx(".", 1)

	list.AddTx([]byte("a"), tx)
	assert.Equal(t, 1, list.Items.Len())

	list.RemoveTx(tx)
	assert.Equal(t, 0, list.Items.Len())
}

func Test_RemoveTransaction_NoPanicWhenTxMissing(t *testing.T) {
	list := NewTxListForSender(".", 0)
	tx := createTx(".", 1)

	list.RemoveTx(tx)
	assert.Equal(t, 0, list.Items.Len())
}

func Test_RemoveHighNonceTransactions(t *testing.T) {
	list := NewTxListForSender(".", 0)

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	list.RemoveHighNonceTxs(50)
	assert.Equal(t, 50, list.Items.Len())
	assert.Equal(t, uint64(49), list.getHighestNonceTx().GetNonce())

	list.RemoveHighNonceTxs(20)
	assert.Equal(t, 30, list.Items.Len())
	assert.Equal(t, uint64(29), list.getHighestNonceTx().GetNonce())

	list.RemoveHighNonceTxs(30)
	assert.Equal(t, 0, list.Items.Len())
	assert.Nil(t, list.getHighestNonceTx())
}

func Test_RemoveHighNonceTransactions_NoPanicWhenCornerCases(t *testing.T) {
	list := NewTxListForSender(".", 0)

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	list.RemoveHighNonceTxs(0)
	assert.Equal(t, 100, list.Items.Len())

	list.RemoveHighNonceTxs(500)
	assert.Equal(t, 0, list.Items.Len())
}

func Test_CopyBatchTo(t *testing.T) {
	list := NewTxListForSender(".", 0)

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	destination := make([]data.TransactionHandler, 1000)

	// Nothing is copied if copy mode isn't started
	copied := list.CopyBatchTo(destination)
	assert.Equal(t, 0, copied)

	// Copy in batches of 50
	list.StartBatchCopying(50)

	// First batch
	copied = list.CopyBatchTo(destination)
	assert.Equal(t, 50, copied)
	assert.NotNil(t, destination[49])
	assert.Nil(t, destination[50])

	// Second batch
	copied = list.CopyBatchTo(destination[50:])
	assert.Equal(t, 50, copied)
	assert.NotNil(t, destination[99])

	// No third batch
	copied = list.CopyBatchTo(destination)
	assert.Equal(t, 0, copied)

	// Restart copy
	list.StartBatchCopying(12345)
	copied = list.CopyBatchTo(destination)
	assert.Equal(t, 100, copied)
}

func Test_CopyBatchTo_NoPanicWhenCornerCases(t *testing.T) {
	list := NewTxListForSender(".", 0)

	for index := 0; index < 100; index++ {
		list.AddTx([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	list.StartBatchCopying(10)

	// When empty destination
	destination := make([]data.TransactionHandler, 0)
	copied := list.CopyBatchTo(destination)
	assert.Equal(t, 0, copied)

	// When small destination
	destination = make([]data.TransactionHandler, 5)
	copied = list.CopyBatchTo(destination)
	assert.Equal(t, 5, copied)
}
