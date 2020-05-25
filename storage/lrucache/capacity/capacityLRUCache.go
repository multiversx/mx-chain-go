package capacity

import (
	"container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go/storage"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(key interface{}, value interface{})

// CapacityLRU implements a non thread safe LRU Cache with a max capacity size
type CapacityLRU struct {
	lock                   sync.Mutex
	size                   int
	maxCapacityInBytes     int64
	currentCapacityInBytes int64
	evictList              *list.List
	items                  map[interface{}]*list.Element
	onEvict                EvictCallback
}

// entry is used to hold a value in the evictList
type entry struct {
	key   interface{}
	value interface{}
	size  int64
}

// NewCapacityLRU constructs an CapacityLRU of the given size with a byte size capacity
func NewCapacityLRU(size int, byteCapacity int64, onEvict EvictCallback) (*CapacityLRU, error) {
	if size <= 0 {
		return nil, storage.ErrCacheSizeInvalid
	}
	if byteCapacity <= 0 {
		return nil, storage.ErrCacheCapacityInvalid
	}
	c := &CapacityLRU{
		size:               size,
		maxCapacityInBytes: byteCapacity,
		evictList:          list.New(),
		items:              make(map[interface{}]*list.Element),
		onEvict:            onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *CapacityLRU) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value.(*entry).value)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *CapacityLRU) Add(key, value interface{}, sizeInBytes int64) (evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.update(key, value, sizeInBytes, ent)
	} else {
		c.addNew(key, value, sizeInBytes)
	}

	evict := c.evictIfNeeded()

	return evict
}

func (c *CapacityLRU) addNew(key interface{}, value interface{}, sizeInBytes int64) {
	ent := &entry{key, value, sizeInBytes}
	entry := c.evictList.PushFront(ent)
	c.items[key] = entry
	c.currentCapacityInBytes += sizeInBytes
}

func (c *CapacityLRU) update(key interface{}, value interface{}, sizeInBytes int64, ent *list.Element) {
	c.evictList.MoveToFront(ent)

	e := ent.Value.(*entry)
	sizeDiff := sizeInBytes - e.size
	e.value = value
	e.size = sizeInBytes
	c.currentCapacityInBytes += sizeDiff

	c.adjustSize(key, sizeInBytes)
}

// Get looks up a key's value from the cache.
func (c *CapacityLRU) Get(key interface{}) (value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		if ent.Value.(*entry) == nil {
			return nil, false
		}
		return ent.Value.(*entry).value, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *CapacityLRU) Contains(key interface{}) (ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, ok = c.items[key]
	return ok
}

// HasOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *CapacityLRU) HasOrAdd(key, value interface{}, sizeInBytes int64) (ok, evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, ok = c.items[key]
	if ok {
		return true, false
	}
	c.addNew(key, value, sizeInBytes)
	evicted = c.evictIfNeeded()

	return false, evicted
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *CapacityLRU) Peek(key interface{}) (value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var ent *list.Element
	if ent, ok = c.items[key]; ok {
		return ent.Value.(*entry).value, true
	}
	return nil, ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *CapacityLRU) Remove(key interface{}) (present bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// RemoveOldest removes the oldest item from the cache.
func (c *CapacityLRU) RemoveOldest() (key interface{}, value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

// GetOldest returns the oldest entry
func (c *CapacityLRU) GetOldest() (key interface{}, value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	ent := c.evictList.Back()
	if ent != nil {
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *CapacityLRU) Keys() []interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()

	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *CapacityLRU) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.evictList.Len()
}

// removeOldest removes the oldest item from the cache.
func (c *CapacityLRU) removeOldest() {
	c.lock.Lock()
	defer c.lock.Unlock()

	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *CapacityLRU) removeElement(e *list.Element) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(c.items, kv.key)
	c.currentCapacityInBytes -= kv.size

	if c.onEvict != nil {
		c.onEvict(kv.key, kv.value)
	}
}

func (c *CapacityLRU) adjustSize(key interface{}, sizeInBytes int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	element := c.items[key]
	if element == nil || element.Value == nil || element.Value.(*entry) == nil {
		return
	}

	v := element.Value.(*entry)
	c.currentCapacityInBytes -= v.size
	v.size = sizeInBytes
	element.Value = v
	c.currentCapacityInBytes += sizeInBytes
	c.evictIfNeeded()
}

func (c *CapacityLRU) evictIfNeeded() (evicted bool) {
	evict := c.evictList.Len() > c.size || c.currentCapacityInBytes > c.maxCapacityInBytes

	if evict {
		c.removeOldest()
	}

	for c.currentCapacityInBytes > c.maxCapacityInBytes {
		c.removeOldest()
	}

	return evict
}
