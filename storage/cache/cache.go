package cache

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-storage/immunitycache"
	"github.com/ElrondNetwork/elrond-go-storage/lrucache"
	"github.com/ElrondNetwork/elrond-go-storage/lrucache/capacity"
	"github.com/ElrondNetwork/elrond-go-storage/timecache"
	"github.com/ElrondNetwork/elrond-go-storage/types"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgTimeCacher is the argument used to create a new timeCacher instance
type ArgTimeCacher = timecache.ArgTimeCacher

// TimeCache is an alias for the imported TimeCache structure
type TimeCache = timecache.TimeCache

// EvictionHandler is an alias to the imported EvictionHandler
type EvictionHandler = types.EvictionHandler

// ImmunityCache is a cache-like structure
type ImmunityCache = immunitycache.ImmunityCache

// CacheConfig holds cache configuration
type CacheConfig = immunitycache.CacheConfig

// TimeCacher defines the cache that can keep a record for a bounded time
type TimeCacher interface {
	Add(key string) error
	Upsert(key string, span time.Duration) error
	Has(key string) bool
	Sweep()
	IsInterfaceNil() bool
}

// PeerBlackListCacher can determine if a certain peer id is or not blacklisted
type PeerBlackListCacher interface {
	Upsert(pid core.PeerID, span time.Duration) error
	Has(pid core.PeerID) bool
	Sweep()
	IsInterfaceNil() bool
}

// NewTimeCache returns an instance of a time cache
func NewTimeCache(defaultSpan time.Duration) *timecache.TimeCache {
	return timecache.NewTimeCache(defaultSpan)
}

// NewTimeCacher creates a new timeCacher
func NewTimeCacher(arg ArgTimeCacher) (storage.Cacher, error) {
	return timecache.NewTimeCacher(arg)
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
func NewCapacityLRU(size int, byteCapacity int64) (storage.AdaptedSizedLRUCache, error) {
	return capacity.NewCapacityLRU(size, byteCapacity)
}

// NewLRUCacheWithEviction creates a new sized LRU cache instance with eviction function
func NewLRUCacheWithEviction(size int, onEvicted func(key interface{}, value interface{})) (storage.Cacher, error) {
	return lrucache.NewCacheWithEviction(size, onEvicted)
}

// NewImmunityCache creates a new cache
func NewImmunityCache(config CacheConfig) (*immunitycache.ImmunityCache, error) {
	return immunitycache.NewImmunityCache(config)
}
