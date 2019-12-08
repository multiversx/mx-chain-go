package txcache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AddTransaction_Sorts(t *testing.T) {
	list := NewTxListForSender()

	list.AddTransaction([]byte("a"), createTx(".", 1))
	list.AddTransaction([]byte("c"), createTx(".", 3))
	list.AddTransaction([]byte("d"), createTx(".", 4))
	list.AddTransaction([]byte("b"), createTx(".", 2))

	txHashes := list.getTxHashes()

	assert.Equal(t, 4, list.Items.Len())
	assert.Equal(t, 4, len(txHashes))

	assert.Equal(t, []byte("a"), txHashes[0])
	assert.Equal(t, []byte("b"), txHashes[1])
	assert.Equal(t, []byte("c"), txHashes[2])
	assert.Equal(t, []byte("d"), txHashes[3])
}

func Test_RemoveTransaction(t *testing.T) {
	list := NewTxListForSender()
	tx := createTx(".", 1)

	list.AddTransaction([]byte("a"), tx)
	assert.Equal(t, 1, list.Items.Len())

	list.RemoveTransaction(tx)
	assert.Equal(t, 0, list.Items.Len())
}

func Test_RemoveTransaction_NoPanicWhenTxMissing(t *testing.T) {
	list := NewTxListForSender()
	tx := createTx(".", 1)

	list.RemoveTransaction(tx)
	assert.Equal(t, 0, list.Items.Len())
}

func Test_RemoveHighNonceTransactions(t *testing.T) {
	list := NewTxListForSender()

	for index := 0; index < 100; index++ {
		list.AddTransaction([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	list.RemoveHighNonceTransactions(50)
	assert.Equal(t, 50, list.Items.Len())
	assert.Equal(t, uint64(49), list.getHighestNonceTx().Nonce)

	list.RemoveHighNonceTransactions(20)
	assert.Equal(t, 30, list.Items.Len())
	assert.Equal(t, uint64(29), list.getHighestNonceTx().Nonce)

	list.RemoveHighNonceTransactions(30)
	assert.Equal(t, 0, list.Items.Len())
	assert.Nil(t, list.getHighestNonceTx())
}

func Test_RemoveHighNonceTransactions_NoPanicWhenCornerCases(t *testing.T) {
	list := NewTxListForSender()

	for index := 0; index < 100; index++ {
		list.AddTransaction([]byte{byte(index)}, createTx(".", uint64(index)))
	}

	list.RemoveHighNonceTransactions(0)
	assert.Equal(t, 100, list.Items.Len())

	list.RemoveHighNonceTransactions(500)
	assert.Equal(t, 0, list.Items.Len())
}
