package txcache

import (
	"sync"
)

// This implementation is a simplified version of:
// https://github.com/ElrondNetwork/concurrent-map

// ConcurrentMap is a thread safe map of type string:Anything.
// To avoid lock bottlenecks this map is divided to several (ShardCount) map shards.
type ConcurrentMap struct {
	shardCount uint32
	shards     []*ConcurrentMapShard
}

// ConcurrentMapShard is a thread safe string to anything map.
type ConcurrentMapShard struct {
	maxSize uint32
	items   map[string]interface{}
	sync.RWMutex
}

// NewConcurrentMap creates a new concurrent map.
func NewConcurrentMap(maxSize uint32, shardCount uint32) *ConcurrentMap {
	m := ConcurrentMap{
		shardCount: shardCount,
		shards:     make([]*ConcurrentMapShard, shardCount),
	}

	shardSize := maxSize / shardCount
	if shardSize == 0 {
		shardSize = 1
	}
	if maxSize%shardCount != 0 {
		shardSize++
	}

	for i := uint32(0); i < shardCount; i++ {
		m.shards[i] = &ConcurrentMapShard{
			maxSize: shardSize,
			items:   make(map[string]interface{}),
		}
	}

	return &m
}

// GetShard returns shard under given key.
func (m *ConcurrentMap) GetShard(key string) *ConcurrentMapShard {
	return m.shards[fnv32(key)%m.shardCount]
}

// Set sets the given value under the specified key.
func (m *ConcurrentMap) Set(key string, value interface{}) {
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

// Get retrieves an element from map under given key.
func (m *ConcurrentMap) Get(key string) (interface{}, bool) {
	shard := m.GetShard(key)
	shard.RLock()
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

// Count returns the number of elements within the map.
func (m *ConcurrentMap) Count() int {
	count := 0
	for i := uint32(0); i < m.shardCount; i++ {
		shard := m.shards[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Has looks up an item under specified key.
func (m *ConcurrentMap) Has(key string) bool {
	shard := m.GetShard(key)
	shard.RLock()
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

// Remove removes an element from the map.
func (m *ConcurrentMap) Remove(key string) {
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// IsEmpty checks if map is empty.
func (m *ConcurrentMap) IsEmpty() bool {
	return m.Count() == 0
}

// IterCb is an iterator callback
type IterCb func(key string, v interface{})

// IterCb iterates over the map (cheapest way to read all elements in a map)
func (m *ConcurrentMap) IterCb(fn IterCb) {
	for idx := range m.shards {
		shard := (m.shards)[idx]
		shard.RLock()
		for key, value := range shard.items {
			fn(key, value)
		}
		shard.RUnlock()
	}
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
