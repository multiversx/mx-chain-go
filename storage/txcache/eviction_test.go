package txcache

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEviction_EvictHighNonceTransactions(t *testing.T) {
	config := CacheConfig{
		NumChunksHint:            16,
		CountThreshold:           400,
		LargeNumOfTxsForASender:  50,
		NumTxsToEvictFromASender: 25,
		MinGasPriceMicroErd:      100,
	}

	cache := NewTxCache(config)

	for index := 0; index < 200; index++ {
		cache.AddTx(createTx([]byte{'a', byte(index)}, "alice", uint64(index)))
	}

	for index := 0; index < 200; index++ {
		cache.AddTx(createTx([]byte{'b', byte(index)}, "bob", uint64(index)))
	}

	cache.AddTx(createTx([]byte("hash-carol"), "carol", uint64(1)))

	require.Equal(t, int64(3), cache.txListBySender.counter.Get())
	require.Equal(t, int64(401), cache.txByHash.counter.Get())

	cache.makeSnapshotOfSenders()
	nTxs, nSenders := cache.evictHighNonceTransactions()

	require.Equal(t, uint32(50), nTxs)
	require.Equal(t, uint32(0), nSenders)
	require.Equal(t, int64(3), cache.txListBySender.counter.Get())
	require.Equal(t, int64(351), cache.txByHash.counter.Get())
}

func TestEviction_EvictHighNonceTransactions_CoverEmptiedSenderList(t *testing.T) {
	config := CacheConfig{
		NumChunksHint:            1,
		CountThreshold:           0,
		LargeNumOfTxsForASender:  0,
		NumTxsToEvictFromASender: 1,
		MinGasPriceMicroErd:      100,
	}

	cache := NewTxCache(config)
	cache.AddTx(createTx([]byte("hash-alice"), "alice", uint64(1)))
	require.Equal(t, int64(1), cache.CountSenders())

	cache.makeSnapshotOfSenders()

	// Alice is also removed from the map of senders, since it has no transaction left
	nTxs, nSenders := cache.evictHighNonceTransactions()
	require.Equal(t, uint32(1), nTxs)
	require.Equal(t, uint32(1), nSenders)
	require.Equal(t, int64(0), cache.CountSenders())
}

func TestEviction_EvictSendersWhileTooManyTxs(t *testing.T) {
	config := CacheConfig{
		NumChunksHint:              16,
		CountThreshold:             100,
		NumSendersToEvictInOneStep: 20,
		NumBytesThreshold:          math.MaxUint32,
		MinGasPriceMicroErd:        100,
	}

	cache := NewTxCache(config)

	// 200 senders, each with 1 transaction
	for index := 0; index < 200; index++ {
		sender := string(createFakeSenderAddress(index))
		cache.AddTx(createTx([]byte{byte(index)}, sender, uint64(1)))
	}

	require.Equal(t, int64(200), cache.txListBySender.counter.Get())
	require.Equal(t, int64(200), cache.txByHash.counter.Get())

	cache.makeSnapshotOfSenders()
	steps, nTxs, nSenders := cache.evictSendersInLoop()

	require.Equal(t, uint32(5), steps)
	require.Equal(t, uint32(100), nTxs)
	require.Equal(t, uint32(100), nSenders)
	require.Equal(t, int64(100), cache.txListBySender.counter.Get())
	require.Equal(t, int64(100), cache.txByHash.counter.Get())
}

