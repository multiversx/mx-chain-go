package txcache

import "testing"

import "github.com/stretchr/testify/assert"

func Test_EvictOldestSenders(t *testing.T) {
	cache := NewTxCache(100, 1)
	config := EvictionStrategyConfig{
		CountThreshold:         1,
		NoOldestSendersToEvict: 2,
	}
	eviction := NewEvictionStrategy(cache, config)

	cache.AddTx([]byte("hash-alice"), createTx("alice", uint64(1)))
	cache.AddTx([]byte("hash-bob"), createTx("bob", uint64(1)))
	cache.AddTx([]byte("hash-carol"), createTx("carol", uint64(1)))

	noTxs, noSenders := eviction.EvictOldestSenders()

	assert.Equal(t, uint32(2), noTxs)
	assert.Equal(t, uint32(2), noSenders)
	assert.Equal(t, int64(1), cache.txListBySender.Counter.Get())
	assert.Equal(t, int64(1), cache.txByHash.Counter.Get())
}

func Test_DoHighNonceTransactionsEviction(t *testing.T) {
	cache := NewTxCache(300, 1)
	config := EvictionStrategyConfig{
		CountThreshold:                 400,
		ALotOfTransactionsForASender:   50,
		NoTxsToEvictForASenderWithALot: 25,
	}
	eviction := NewEvictionStrategy(cache, config)

	for index := 0; index < 200; index++ {
		cache.AddTx([]byte{'a', byte(index)}, createTx("alice", uint64(index)))
	}

	for index := 0; index < 200; index++ {
		cache.AddTx([]byte{'b', byte(index)}, createTx("bob", uint64(index)))
	}

	cache.AddTx([]byte("hash-carol"), createTx("carol", uint64(1)))

	assert.Equal(t, int64(3), cache.txListBySender.Counter.Get())
	assert.Equal(t, int64(401), cache.txByHash.Counter.Get())

	noTxs, noSenders := eviction.EvictHighNonceTransactions()

	assert.Equal(t, uint32(50), noTxs)
	assert.Equal(t, uint32(0), noSenders)
	assert.Equal(t, int64(3), cache.txListBySender.Counter.Get())
	assert.Equal(t, int64(351), cache.txByHash.Counter.Get())
}

func Test_EvictSendersWhileTooManyTxs(t *testing.T) {
	cache := NewTxCache(300, 1)
	config := EvictionStrategyConfig{
		CountThreshold:         100,
		NoOldestSendersToEvict: 20,
	}
	eviction := NewEvictionStrategy(cache, config)

	// 200 senders, each with 1 transaction
	for index := 0; index < 200; index++ {
		sender := string(createFakeSenderAddress(index))
		cache.AddTx([]byte{byte(index)}, createTx(sender, uint64(1)))
	}

	assert.Equal(t, int64(200), cache.txListBySender.Counter.Get())
	assert.Equal(t, int64(200), cache.txByHash.Counter.Get())

	steps, noTxs, noSenders := eviction.EvictSendersWhileTooManyTxs()

	assert.Equal(t, uint32(6), steps)
	assert.Equal(t, uint32(100), noTxs)
	assert.Equal(t, uint32(100), noSenders)
	assert.Equal(t, int64(100), cache.txListBySender.Counter.Get())
	assert.Equal(t, int64(100), cache.txByHash.Counter.Get())
}
