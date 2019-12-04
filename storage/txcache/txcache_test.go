package txcache

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/assert"
)

func Test_AddTx(t *testing.T) {
	cache := NewTxCache(1000, 16)
	txHash := []byte("hash-1")
	tx := &transaction.Transaction{}

	cache.AddTx(txHash, tx)
	foundTx, ok := cache.GetByTxHash(txHash)

	assert.True(t, ok)
	assert.Equal(t, tx, foundTx)
}

func Test_RemoveByTxHash(t *testing.T) {
	cache := NewTxCache(1000, 16)
	txHash := []byte("hash-1")
	tx := &transaction.Transaction{}

	cache.AddTx(txHash, tx)
	cache.RemoveByTxHash(txHash)
	foundTx, ok := cache.GetByTxHash(txHash)

	assert.False(t, ok)
	assert.Nil(t, foundTx)
}
