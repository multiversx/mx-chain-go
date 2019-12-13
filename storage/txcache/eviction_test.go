package txcache

import "testing"

import "github.com/stretchr/testify/assert"

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

	assert.Equal(t, uint32(2), nTxs)
	assert.Equal(t, uint32(2), nSenders)
	assert.Equal(t, int64(1), cache.txListBySender.counter.Get())
	assert.Equal(t, int64(1), cache.txByHash.counter.Get())
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

	assert.Equal(t, int64(3), cache.txListBySender.counter.Get())
	assert.Equal(t, int64(401), cache.txByHash.counter.Get())

	nTxs, nSenders := cache.evictHighNonceTransactions()

	assert.Equal(t, uint32(50), nTxs)
	assert.Equal(t, uint32(0), nSenders)
	assert.Equal(t, int64(3), cache.txListBySender.counter.Get())
	assert.Equal(t, int64(351), cache.txByHash.counter.Get())
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

	assert.Equal(t, int64(200), cache.txListBySender.counter.Get())
	assert.Equal(t, int64(200), cache.txByHash.counter.Get())

	steps, nTxs, nSenders := cache.evictSendersWhileTooManyTxs()

	assert.Equal(t, uint32(6), steps)
	assert.Equal(t, uint32(100), nTxs)
	assert.Equal(t, uint32(100), nSenders)
	assert.Equal(t, int64(100), cache.txListBySender.counter.Get())
	assert.Equal(t, int64(100), cache.txByHash.counter.Get())
}
