package fifocache

import (
	"sync"

	"github.com/ElrondNetwork/concurrent-map"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
)

var log = logger.DefaultLogger()

// FIFOShardedCache implements a First In First Out eviction cache
type FIFOShardedCache struct {
	cache                *cmap.ConcurrentMap
	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(key []byte)
}

// NewShardedCache creates a new cache instance
func NewShardedCache(size int, shards int) (*FIFOShardedCache, error) {
	cache := cmap.New(size, shards)
	fifoShardedCache := &FIFOShardedCache{
		cache:                cache,
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(key []byte), 0),
	}

	return fifoShardedCache, nil
}

// Clear is used to completely clear the cache.
func (c *FIFOShardedCache) Clear() {
	keys := c.cache.Keys()
	for _, key := range keys {
		c.cache.Remove(key)
	}
}

// Put adds a value to the cache.  Returns true if an eviction occurred.
func (c *FIFOShardedCache) Put(key []byte, value interface{}) (evicted bool) {
	c.cache.Set(string(key), value)
	c.callAddedDataHandlers(key)

	return true
}

// RegisterHandler registers a new handler to be called when a new data is added
func (c *FIFOShardedCache) RegisterHandler(handler func(key []byte)) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a cacher object")
		return
	}

	c.mutAddedDataHandlers.Lock()
	c.addedDataHandlers = append(c.addedDataHandlers, handler)
	c.mutAddedDataHandlers.Unlock()
}

// Get looks up a key's value from the cache.
func (c *FIFOShardedCache) Get(key []byte) (value interface{}, ok bool) {
	return c.cache.Get(string(key))
}

// Has checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *FIFOShardedCache) Has(key []byte) bool {
	return c.cache.Has(string(key))
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *FIFOShardedCache) Peek(key []byte) (value interface{}, ok bool) {
	return c.cache.Get(string(key))
}

// HasOrAdd checks if a key is in the cache  without updating the
// recent-ness or deleting it for being stale,  and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *FIFOShardedCache) HasOrAdd(key []byte, value interface{}) (found, evicted bool) {
	added := c.cache.SetIfAbsent(string(key), value)

	if added {
		c.callAddedDataHandlers(key)
	}

	return found, true
}

func (c *FIFOShardedCache) callAddedDataHandlers(key []byte) {
	c.mutAddedDataHandlers.RLock()
	for _, handler := range c.addedDataHandlers {
		go handler(key)
	}
	c.mutAddedDataHandlers.RUnlock()
}

// Remove removes the provided key from the cache.
func (c *FIFOShardedCache) Remove(key []byte) {
	c.cache.Remove(string(key))
}

// RemoveOldest removes the oldest item from the cache.
func (c *FIFOShardedCache) RemoveOldest() {
	// nothing to do, oldest is automatically removed when adding a new item.
	log.Warn("remove oldest item not done, oldest item will be automatically cleared on reaching capacity")
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *FIFOShardedCache) Keys() [][]byte {
	res := c.cache.Keys()
	r := make([][]byte, len(res))

	for i := 0; i < len(res); i++ {
		r[i] = []byte(res[i])
	}

	return r
}

// Len returns the number of items in the cache.
func (c *FIFOShardedCache) Len() int {
	return c.cache.Count()
}
