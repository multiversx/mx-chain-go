package transactionPool_test

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/stretchr/testify/assert"
)

func TestTransactionPool_AddTransaction(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool()

	txp.AddTransaction([]byte("abc123"), 1)

	has := (*txp.MiniPoolTxStore(1)).Has([]byte("abc123"))
	assert.True(t, has, "Key was not added to minipool")
	assert.True(t, (*txp.MiniPoolTxStore(1)).Len() == 1,
		"Transaction pool length is not 1 after one element was added")

	txp.AddTransaction([]byte("tx_second_shard"), 2)

	assert.False(t, (*txp.MiniPoolTxStore(1)).Has([]byte("tx_second_shard")))
	assert.True(t, (*txp.MiniPoolTxStore(2)).Has([]byte("tx_second_shard")))
}

func TestTransactionPool_MiniPoolStorageEvictsTx(t *testing.T) {
	t.Parallel()
	size := int(config.TestnetBlockchainConfig.TxPoolStorage.Size)
	txp := transactionPool.NewTransactionPool()
	for i:= 1; i < size + 2; i++ {
		key := []byte(strconv.Itoa(i))
		txp.AddTransaction(key, 1)
	}

	assert.Equal(t, size, (*txp.MiniPoolTxStore(1)).Len(),
		"Transaction pool entries excedes the maximum configured number")
}

func TestTransactionPool_NoDuplicates(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool()

	txp.AddTransaction([]byte("abc123"), 1)
	txp.AddTransaction([]byte("abc123"), 1)
	assert.Equal(t, 1, (*txp.MiniPoolTxStore(1)).Len(),
		"Transaction pool should not contain duplicates")
}

func TestTransactionPool_AddTransactionsInParallel(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool()

	for i:= 1; i < 10000 + 2; i++ {
		go func() {
			key := []byte(strconv.Itoa(i))
			txp.AddTransaction(key, 1)
		}()
	}
}

func TestTransactionPool_RemoveTransaction(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool()

	txp.AddTransaction([]byte("tx_hash1"), 1)
	assert.Equal(t, 1, (*txp.MiniPoolTxStore(1)).Len(),
		"AddTransaction failed, length should be 1")
	txp.RemoveTransaction([]byte("tx_hash1"), 1)
	assert.Equal(t, 0, (*txp.MiniPoolTxStore(1)).Len(),
		"RemoveTransaction failed, length should be 0")

	txp.AddTransaction([]byte("tx_hash1"), 1)
	txp.AddTransaction([]byte("tx_hash2"), 2)
	txp.AddTransaction([]byte("tx_hash1"), 2)
	assert.Equal(t, 1, (*txp.MiniPoolTxStore(1)).Len(),
		"AddTransaction failed, length should be 1")
	assert.Equal(t, 2, (*txp.MiniPoolTxStore(2)).Len(),
		"AddTransaction failed, length should be 2")

	txp.FindAndRemoveTransaction([]byte("tx_hash1"))
	assert.Equal(t, 0, (*txp.MiniPoolTxStore(1)).Len(),
		"FindAndRemoveTransaction failed, length should be 0 in shard 1")
	assert.Equal(t, 1, (*txp.MiniPoolTxStore(2)).Len(),
		"FindAndRemoveTransaction failed, length should be 1 in shard 2")
}

func TestTransactionPool_Clear(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool()
	txp.Clear()
	txp.ClearMiniPool(1)

	txp.AddTransaction([]byte("tx_hash1"), 1)
	txp.AddTransaction([]byte("tx_hash2"), 2)
	txp.AddTransaction([]byte("tx_hash1"), 2)

	txp.ClearMiniPool(2)
	assert.Equal(t, 0, (*txp.MiniPoolTxStore(2)).Len(),
		"Mini pool for shard 2 should be empty after clear")
	assert.Equal(t, 1, (*txp.MiniPoolTxStore(1)).Len(),
		"Mini pool for shard 1 should still have one element")

	txp.Clear()
	assert.Equal(t, 0, len(txp.MiniPoolsStore), "MiniPoolStore map should be empty after Clear")
}

func TestTransactionPool_MergeMiniPools(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool()
	txp.AddTransaction([]byte("tx_hash1"), 1)
	txp.AddTransaction([]byte("tx_hash2"), 2)
	txp.AddTransaction([]byte("tx_hash3"), 2)

	txp.MergeMiniPools(1, 2)
	assert.Equal(t, 3, (*txp.MiniPoolTxStore(1)).Len(),
		"Mini pool for shard 1 should have 3 elements")
	assert.Nil(t, txp.MiniPoolTxStore(2))
}

func TestTransactionPool_SplitMiniPool(t *testing.T) {
	t.Parallel()
	txp := transactionPool.NewTransactionPool()
	txp.AddTransaction([]byte("tx_hash1"), 1)
	txp.AddTransaction([]byte("tx_hash2"), 2)
	txp.AddTransaction([]byte("tx_hash3"), 2)
	txp.AddTransaction([]byte("tx_hash4"), 2)
	txp.AddTransaction([]byte("tx_hash5"), 2)
	txp.AddTransaction([]byte("tx_hash6"), 2)

	txp.SplitMiniPool(2, 3, [][]byte{[]byte("tx_hash5"), []byte("tx_hash6")})

	assert.Equal(t, 3, (*txp.MiniPoolTxStore(2)).Len(),
		"Mini pool for shard 2 should have 3 elements")
	assert.Equal(t, 2, (*txp.MiniPoolTxStore(3)).Len(),
		"Mini pool for shard 3 should have 2 elements")
}