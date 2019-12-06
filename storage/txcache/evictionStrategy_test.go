package txcache

import "testing"

func Test_doArbitrarySendersEviction(t *testing.T) {
	cache := NewTxCache(100, 1)
	eviction := NewEvictionStrategy(1, cache)

	cache.AddTx([]byte("hash-alice"), createTx("alice", uint64(1)))
	cache.AddTx([]byte("hash-bob"), createTx("bob", uint64(1)))
	cache.AddTx([]byte("hash-carol"), createTx("carol", uint64(1)))

	eviction.doArbitrarySendersEviction()
}