func TestEviction_EvictSendersWhileTooManyBytes(t *testing.T) {
	numBytesPerTx := uint32(1000)

	config := CacheConfig{
		NumChunksHint:              16,
		CountThreshold:             math.MaxUint32,
		NumBytesThreshold:          numBytesPerTx * 100,
		NumSendersToEvictInOneStep: 20,
		MinGasPriceMicroErd:        100,
	}

	cache := NewTxCache(config)

	// 200 senders, each with 1 transaction
	for index := 0; index < 200; index++ {
		sender := string(createFakeSenderAddress(index))
		cache.AddTx(createTxWithParams([]byte{byte(index)}, sender, uint64(1), uint64(numBytesPerTx), 10000, 100*oneTrilion))
	}

	require.Equal(t, int64(200), cache.txListBySender.counter.Get())
	require.Equal(t, int64(200), cache.txByHash.counter.Get())

	cache.makeSnapshotOfSenders()
	steps, nTxs, nSenders := cache.evictSendersInLoop()

	require.Equal(t, uint32(5), steps)
	require.Equal(t, uint32(100), nTxs)
	require.Equal(t, uint32(100), nSenders)
	require.Equal(t, int64(100), cache.txListBySender.counter.Get())
	require.Equal(t, int64(100), cache.txByHash.counter.Get())
}

func TestEviction_DoEvictionDoneInPassTwo_BecauseOfCount(t *testing.T) {
	config := CacheConfig{
		NumChunksHint:              16,
		NumBytesThreshold:          math.MaxUint32,
		CountThreshold:             2,
		NumSendersToEvictInOneStep: 2,
		MinGasPriceMicroErd:        100,
	}

	cache := NewTxCache(config)
	cache.AddTx(createTxWithParams([]byte("hash-alice"), "alice", uint64(1), 1000, 100000, 100*oneTrilion))
	cache.AddTx(createTxWithParams([]byte("hash-bob"), "bob", uint64(1), 1000, 100000, 100*oneTrilion))
	cache.AddTx(createTxWithParams([]byte("hash-carol"), "carol", uint64(1), 1000, 100000, 700*oneTrilion))

	cache.doEviction()
	require.Equal(t, uint32(0), cache.evictionJournal.passOneNumTxs)
	require.Equal(t, uint32(0), cache.evictionJournal.passOneNumSenders)
	require.Equal(t, uint32(2), cache.evictionJournal.passTwoNumTxs)
	require.Equal(t, uint32(2), cache.evictionJournal.passTwoNumSenders)
	require.Equal(t, uint32(1), cache.evictionJournal.passTwoNumSteps)

	// Alice and Bob evicted. Carol still there.
	_, ok := cache.GetByTxHash([]byte("hash-carol"))
	require.True(t, ok)
	require.Equal(t, int64(1), cache.CountSenders())
	require.Equal(t, int64(1), cache.CountTx())
}

func TestEviction_DoEvictionDoneInPassTwo_BecauseOfSize(t *testing.T) {
	config := CacheConfig{
		NumChunksHint:              16,
		CountThreshold:             math.MaxUint32,
		NumBytesThreshold:          1000,
		NumSendersToEvictInOneStep: 2,
		MinGasPriceMicroErd:        100,
	}

	cache := NewTxCache(config)
	cache.AddTx(createTxWithParams([]byte("hash-alice"), "alice", uint64(1), 800, 100000, 100*oneTrilion))
	cache.AddTx(createTxWithParams([]byte("hash-bob"), "bob", uint64(1), 500, 100000, 100*oneTrilion))
	cache.AddTx(createTxWithParams([]byte("hash-carol"), "carol", uint64(1), 200, 100000, 700*oneTrilion))

	require.InDelta(t, float64(19.50394606), cache.getRawScoreOfSender("alice"), delta)
	require.InDelta(t, float64(23.68494667), cache.getRawScoreOfSender("bob"), delta)
	require.InDelta(t, float64(100), cache.getRawScoreOfSender("carol"), delta)

	cache.doEviction()
	require.Equal(t, uint32(0), cache.evictionJournal.passOneNumTxs)
	require.Equal(t, uint32(0), cache.evictionJournal.passOneNumSenders)
	require.Equal(t, uint32(2), cache.evictionJournal.passTwoNumTxs)
	require.Equal(t, uint32(2), cache.evictionJournal.passTwoNumSenders)
	require.Equal(t, uint32(1), cache.evictionJournal.passTwoNumSteps)

	// Alice and Bob evicted (lower score). Carol still there.
	_, ok := cache.GetByTxHash([]byte("hash-carol"))
	require.True(t, ok)
	require.Equal(t, int64(1), cache.CountSenders())
	require.Equal(t, int64(1), cache.CountTx())
}

