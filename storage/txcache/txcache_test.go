package txcache

import (
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

func createTx(sender string, nonce uint64) *transaction.Transaction {
	return &transaction.Transaction{
		SndAddr: []byte(sender),
		Nonce:   nonce,
	}
}
