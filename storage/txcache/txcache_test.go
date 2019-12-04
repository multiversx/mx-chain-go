package txcache

import (
	"fmt"
	"math"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/assert"
)

func Test_AddTx(t *testing.T) {
	cache := NewTxCache(1000, 16)
	txHash := []byte("hash-1")
	tx := createTx("alice", 1)

	cache.AddTx(txHash, tx)
	foundTx, ok := cache.GetByTxHash(txHash)

	assert.True(t, ok)
	assert.Equal(t, tx, foundTx)
}

func Test_RemoveByTxHash(t *testing.T) {
	cache := NewTxCache(1000, 16)
	txHash := []byte("hash-1")
	tx := createTx("alice", 1)

	cache.AddTx(txHash, tx)
	cache.RemoveByTxHash(txHash)
	foundTx, ok := cache.GetByTxHash(txHash)

	assert.False(t, ok)
	assert.Nil(t, foundTx)
}

func Test_GetSorted(t *testing.T) {
	cache := NewTxCache(1000, 16)

	tx1 := createTx("alice", 1)
	tx2 := createTx("alice", 2)
	tx3 := createTx("alice", 3)

	cache.AddTx([]byte("hash-3"), tx3)
	cache.AddTx([]byte("hash-2"), tx2)
	cache.AddTx([]byte("hash-1"), tx1)

	sorted := cache.GetSorted(math.MaxInt16)
	assert.Len(t, sorted, 3)
}

func Benchmark_Add_Get_Remove_Many(b *testing.B) {
	size := 250000
	noTransactions := 10000

	cache := NewTxCache(size, 16)

	for index := 0; index < noTransactions; index++ {
		hash := fmt.Sprintf("hash%d", index)
		tx := createTx("alice", uint64(index))
		cache.AddTx([]byte(hash), tx)

		foundTx, ok := cache.GetByTxHash([]byte(hash))
		assert.True(b, ok)
		assert.Equal(b, tx, foundTx)
	}

	for index := 0; index < noTransactions; index++ {
		hash := fmt.Sprintf("hash%d", index)
		cache.RemoveByTxHash([]byte(hash))

		foundTx, ok := cache.GetByTxHash([]byte(hash))
		assert.False(b, ok)
		assert.Nil(b, foundTx)
	}
}

func createTx(sender string, nonce uint64) *transaction.Transaction {
	return &transaction.Transaction{
		SndAddr: []byte(sender),
		Nonce:   nonce,
	}
}