func TestEviction_doEvictionDoesNothingWhenAlreadyInProgress(t *testing.T) {
	config := CacheConfig{
		NumChunksHint:              1,
		CountThreshold:             0,
		NumSendersToEvictInOneStep: 1,
	}

	cache := NewTxCache(config)
	cache.AddTx(createTx([]byte("hash-alice"), "alice", uint64(1)))

	cache.isEvictionInProgress.Set()
	cache.doEviction()

	require.False(t, cache.evictionJournal.evictionPerformed)
}

func TestEviction_evictSendersInLoop_CoverLoopBreak_WhenSmallBatch(t *testing.T) {
	config := CacheConfig{
		NumChunksHint:              1,
		CountThreshold:             0,
		NumSendersToEvictInOneStep: 42,
	}

	cache := NewTxCache(config)
	cache.AddTx(createTx([]byte("hash-alice"), "alice", uint64(1)))

	cache.makeSnapshotOfSenders()

	steps, nTxs, nSenders := cache.evictSendersInLoop()
	require.Equal(t, uint32(0), steps)
	require.Equal(t, uint32(1), nTxs)
	require.Equal(t, uint32(1), nSenders)
}

func TestEviction_evictSendersWhile_ShouldContinueBreak(t *testing.T) {
	config := CacheConfig{
		NumChunksHint:              1,
		CountThreshold:             0,
		NumSendersToEvictInOneStep: 1,
	}

	cache := NewTxCache(config)
	cache.AddTx(createTx([]byte("hash-alice"), "alice", uint64(1)))
	cache.AddTx(createTx([]byte("hash-bob"), "bob", uint64(1)))

	cache.makeSnapshotOfSenders()

	steps, nTxs, nSenders := cache.evictSendersWhile(func() bool {
		return false
	})

	require.Equal(t, uint32(0), steps)
	require.Equal(t, uint32(0), nTxs)
	require.Equal(t, uint32(0), nSenders)
}

// This seems to be the worst case in terms of eviction complexity
// Eviction is triggered often and little eviction (only 10 senders) is done
func Test_AddWithEviction_UniformDistribution_250000x1_WithConfig_NumSendersToEvictInOneStep_10(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	config := CacheConfig{
		NumChunksHint:              16,
		EvictionEnabled:            true,
		NumBytesThreshold:          1000000000,
		CountThreshold:             240000,
		NumSendersToEvictInOneStep: 10,
		LargeNumOfTxsForASender:    1000,
		NumTxsToEvictFromASender:   250,
	}

	cache := NewTxCache(config)
	addManyTransactionsWithUniformDistribution(cache, 250000, 1)
	require.Equal(t, int64(240000), cache.CountTx())
}

func Test_AddWithEviction_UniformDistribution_250000x1_WithConfig_NumSendersToEvictInOneStep_100(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	config := CacheConfig{
		NumChunksHint:              16,
		EvictionEnabled:            true,
		NumBytesThreshold:          1000000000,
		CountThreshold:             240000,
		NumSendersToEvictInOneStep: 100,
		LargeNumOfTxsForASender:    1000,
		NumTxsToEvictFromASender:   250,
	}

	cache := NewTxCache(config)
	addManyTransactionsWithUniformDistribution(cache, 250000, 1)
	require.Equal(t, int64(240000), cache.CountTx())
}

func Test_AddWithEviction_UniformDistribution_250000x1_WithConfig_NumSendersToEvictInOneStep_1000(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	config := CacheConfig{
		NumChunksHint:              16,
		EvictionEnabled:            true,
		NumBytesThreshold:          1000000000,
		CountThreshold:             240000,
		NumSendersToEvictInOneStep: 1000,
		LargeNumOfTxsForASender:    1000,
		NumTxsToEvictFromASender:   250,
	}

	cache := NewTxCache(config)
	addManyTransactionsWithUniformDistribution(cache, 250000, 1)
	require.Equal(t, int64(240000), cache.CountTx())
}
