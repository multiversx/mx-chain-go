package ccache

import (
	"errors"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/concurrent-map"
)

var log = logger.DefaultLogger()

// CCache implements a Concurrency type cache
type CCache struct {
	maxSize              int
	mapKeys              []string
	cache                cmap.ConcurrentMap
	mkGuard              sync.RWMutex
	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func([]byte)
}

// NewCCache instantiates a new concurrency map cache
func NewCCache(maxSize int) (*CCache, error) {
	if maxSize <= 0 {
		return nil, errors.New("must provide a positive value for cache size")
	}

	cache := cmap.New()

	return &CCache{
		maxSize:              maxSize,
		mapKeys:              nil,
		cache:                cache,
		mkGuard:              sync.RWMutex{},
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(key []byte), 0),
	}, nil
}

// Clear remove the cache items and clears the keys slice
func (c *CCache) Clear() {
	for key := range c.cache.Items() {
		c.cache.Remove(key)
	}
	c.mkGuard.Lock()
	c.mapKeys = nil
	c.mkGuard.Unlock()
}

// Put inserts a new element into the cache.
// It also removes the oldest entry in case the cache size exeeds the maximum size.
func (c *CCache) Put(key []byte, value interface{}) (evicted bool) {
	// Remove the oldest values in case the cache size exceeds the maximum allowed size
	if c.cache.Count() > c.maxSize {
		c.RemoveOldest()
	}

	// Save the keys into a separate slice prior inserting the key/value into the map
	c.mkGuard.Lock()
	c.mapKeys = append(c.mapKeys, string(key))
	c.mkGuard.Unlock()

	c.cache.Set(string(key), value)

	c.callAddedDataHandlers(key)

	return false
}

// Get retrieves a cache element based on the provided key
func (c *CCache) Get(key []byte) (interface{}, bool) {
	item, ok := c.cache.Get(string(key))
	if !ok {
		return nil, false
	}
	return item, ok
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
		c.mkGuard.Lock()
		c.mapKeys = append(c.mapKeys, string(key))
		c.mkGuard.Unlock()

		return ok, false
	}

	c.callAddedDataHandlers(key)
	if c.cache.Count() > c.maxSize {
		c.RemoveOldest()
	}

	return ok, false
}

// Remove removes the cache item based on the key
func (c *CCache) Remove(key []byte) {
	c.removeMapKey(key)
	c.cache.Remove(string(key))
}

// RemoveOldest removes the oldest item from the cache
func (c *CCache) RemoveOldest() {
	if c.cache.IsEmpty() {
		return
	}

	oldest := c.FindOldest()
	c.cache.Remove(string(oldest))

	c.removeMapKey(oldest)
}

// FindOldest finds the oldest entry
func (c *CCache) FindOldest() []byte {
	c.mkGuard.Lock()
	key := c.mapKeys[0]
	c.mkGuard.Unlock()

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

// removeMapKey removes the corresponding map key from the keys slice
func (c *CCache) removeMapKey(key []byte) {
	c.mkGuard.Lock()

	if len(c.mapKeys) > 0 {
		for i := 0; i < len(c.mapKeys); i++ {
			if string(key) == c.mapKeys[i] {
				// delete key from keys slice
				copy(c.mapKeys[i:], c.mapKeys[i+1:])
				c.mapKeys = c.mapKeys[:len(c.mapKeys)-1]
			}
		}
	}
	c.mkGuard.Unlock()
}
