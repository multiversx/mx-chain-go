package ccache

import (
	"errors"
	"sync"

	"github.com/ElrondNetwork/concurrent-map"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
)

var log = logger.DefaultLogger()

// CCache implements a Concurrency type cache
type CCache struct {
	maxSize              int
	cache                cmap.ConcurrentMap
	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func([]byte)
}

// NewCCache instantiates a new concurrency map cache
func NewCCache(maxSize int) (*CCache, error) {
	if maxSize <= 0 {
		return nil, errors.New("must provide a positive value for cache size")
	}

	cache := cmap.New(maxSize)

	return &CCache{
		maxSize:              maxSize,
		cache:                cache,
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(key []byte), 0),
	}, nil
}

// Clear remove the cache items and clears the keys slice
func (c *CCache) Clear() {
	for key := range c.cache.Items() {
		c.cache.Remove(key)
	}
}

// Put inserts a new element into the cache.
// It also removes the oldest entry in case the cache size exeeds the maximum size.
func (c *CCache) Put(key []byte, value interface{}) (evicted bool) {
	c.cache.Set(string(key), value)
	// Remove the oldest values in case the cache size exceeds the maximum allowed size
	if c.cache.Count() > c.maxSize {
		c.RemoveOldest()
	}
	c.callAddedDataHandlers(key)

	return false
}

// Get retrieves a cache element based on the provided key
func (c *CCache) Get(key []byte) (interface{}, bool) {
	return c.cache.Get(string(key))
}

// Has checks if the cache contains the key
func (c *CCache) Has(key []byte) bool {
	if c.cache.Has(string(key)) {
		return true
	}
	return false
}

// Peek is identical with the Get method.
// It has been implemented only to satisfy the interface method signatures.
func (c *CCache) Peek(key []byte) (interface{}, bool) {
	return c.Get(key)
}

// HasOrAdd checks if a key exists in the cache map and if not inserts the value.
// In case the cache size exceeds the maximum allowed value, it removes the oldest entries.
func (c *CCache) HasOrAdd(key []byte, value interface{}) (ok, evicted bool) {
	if ok = c.cache.SetIfAbsent(string(key), value); ok {
		return ok, false
	}
	c.callAddedDataHandlers(key)
	return ok, false
}

// Remove removes the cache item based on the key
func (c *CCache) Remove(key []byte) {
	c.cache.Remove(string(key))
}

// RemoveOldest removes the oldest item from the cache
func (c *CCache) RemoveOldest() {
	if c.cache.IsEmpty() {
		return
	}
	c.cache.RemoveOldest()
}

// FindOldest finds the oldest entry
func (c *CCache) FindOldest() []byte {
	key := c.cache.FindOldest()

	if _, ok := c.Get([]byte(key)); ok {
		return []byte(key)
	}
	return nil
}

// Keys returns a slice of the keys in the cache, from oldest to newest
func (c *CCache) Keys() [][]byte {
	res := c.cache.Keys()
	r := make([][]byte, len(res))

	for i := 0; i < len(res); i++ {
		r[i] = []byte(res[i])
	}

	return r
}

// Len returns the number of items in the cache
func (c *CCache) Len() int {
	return c.cache.Count()
}

// RegisterHandler registers a new handler to be called when a new data is added
func (c *CCache) RegisterHandler(handler func(key []byte)) {
	if handler == nil {
		log.Error("attempt to register a nil handler to a cacher object")
		return
	}

	c.mutAddedDataHandlers.Lock()
	c.addedDataHandlers = append(c.addedDataHandlers, handler)
	c.mutAddedDataHandlers.Unlock()
}

func (c *CCache) callAddedDataHandlers(key []byte) {
	c.mutAddedDataHandlers.RLock()
	for _, handler := range c.addedDataHandlers {
		go handler(key)
	}
	c.mutAddedDataHandlers.RUnlock()
}
