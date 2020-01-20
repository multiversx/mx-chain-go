package txcache

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_EvictHighNonceTransactions(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:                  400,
		ALotOfTransactionsForASender:    50,
		NumTxsToEvictForASenderWithALot: 25,
	}

	cache := NewTxCacheWithEviction(16, config)

	for index := 0; index < 200; index++ {
		cache.AddTx([]byte{'a', byte(index)}, createTx("alice", uint64(index)))
	}

	for index := 0; index < 200; index++ {
		cache.AddTx([]byte{'b', byte(index)}, createTx("bob", uint64(index)))
	}

	cache.AddTx([]byte("hash-carol"), createTx("carol", uint64(1)))

	require.Equal(t, int64(3), cache.txListBySender.counter.Get())
	require.Equal(t, int64(401), cache.txByHash.counter.Get())

	nTxs, nSenders := cache.evictHighNonceTransactions()

	require.Equal(t, uint32(50), nTxs)
	require.Equal(t, uint32(0), nSenders)
	require.Equal(t, int64(3), cache.txListBySender.counter.Get())
	require.Equal(t, int64(351), cache.txByHash.counter.Get())
}

func Test_EvictHighNonceTransactions_CoverEmptiedSenderList(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:                  0,
		ALotOfTransactionsForASender:    0,
		NumTxsToEvictForASenderWithALot: 1,
	}

	cache := NewTxCacheWithEviction(1, config)
	cache.AddTx([]byte("hash-alice"), createTx("alice", uint64(1)))
	require.Equal(t, int64(1), cache.CountSenders())

	// Alice is also removed from the map of senders, since it has no transaction left
	nTxs, nSenders := cache.evictHighNonceTransactions()
	require.Equal(t, uint32(1), nTxs)
	require.Equal(t, uint32(1), nSenders)
	require.Equal(t, int64(0), cache.CountSenders())
}

func Test_EvictSendersWhileTooManyTxs(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:             100,
		NumSendersToEvictInOneStep: 20,
		NumBytesThreshold:          math.MaxUint32,
	}

	cache := NewTxCacheWithEviction(16, config)

	// 200 senders, each with 1 transaction
	for index := 0; index < 200; index++ {
		sender := string(createFakeSenderAddress(index))
		cache.AddTx([]byte{byte(index)}, createTx(sender, uint64(1)))
	}

	require.Equal(t, int64(200), cache.txListBySender.counter.Get())
	require.Equal(t, int64(200), cache.txByHash.counter.Get())

	steps, nTxs, nSenders := cache.evictSendersInLoop()

	require.Equal(t, uint32(5), steps)
	require.Equal(t, uint32(100), nTxs)
	require.Equal(t, uint32(100), nSenders)
	require.Equal(t, int64(100), cache.txListBySender.counter.Get())
	require.Equal(t, int64(100), cache.txByHash.counter.Get())
}

// TODO-TXCACHE: perhaps remove test?
func Test_EvictSendersWhileTooManyTxs_CoverLoopBreak_WhenSmallBatch(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:             0,
		NumSendersToEvictInOneStep: 42,
	}

	cache := NewTxCacheWithEviction(1, config)
	cache.AddTx([]byte("hash-alice"), createTx("alice", uint64(1)))

	// Eviction done in 1 step, since "NumSendersToEvictInOneStep" > number of senders
	steps, nTxs, nSenders := cache.evictSendersInLoop()
	require.Equal(t, uint32(1), steps)
	require.Equal(t, uint32(1), nTxs)
	require.Equal(t, uint32(1), nSenders)
}

func Test_EvictSendersWhileTooManyBytes(t *testing.T) {
	numBytesPerTx := uint32(1000)

	config := EvictionConfig{
		CountThreshold:             math.MaxUint32,
		NumBytesThreshold:          numBytesPerTx * 100,
		NumSendersToEvictInOneStep: 20,
	}

	cache := NewTxCacheWithEviction(16, config)

	// 200 senders, each with 1 transaction
	for index := 0; index < 200; index++ {
		sender := string(createFakeSenderAddress(index))
		cache.AddTx([]byte{byte(index)}, createTxWithData(sender, uint64(1), uint64(numBytesPerTx)))
	}

	require.Equal(t, int64(200), cache.txListBySender.counter.Get())
	require.Equal(t, int64(200), cache.txByHash.counter.Get())

	steps, nTxs, nSenders := cache.evictSendersInLoop()

	require.Equal(t, uint32(5), steps)
	require.Equal(t, uint32(100), nTxs)
	require.Equal(t, uint32(100), nSenders)
	require.Equal(t, int64(100), cache.txListBySender.counter.Get())
	require.Equal(t, int64(100), cache.txByHash.counter.Get())
}

func TestEviction_DoEvictionDoneInPassTwo_BecauseOfCount(t *testing.T) {
	config := EvictionConfig{
		NumBytesThreshold:          40960,
		CountThreshold:             2,
		NumSendersToEvictInOneStep: 2,
	}

	cache := NewTxCacheWithEviction(16, config)
	cache.AddTx([]byte("hash-alice"), createTx("alice", uint64(1)))
	cache.AddTx([]byte("hash-bob"), createTx("bob", uint64(1)))
	cache.AddTx([]byte("hash-carol"), createTx("carol", uint64(1)))

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
	config := EvictionConfig{
		CountThreshold:             math.MaxUint32,
		NumBytesThreshold:          1500,
		NumSendersToEvictInOneStep: 2,
	}

	cache := NewTxCacheWithEviction(16, config)
	cache.AddTx([]byte("hash-alice"), createTxWithGas("alice", uint64(1), 800, 8))
	cache.AddTx([]byte("hash-bob"), createTxWithGas("bob", uint64(1), 400, 4))
	cache.AddTx([]byte("hash-carol"), createTxWithGas("carol", uint64(1), 500, 5))

	cache.doEviction()
	require.Equal(t, uint32(0), cache.evictionJournal.passOneNumTxs)
	require.Equal(t, uint32(0), cache.evictionJournal.passOneNumSenders)
	require.Equal(t, uint32(0), cache.evictionJournal.passTwoNumTxs)
	require.Equal(t, uint32(0), cache.evictionJournal.passTwoNumSenders)

	// Alice and Bob evicted (lower score). Carol still there.
	_, ok := cache.GetByTxHash([]byte("hash-carol"))
	require.True(t, ok)
	require.Equal(t, int64(1), cache.CountSenders())
	require.Equal(t, int64(1), cache.CountTx())
}
