package cache

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-storage/lrucache"
	"github.com/ElrondNetwork/elrond-go-storage/lrucache/capacity"
	"github.com/ElrondNetwork/elrond-go-storage/mapTimeCache"
	"github.com/ElrondNetwork/elrond-go-storage/timecache"
	"github.com/ElrondNetwork/elrond-go-storage/types"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// TimeCache is an alias for the imported TimeCache structure
type TimeCache = timecache.TimeCache

// EvictionHandler is an alias to the imported EvictionHandler
type EvictionHandler = types.EvictionHandler

// ArgMapTimeCacher is an alias for the imported ArgMapTimeCacher
type ArgMapTimeCacher = mapTimeCache.ArgMapTimeCacher

// TimeCacher defines the cache that can keep a record for a bounded time
type TimeCacher interface {
	Add(key string) error
	Upsert(key string, span time.Duration) error
	Has(key string) bool
	Sweep()
	RegisterEvictionHandler(handler EvictionHandler)
	IsInterfaceNil() bool
}

// PeerBlackListCacher can determine if a certain peer id is or not blacklisted
type PeerBlackListCacher interface {
	Upsert(pid core.PeerID, span time.Duration) error
	Has(pid core.PeerID) bool
	Sweep()
	IsInterfaceNil() bool
}

// SizedLRUCacheHandler is the interface for size capable LRU cache.
type SizedLRUCacheHandler interface {
	AddSized(key, value interface{}, sizeInBytes int64) bool
	Get(key interface{}) (value interface{}, ok bool)
	Contains(key interface{}) (ok bool)
	AddSizedIfMissing(key, value interface{}, sizeInBytes int64) (ok, evicted bool)
	Peek(key interface{}) (value interface{}, ok bool)
	Remove(key interface{}) bool
	Keys() []interface{}
	Len() int
	SizeInBytesContained() uint64
	Purge()
}

// AdaptedSizedLRUCache defines a cache that returns the evicted value
type AdaptedSizedLRUCache interface {
	SizedLRUCacheHandler
	AddSizedAndReturnEvicted(key, value interface{}, sizeInBytes int64) map[interface{}]interface{}
	IsInterfaceNil() bool
}

// NewTimeCache returns an instance of a time cache
func NewTimeCache(defaultSpan time.Duration) *TimeCache {
	return timecache.NewTimeCache(defaultSpan)
}

// NewLRUCache returns an instance of a LRU cache
func NewLRUCache(size int) (storage.Cacher, error) {
	return lrucache.NewCache(size)
}

// NewPeerTimeCache returns an instance of a peer time cacher
func NewPeerTimeCache(cache TimeCacher) (PeerBlackListCacher, error) {
	return timecache.NewPeerTimeCache(cache)
}

// NewCapacityLRU constructs an LRU cache of the given size with a byte size capacity
func NewCapacityLRU(size int, byteCapacity int64) (AdaptedSizedLRUCache, error) {
	return capacity.NewCapacityLRU(size, byteCapacity)
}

// NewLRUCacheWithEviction creates a new sized LRU cache instance with eviction function
func NewLRUCacheWithEviction(size int, onEvicted func(key interface{}, value interface{})) (storage.Cacher, error) {
	return lrucache.NewCacheWithEviction(size, onEvicted)
}

// NewMapTimeCache creates a new mapTimeCacher
func NewMapTimeCache(arg ArgMapTimeCacher) (storage.Cacher, error) {
	return mapTimeCache.NewMapTimeCache(arg)
}
