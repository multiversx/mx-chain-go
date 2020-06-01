package lrucache

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache/capacity"
	lru "github.com/hashicorp/golang-lru"
)

var _ storage.Cacher = (*LRUCache)(nil)

var log = logger.GetOrCreate("storage/lrucache")

// LRUCache implements a Least Recently Used eviction cache
type LRUCache struct {
	cache   storage.SizeLRUCacheHandler
	maxsize int

	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(key []byte, value interface{})
}

// NewCache creates a new LRU cache instance
func NewCache(size int) (*LRUCache, error) {
	cache, err := lru.New(size)
	if err != nil {
		return nil, err
	}

	lruCache := &LRUCache{
		cache: &simpleLRUCacheAdapter{
			LRUCacheHandler: cache,
		},
		maxsize:              size,
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(key []byte, value interface{}), 0),
	}

	return lruCache, nil
}

// NewCacheWithSizeInBytes creates a new sized LRU cache instance
func NewCacheWithSizeInBytes(size int, sizeInBytes int64) (*LRUCache, error) {
	cache, err := capacity.NewCapacityLRU(size, sizeInBytes)
	if err != nil {
		return nil, err
	}

	lruCache := &LRUCache{
		cache:                cache,
		maxsize:              size,
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(key []byte, value interface{}), 0),
	}

	return lruCache, nil
}

// Clear is used to completely clear the cache.
func (c *LRUCache) Clear() {
	c.cache.Purge()
}

// Put adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRUCache) Put(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
	evicted = c.cache.AddSized(string(key), value, int64(sizeInBytes))

	c.callAddedDataHandlers(key, value)

	return evicted
}

// RegisterHandler registers a new handler to be called when a new data is added
func (c *LRUCache) RegisterHandler(handler func(key []byte, value interface{})) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a cacher object")
		return
	}

	c.mutAddedDataHandlers.Lock()
	c.addedDataHandlers = append(c.addedDataHandlers, handler)
	c.mutAddedDataHandlers.Unlock()
}

// Get looks up a key's value from the cache.
func (c *LRUCache) Get(key []byte) (value interface{}, ok bool) {
	return c.cache.Get(string(key))
}

// Has checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *LRUCache) Has(key []byte) bool {
	return c.cache.Contains(string(key))
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRUCache) Peek(key []byte) (value interface{}, ok bool) {
	v, ok := c.cache.Peek(string(key))

	if !ok {
		return nil, ok
	}

	return v, ok
}

// HasOrAdd checks if a key is in the cache  without updating the
// recent-ness or deleting it for being stale,  and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *LRUCache) HasOrAdd(key []byte, value interface{}, sizeInBytes int) (found, evicted bool) {
	found, evicted = c.cache.ContainsOrAddSized(string(key), value, int64(sizeInBytes))

	if !found {
		c.callAddedDataHandlers(key, value)
	}

	return
}

func (c *LRUCache) callAddedDataHandlers(key []byte, value interface{}) {
	c.mutAddedDataHandlers.RLock()
	for _, handler := range c.addedDataHandlers {
		go handler(key, value)
	}
	c.mutAddedDataHandlers.RUnlock()
}

// Remove removes the provided key from the cache.
func (c *LRUCache) Remove(key []byte) {
	c.cache.Remove(string(key))
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRUCache) Keys() [][]byte {
	res := c.cache.Keys()
	r := make([][]byte, len(res))

	for i := 0; i < len(res); i++ {
		r[i] = []byte(res[i].(string))
	}

	return r
}

// Len returns the number of items in the cache.
func (c *LRUCache) Len() int {
	return c.cache.Len()
}

// MaxSize returns the maximum number of items which can be stored in cache.
func (c *LRUCache) MaxSize() int {
	return c.maxsize
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *LRUCache) IsInterfaceNil() bool {
	return c == nil
}
