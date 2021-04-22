package capacity

import (
	"container/list"
	"fmt"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("storage/lrucache/capacity")

// capacityLRU implements a non thread safe LRU Cache with a max capacity size
type capacityLRU struct {
	lock                   sync.Mutex
	size                   int
	maxCapacityInBytes     int64
	currentCapacityInBytes int64
	//TODO investigate if we can replace this list with a binary tree. Check also the other implementation lruCache
	evictList *list.List
	items     map[interface{}]*list.Element
}

// entry is used to hold a value in the evictList
type entry struct {
	key   interface{}
	value interface{}
	size  int64
}

// NewCapacityLRU constructs an CapacityLRU of the given size with a byte size capacity
func NewCapacityLRU(size int, byteCapacity int64) (*capacityLRU, error) {
	if size < 1 {
		return nil, storage.ErrCacheSizeInvalid
	}
	if byteCapacity < 1 {
		return nil, storage.ErrCacheCapacityInvalid
	}
	c := &capacityLRU{
		size:               size,
		maxCapacityInBytes: byteCapacity,
		evictList:          list.New(),
		items:              make(map[interface{}]*list.Element),
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *capacityLRU) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.items = make(map[interface{}]*list.Element)
	c.evictList.Init()
	c.currentCapacityInBytes = 0
}

// AddSized adds a value to the cache.  Returns true if an eviction occurred.
func (c *capacityLRU) AddSized(key, value interface{}, sizeInBytes int64) bool {
	if sizeInBytes < 0 {
		log.Error("size LRU cache add error",
			"key", fmt.Sprintf("%v", key),
			"value", fmt.Sprintf("%v", value),
			"error", storage.ErrNegativeSizeInBytes,
		)

		return false
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.update(key, value, sizeInBytes, ent)
	} else {
		c.addNew(key, value, sizeInBytes)
	}

	return c.evictIfNeeded()
}

func (c *capacityLRU) addNew(key interface{}, value interface{}, sizeInBytes int64) {
	ent := &entry{
		key:   key,
		value: value,
		size:  sizeInBytes,
	}
	e := c.evictList.PushFront(ent)
	c.items[key] = e
	c.currentCapacityInBytes += sizeInBytes
}

func (c *capacityLRU) update(key interface{}, value interface{}, sizeInBytes int64, ent *list.Element) {
	c.evictList.MoveToFront(ent)

	e := ent.Value.(*entry)
	sizeDiff := sizeInBytes - e.size
	e.value = value
	e.size = sizeInBytes
	c.currentCapacityInBytes += sizeDiff

	c.adjustSize(key, sizeInBytes)
}

// Get looks up a key's value from the cache.
func (c *capacityLRU) Get(key interface{}) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		if ent.Value.(*entry) == nil {
			return nil, false
		}

		return ent.Value.(*entry).value, true
	}

	return nil, false
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *capacityLRU) Contains(key interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, ok := c.items[key]

	return ok
}

// AddSizedIfMissing checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *capacityLRU) AddSizedIfMissing(key, value interface{}, sizeInBytes int64) (bool, bool) {
	if sizeInBytes < 0 {
		log.Error("size LRU cache contains or add error",
			"key", fmt.Sprintf("%v", key),
			"value", fmt.Sprintf("%v", value),
			"error", "size in bytes is negative",
		)

		return false, false
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	_, ok := c.items[key]
	if ok {
		return true, false
	}
	c.addNew(key, value, sizeInBytes)
	evicted := c.evictIfNeeded()

	return false, evicted
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *capacityLRU) Peek(key interface{}) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	ent, ok := c.items[key]
	if ok {
		return ent.Value.(*entry).value, true
	}
	return nil, ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *capacityLRU) Remove(key interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *capacityLRU) Keys() []interface{} {
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
func (c *capacityLRU) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.evictList.Len()
}

// SizeInBytesContained returns the size in bytes of all contained elements
func (c *capacityLRU) SizeInBytesContained() uint64 {
	c.lock.Lock()
	defer c.lock.Unlock()

	return uint64(c.currentCapacityInBytes)
}

// removeOldest removes the oldest item from the cache.
func (c *capacityLRU) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *capacityLRU) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(c.items, kv.key)
	c.currentCapacityInBytes -= kv.size
}

func (c *capacityLRU) adjustSize(key interface{}, sizeInBytes int64) {
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

func (c *capacityLRU) shouldEvict() bool {
	if c.evictList.Len() == 1 {
		// keep at least one element, no matter how large it is
		return false
	}

	return c.evictList.Len() > c.size || c.currentCapacityInBytes > c.maxCapacityInBytes
}

func (c *capacityLRU) evictIfNeeded() bool {
	evicted := false
	for c.shouldEvict() {
		c.removeOldest()
		evicted = true
	}

	return evicted
}
