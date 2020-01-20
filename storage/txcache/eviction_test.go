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
		cache.AddTx([]byte{byte(index)}, createTxWithParams(sender, uint64(1), uint64(numBytesPerTx), 10000, 100*oneTrilion))
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
		NumBytesThreshold:          math.MaxUint32,
		CountThreshold:             2,
		NumSendersToEvictInOneStep: 2,
	}

	cache := NewTxCacheWithEviction(16, config)
	cache.AddTx([]byte("hash-alice"), createTxWithParams("alice", uint64(1), 1000, 100000, 100*oneTrilion))
	cache.AddTx([]byte("hash-bob"), createTxWithParams("bob", uint64(1), 1000, 100000, 100*oneTrilion))
	cache.AddTx([]byte("hash-carol"), createTxWithParams("carol", uint64(1), 1000, 100000, 700*oneTrilion))

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
		NumBytesThreshold:          1000,
		NumSendersToEvictInOneStep: 2,
	}

	cache := NewTxCacheWithEviction(16, config)
	cache.AddTx([]byte("hash-alice"), createTxWithParams("alice", uint64(1), 800, 100000, 100*oneTrilion))
	cache.AddTx([]byte("hash-bob"), createTxWithParams("bob", uint64(1), 500, 100000, 100*oneTrilion))
	cache.AddTx([]byte("hash-carol"), createTxWithParams("carol", uint64(1), 200, 100000, 700*oneTrilion))

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
