package txcache

import "testing"

import "github.com/stretchr/testify/require"

func Test_EvictOldestSenders(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:          1,
		NumOldestSendersToEvict: 2,
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

func Test_DoHighNonceTransactionsEviction(t *testing.T) {
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

func Test_EvictSendersWhileTooManyTxs(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:          100,
		NumOldestSendersToEvict: 20,
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

func Test_DoEviction_DoneInPass1_WhenTooManySenders(t *testing.T) {
	config := EvictionConfig{
		CountThreshold:          2,
		NumOldestSendersToEvict: 2,
	}

	cache := NewTxCacheWithEviction(16, config)
	cache.AddTx([]byte("hash-alice"), createTx("alice", uint64(1)))
	cache.AddTx([]byte("hash-bob"), createTx("bob", uint64(1)))
	cache.AddTx([]byte("hash-carol"), createTx("carol", uint64(1)))

	journal := cache.doEviction()
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
