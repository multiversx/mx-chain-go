package lrucache

import "github.com/hashicorp/golang-lru"

type LRUCache struct {
	cache   *lru.Cache
	maxsize int
}

// Create a new LRU cache instance
func NewCache(size int) (*LRUCache, error) {
	cache, err := lru.New(size)

	if err != nil {
		return nil, err
	}

	lruCache := &LRUCache{
		cache:   cache,
		maxsize: size,
	}

	return lruCache, nil
}

// Clear is used to completely clear the cache.
func (c *LRUCache) Clear() {
	c.cache.Purge()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRUCache) Add(key, value []byte) (evicted bool) {
	return c.cache.Add(string(key), value)
}

// Get looks up a key's value from the cache.
func (c *LRUCache) Get(key []byte) (value []byte, ok bool) {
	v, ok := c.cache.Get(string(key))
	if ok == false {
		return nil, ok
	}
	value = v.([]byte)
	return value, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *LRUCache) Contains(key []byte) bool {
	return c.cache.Contains(string(key))
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRUCache) Peek(key []byte) (value []byte, ok bool) {
	v, ok := c.cache.Peek(string(key))

	if ok == false {
		return nil, ok
	}

	value = v.([]byte)
	return value, ok
}

// ContainsOrAdd checks if a key is in the cache  without updating the
// recent-ness or deleting it for being stale,  and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *LRUCache) ContainsOrAdd(key, value []byte) (ok, evicted bool) {
	return c.cache.ContainsOrAdd(string(key), value)
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
