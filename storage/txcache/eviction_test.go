package txcache

import "testing"

import "github.com/stretchr/testify/require"

import "math"

func Test_EvictOldestSenders(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:             1,
		NumSendersToEvictInOneStep: 2,
	}

	cache := NewTxCacheWithEviction(16, config)

	cache.AddTx([]byte("hash-alice"), createTx("alice", uint64(1)))
	cache.AddTx([]byte("hash-bob"), createTx("bob", uint64(1)))
	cache.AddTx([]byte("hash-carol"), createTx("carol", uint64(1)))

	nTxs, nSenders := cache.evictOldestSenders()

	require.Equal(t, uint32(2), nTxs)
	require.Equal(t, uint32(2), nSenders)
	require.Equal(t, int64(1), cache.txListBySender.counter.Get())
	require.Equal(t, int64(1), cache.txByHash.counter.Get())
}

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
	}

	cache := NewTxCacheWithEviction(16, config)

	// 200 senders, each with 1 transaction
	for index := 0; index < 200; index++ {
		sender := string(createFakeSenderAddress(index))
		cache.AddTx([]byte{byte(index)}, createTx(sender, uint64(1)))
	}

	require.Equal(t, int64(200), cache.txListBySender.counter.Get())
	require.Equal(t, int64(200), cache.txByHash.counter.Get())

	steps, nTxs, nSenders := cache.evictSendersWhileTooManyTxs()

	require.Equal(t, uint32(6), steps)
	require.Equal(t, uint32(100), nTxs)
	require.Equal(t, uint32(100), nSenders)
	require.Equal(t, int64(100), cache.txListBySender.counter.Get())
	require.Equal(t, int64(100), cache.txByHash.counter.Get())
}

func Test_EvictSendersWhileTooManyTxs_CoverLoopBreak_WhenSmallBatch(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:             0,
		NumSendersToEvictInOneStep: 42,
	}

	cache := NewTxCacheWithEviction(1, config)
	cache.AddTx([]byte("hash-alice"), createTx("alice", uint64(1)))

	// Eviction done in 1 step, since "NumSendersToEvictInOneStep" > number of senders
	steps, nTxs, nSenders := cache.evictSendersWhileTooManyTxs()
	require.Equal(t, uint32(1), steps)
	require.Equal(t, uint32(1), nTxs)
	require.Equal(t, uint32(1), nSenders)
}

func Test_EvictSendersWhileTooManyBytes(t *testing.T) {
	numBytesPerTx := uint32(1000)

	config := EvictionConfig{
		// 128 bytes for extra fields, capacity of 100 transactions
		NumBytesThreshold:          (numBytesPerTx + 128) * 100,
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

	steps, nTxs, nSenders := cache.evictSendersWhileTooManyBytes()

	require.Equal(t, uint32(6), steps)
	require.Equal(t, uint32(100), nTxs)
	require.Equal(t, uint32(100), nSenders)
	require.Equal(t, int64(100), cache.txListBySender.counter.Get())
	require.Equal(t, int64(100), cache.txByHash.counter.Get())
}

func Test_DoEviction_DoneInPass0(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:             math.MaxUint32,
		NumBytesThreshold:          1500,
		NumSendersToEvictInOneStep: 2,
	}

	cache := NewTxCacheWithEviction(16, config)
	cache.AddTx([]byte("hash-alice"), createTxWithGas("alice", uint64(1), 800, 1))
	cache.AddTx([]byte("hash-bob"), createTxWithGas("bob", uint64(1), 400, 1))
	cache.AddTx([]byte("hash-carol"), createTxWithGas("carol", uint64(1), 500, 1))

	journal := cache.doEviction()
	require.Equal(t, uint32(2), journal.passZeroNumSteps)
	require.Equal(t, uint32(2), journal.passZeroNumTxs)
	require.Equal(t, uint32(2), journal.passZeroNumSenders)
	require.Equal(t, uint32(0), journal.passOneNumTxs)
	require.Equal(t, uint32(0), journal.passOneNumSenders)
	require.Equal(t, uint32(0), journal.passTwoNumTxs)
	require.Equal(t, uint32(0), journal.passTwoNumSenders)
	require.Equal(t, uint32(0), journal.passThreeNumTxs)
	require.Equal(t, uint32(0), journal.passThreeNumSenders)

	// Bob and Carol evicted (lower total gas). Alice still there.
	_, ok := cache.GetByTxHash([]byte("hash-alice"))
	require.True(t, ok)
	require.Equal(t, int64(1), cache.CountSenders())
	require.Equal(t, int64(1), cache.CountTx())
}

func Test_DoEviction_DoneInPass1(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:             2,
		NumSendersToEvictInOneStep: 2,
	}

	cache := NewTxCacheWithEviction(16, config)
	cache.AddTx([]byte("hash-alice"), createTx("alice", uint64(1)))
	cache.AddTx([]byte("hash-bob"), createTx("bob", uint64(1)))
	cache.AddTx([]byte("hash-carol"), createTx("carol", uint64(1)))

	journal := cache.doEviction()
	require.Equal(t, uint32(0), journal.passZeroNumTxs)
	require.Equal(t, uint32(0), journal.passZeroNumSenders)
	require.Equal(t, uint32(2), journal.passOneNumTxs)
	require.Equal(t, uint32(2), journal.passOneNumSenders)
	require.Equal(t, uint32(0), journal.passTwoNumTxs)
	require.Equal(t, uint32(0), journal.passTwoNumSenders)
	require.Equal(t, uint32(0), journal.passThreeNumTxs)
	require.Equal(t, uint32(0), journal.passThreeNumSenders)

	// Alice and Bob evicted. Carol still there.
	_, ok := cache.GetByTxHash([]byte("hash-carol"))
	require.True(t, ok)
	require.Equal(t, int64(1), cache.CountSenders())
	require.Equal(t, int64(1), cache.CountTx())
}
