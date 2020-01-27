package txcache

import (
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestMonitoring_numTxAddedAndRemovedDuringEviction(t *testing.T) {
	config := CacheConfig{
		NumChunksHint:              16,
		CountThreshold:             math.MaxUint32,
		NumBytesThreshold:          math.MaxUint32,
		NumSendersToEvictInOneStep: 1,
	}

	cache := NewTxCache(config)

	cache.isEvictionInProgress.Set()

	cache.AddTx([]byte("hash-1"), createTx("alice", 1))
	cache.AddTx([]byte("hash-2"), createTx("alice", 2))
	cache.AddTx([]byte("hash-3"), createTx("alice", 3))
	_ = cache.RemoveTxByHash([]byte("hash-1"))

	cache.isEvictionInProgress.Unset()

	cache.AddTx([]byte("hash-4"), createTx("alice", 4))
	_ = cache.RemoveTxByHash([]byte("hash-2"))

	require.Equal(t, int64(3), cache.numTxAddedDuringEviction.Get())
	require.Equal(t, int64(1), cache.numTxRemovedDuringEviction.Get())
}

func TestMonitoring_numTxAddedAndRemovedBetweenSelections(t *testing.T) {
	config := CacheConfig{
		NumChunksHint:              16,
		CountThreshold:             math.MaxUint32,
		NumBytesThreshold:          math.MaxUint32,
		NumSendersToEvictInOneStep: 1,
	}

	cache := NewTxCache(config)

	require.Equal(t, int64(0), cache.numTxAddedBetweenSelections.Get())

	cache.AddTx([]byte("hash-1"), createTx("alice", 1))
	cache.AddTx([]byte("hash-2"), createTx("alice", 2))
	cache.AddTx([]byte("hash-3"), createTx("alice", 3))
	_ = cache.RemoveTxByHash([]byte("hash-1"))

	require.Equal(t, int64(3), cache.numTxAddedBetweenSelections.Get())

	cache.SelectTransactions(1000, 10)

	cache.AddTx([]byte("hash-4"), createTx("alice", 4))
	cache.AddTx([]byte("hash-5"), createTx("alice", 5))
	_ = cache.RemoveTxByHash([]byte("hash-2"))
	_ = cache.RemoveTxByHash([]byte("hash-3"))

	require.Equal(t, int64(2), cache.numTxAddedBetweenSelections.Get())
	require.Equal(t, int64(2), cache.numTxRemovedBetweenSelections.Get())
}
