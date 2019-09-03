package lrucache

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/hashicorp/golang-lru"
)

var log = logger.DefaultLogger()

// LRUCache implements a Least Recently Used eviction cache
type LRUCache struct {
	cache   *lru.Cache
	maxsize int

	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(key []byte)
}

// NewCache creates a new LRU cache instance
func NewCache(size int) (*LRUCache, error) {
	cache, err := lru.New(size)

	if err != nil {
		return nil, err
	}

	lruCache := &LRUCache{
		cache:                cache,
		maxsize:              size,
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(key []byte), 0),
	}

	return lruCache, nil
}

// Clear is used to completely clear the cache.
func (c *LRUCache) Clear() {
	c.cache.Purge()
}

// Put adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRUCache) Put(key []byte, value interface{}) (evicted bool) {
	evicted = c.cache.Add(string(key), value)

	c.callAddedDataHandlers(key)

	return evicted
}

// RegisterHandler registers a new handler to be called when a new data is added
func (c *LRUCache) RegisterHandler(handler func(key []byte)) {
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
	v, ok := c.cache.Get(string(key))
	if ok == false {
		return nil, ok
	}
	return v, ok
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

	if ok == false {
		return nil, ok
	}

	return v, ok
}

// HasOrAdd checks if a key is in the cache  without updating the
// recent-ness or deleting it for being stale,  and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *LRUCache) HasOrAdd(key []byte, value interface{}) (found, evicted bool) {
	found, evicted = c.cache.ContainsOrAdd(string(key), value)

	if !found {
		c.callAddedDataHandlers(key)
	}

	return
}

func (c *LRUCache) callAddedDataHandlers(key []byte) {
	c.mutAddedDataHandlers.RLock()
	for _, handler := range c.addedDataHandlers {
		go handler(key)
	}
	c.mutAddedDataHandlers.RUnlock()
}

// Remove removes the provided key from the cache.
func (c *LRUCache) Remove(key []byte) {
	c.cache.Remove(string(key))
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRUCache) RemoveOldest() {
	c.cache.RemoveOldest()
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

// IsInterfaceNil returns true if there is no value under the interface
func (c *LRUCache) IsInterfaceNil() bool {
	if c == nil {
		return true
	}
	return false
}
