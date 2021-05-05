package fifocache

import (
	"sync"

	cmap "github.com/ElrondNetwork/concurrent-map"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ storage.Cacher = (*FIFOShardedCache)(nil)

var log = logger.GetOrCreate("storage/fifocache")

// FIFOShardedCache implements a First In First Out eviction cache
type FIFOShardedCache struct {
	cache   *cmap.ConcurrentMap
	maxsize int

	mutAddedDataHandlers sync.RWMutex
	mapDataHandlers      map[string]func(key []byte, value interface{})
}

// NewShardedCache creates a new cache instance
func NewShardedCache(size int, shards int) (*FIFOShardedCache, error) {
	cache := cmap.New(size, shards)
	fifoShardedCache := &FIFOShardedCache{
		cache:                cache,
		maxsize:              size,
		mutAddedDataHandlers: sync.RWMutex{},
		mapDataHandlers:      make(map[string]func(key []byte, value interface{})),
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
// the int parameter for size is not used as, for now, fifo sharded cache can not count for its contained data size
func (c *FIFOShardedCache) Put(key []byte, value interface{}, _ int) (evicted bool) {
	c.cache.Set(string(key), value)
	c.callAddedDataHandlers(key, value)

	return true
}

// RegisterHandler registers a new handler to be called when a new data is added
func (c *FIFOShardedCache) RegisterHandler(handler func(key []byte, value interface{}), id string) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a cacher object")
		return
	}

	c.mutAddedDataHandlers.Lock()
	c.mapDataHandlers[id] = handler
	c.mutAddedDataHandlers.Unlock()
}

// UnRegisterHandler removes the handler from the list
func (c *FIFOShardedCache) UnRegisterHandler(id string) {
	c.mutAddedDataHandlers.Lock()
	delete(c.mapDataHandlers, id)
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

// HasOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether the item existed before and whether it has been added.
func (c *FIFOShardedCache) HasOrAdd(key []byte, value interface{}, _ int) (has, added bool) {
	added = c.cache.SetIfAbsent(string(key), value)

	if added {
		c.callAddedDataHandlers(key, value)
	}

	return !added, added
}

func (c *FIFOShardedCache) callAddedDataHandlers(key []byte, value interface{}) {
	c.mutAddedDataHandlers.RLock()
	for _, handler := range c.mapDataHandlers {
		go handler(key, value)
	}
	c.mutAddedDataHandlers.RUnlock()
}

// Remove removes the provided key from the cache.
func (c *FIFOShardedCache) Remove(key []byte) {
	c.cache.Remove(string(key))
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

// SizeInBytesContained returns 0
func (c *FIFOShardedCache) SizeInBytesContained() uint64 {
	return 0
}

// MaxSize returns the maximum number of items which can be stored in cache.
func (c *FIFOShardedCache) MaxSize() int {
	return c.maxsize
}

// Close does nothing for this cacher implementation
func (c *FIFOShardedCache) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *FIFOShardedCache) IsInterfaceNil() bool {
	return c == nil
}
