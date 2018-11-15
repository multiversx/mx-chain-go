package transactionPool_test

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/stretchr/testify/assert"
)

func TestTransactionPool_AddTransaction(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool(nil, nil)

	txp.AddTransaction([]byte("hash_tx1"), []byte("tx1"), 1)

	has := txp.MiniPoolTxStore(1).Has([]byte("hash_tx1"))
	assert.True(t, has, "Key was not added to minipool")
	assert.True(t, txp.MiniPoolTxStore(1).Len() == 1,
		"Transaction pool length is not 1 after one element was added")

	txp.AddTransaction([]byte("hash_tx2"), []byte("tx2"), 2)

	assert.False(t, txp.MiniPoolTxStore(1).Has([]byte("hash_tx2")))
	assert.True(t, txp.MiniPoolTxStore(2).Has([]byte("hash_tx2")))
}

func TestTransactionPool_MiniPoolStorageEvictsTx(t *testing.T) {
	t.Parallel()
	size := int(config.TestnetBlockchainConfig.TxPoolStorage.Size)
	txp := transactionPool.NewTransactionPool(nil, nil)
	for i := 1; i < size+2; i++ {
		key := []byte(strconv.Itoa(i))
		val := []byte(strconv.Itoa(i * 2))
		txp.AddTransaction(key, val, 1)
	}

	assert.Equal(t, size, txp.MiniPoolTxStore(1).Len(),
		"Transaction pool entries excedes the maximum configured number")
}

func TestTransactionPool_NoDuplicates(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool(nil, nil)

	txp.AddTransaction([]byte("tx_hash1"), []byte("tx1"), 1)
	txp.AddTransaction([]byte("tx_hash1"), []byte("tx1"), 1)
	assert.Equal(t, 1, txp.MiniPoolTxStore(1).Len(),
		"Transaction pool should not contain duplicates")
}

func TestTransactionPool_ReceivedTransaction(t *testing.T) {
	t.Parallel()

	tx := transaction.Transaction{}

	txp := transactionPool.NewTransactionPool(nil, nil)

	r := txp.ReceivedTransaction(nil, &tx, 1)
	assert.NotNil(t, r)
	assert.Equal(t, "marsher not defined", r.Error())
	assert.Nil(t, txp.MiniPoolTxStore(1))

	txp = transactionPool.NewTransactionPool(marshal.DefMarshalizer, nil)

	r = txp.ReceivedTransaction(nil, &tx, 1)
	assert.NotNil(t, r)
	assert.Equal(t, "hasher not defined", r.Error())
	assert.Nil(t, txp.MiniPoolTxStore(1))

	txp = transactionPool.NewTransactionPool(marshal.DefMarshalizer, hashing.DefHasher)

	r = txp.ReceivedTransaction(nil, &tx, 1)
	assert.Equal(t, 1, txp.MiniPoolTxStore(1).Len(),
		"Mini pool for shard 1 should have 1 elements")
	assert.Nil(t, r)
}

func TestTransactionPool_AddTransactionsInParallel(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool(nil, nil)

	for i := 1; i < 10000+2; i++ {
		key := []byte(strconv.Itoa(i))
		val := []byte(strconv.Itoa(i * 2))
		go func() {
			txp.AddTransaction(key, val, 1)
		}()
	}
}

func TestTransactionPool_RemoveTransaction(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool(nil, nil)

	txp.AddTransaction([]byte("tx_hash1"), []byte("tx1"), 1)
	assert.Equal(t, 1, txp.MiniPoolTxStore(1).Len(),
		"AddTransaction failed, length should be 1")
	txp.RemoveTransaction([]byte("tx_hash1"), 1)
	assert.Equal(t, 0, txp.MiniPoolTxStore(1).Len(),
		"RemoveTransaction failed, length should be 0")

	txp.AddTransaction([]byte("tx_hash1"), []byte("tx1"), 1)
	txp.AddTransaction([]byte("tx_hash2"), []byte("tx2"), 2)
	txp.AddTransaction([]byte("tx_hash1"), []byte("tx1"), 2)
	assert.Equal(t, 1, txp.MiniPoolTxStore(1).Len(),
		"AddTransaction failed, length should be 1")
	assert.Equal(t, 2, txp.MiniPoolTxStore(2).Len(),
		"AddTransaction failed, length should be 2")

	txp.RemoveTransactionFromAllShards([]byte("tx_hash1"))
	assert.Equal(t, 0, txp.MiniPoolTxStore(1).Len(),
		"FindAndRemoveTransaction failed, length should be 0 in shard 1")
	assert.Equal(t, 1, txp.MiniPoolTxStore(2).Len(),
		"FindAndRemoveTransaction failed, length should be 1 in shard 2")
}

func TestTransactionPool_Clear(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool(nil, nil)
	txp.Clear()
	txp.ClearMiniPool(1)

	txp.AddTransaction([]byte("tx_hash1"), []byte("tx1"), 1)
	txp.AddTransaction([]byte("tx_hash2"), []byte("tx2"), 2)
	txp.AddTransaction([]byte("tx_hash1"), []byte("tx1"), 2)

	txp.ClearMiniPool(2)
	assert.Equal(t, 0, txp.MiniPoolTxStore(2).Len(),
		"Mini pool for shard 2 should be empty after clear")
	assert.Equal(t, 1, txp.MiniPoolTxStore(1).Len(),
		"Mini pool for shard 1 should still have one element")

	txp.Clear()
	assert.Nil(t, txp.MiniPool(1), "Shard 1 should not be in the store anymore")
	assert.Nil(t, txp.MiniPool(2), "Shard 2 should not be in the store anymore")
}

func TestTransactionPool_MergeMiniPools(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool(nil, nil)
	txp.AddTransaction([]byte("tx_hash1"), []byte("tx1"), 1)
	txp.AddTransaction([]byte("tx_hash2"), []byte("tx2"), 2)
	txp.AddTransaction([]byte("tx_hash3"), []byte("tx3"), 2)

	txp.MergeMiniPools(1, 2)
	assert.Equal(t, 3, txp.MiniPoolTxStore(2).Len(),
		"Mini pool for shard 1 should have 3 elements")
	assert.Nil(t, txp.MiniPoolTxStore(1))
}

func TestTransactionPool_MoveTransactions(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool(nil, nil)
	txp.AddTransaction([]byte("tx_hash1"), []byte("tx1"), 1)
	txp.AddTransaction([]byte("tx_hash2"), []byte("tx2"), 2)
	txp.AddTransaction([]byte("tx_hash3"), []byte("tx3"), 2)
	txp.AddTransaction([]byte("tx_hash4"), []byte("tx4"), 2)
	txp.AddTransaction([]byte("tx_hash5"), []byte("tx5"), 2)
	txp.AddTransaction([]byte("tx_hash6"), []byte("tx6"), 2)

	txp.MoveTransactions(2, 3, [][]byte{[]byte("tx_hash5"), []byte("tx_hash6")})

	assert.Equal(t, 3, txp.MiniPoolTxStore(2).Len(),
		"Mini pool for shard 2 should have 3 elements")
	assert.Equal(t, 2, txp.MiniPoolTxStore(3).Len(),
		"Mini pool for shard 3 should have 2 elements")
}
