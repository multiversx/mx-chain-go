package txcache

import (
	"fmt"
	"math"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/assert"
)

type testContext struct {
	T     *testing.T
	B     *testing.B
	Cache *TxCache
}

func Test_AddTx(t *testing.T) {
	context := setupTestContext(t, nil)

	txHash := []byte("hash-1")
	tx := createTx("alice", 1)

	context.Cache.AddTx(txHash, tx)
	foundTx, ok := context.Cache.GetByTxHash(txHash)

	assert.True(t, ok)
	assert.Equal(t, tx, foundTx)
}

func Test_RemoveByTxHash(t *testing.T) {
	context := setupTestContext(t, nil)

	txHash := []byte("hash-1")
	tx := createTx("alice", 1)

	context.Cache.AddTx(txHash, tx)
	context.Cache.RemoveByTxHash(txHash)
	foundTx, ok := context.Cache.GetByTxHash(txHash)

	assert.False(t, ok)
	assert.Nil(t, foundTx)
}

func Test_GetSorted(t *testing.T) {
	context := setupTestContext(t, nil)

	tx1 := createTx("alice", 1)
	tx2 := createTx("alice", 2)
	tx3 := createTx("alice", 3)

	context.Cache.AddTx([]byte("hash-3"), tx3)
	context.Cache.AddTx([]byte("hash-2"), tx2)
	context.Cache.AddTx([]byte("hash-1"), tx1)

	sorted := context.Cache.GetSorted(math.MaxInt16, 2)
	assert.Len(t, sorted, 3)
}

func Benchmark_Add_Get_Remove_Many(b *testing.B) {
	context := setupTestContext(nil, b)

	noTransactions := 10000

	for index := 0; index < noTransactions; index++ {
		hash := fmt.Sprintf("hash%d", index)
		tx := createTx("alice", uint64(index))
		context.Cache.AddTx([]byte(hash), tx)

		foundTx, ok := context.Cache.GetByTxHash([]byte(hash))
		assert.True(b, ok)
		assert.Equal(b, tx, foundTx)
	}

	for index := 0; index < noTransactions; index++ {
		hash := fmt.Sprintf("hash%d", index)
		context.Cache.RemoveByTxHash([]byte(hash))

		foundTx, ok := context.Cache.GetByTxHash([]byte(hash))
		assert.False(b, ok)
		assert.Nil(b, foundTx)
	}
}

func setupTestContext(t *testing.T, b *testing.B) *testContext {
	cache := NewTxCache(250000, 16)

	context := &testContext{
		T:     t,
		B:     b,
		Cache: cache,
	}

	return context
}

func createTx(sender string, nonce uint64) *transaction.Transaction {
	return &transaction.Transaction{
		SndAddr: []byte(sender),
		Nonce:   nonce,
	}
}
