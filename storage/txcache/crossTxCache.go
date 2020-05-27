package txcache

import (
	"sync"
)

// crossTxCache holds cross-shard transactions (where destination == me)
type crossTxCache struct {
	config crossTxCacheConfig
	chunks []*crossTxChunk
	mutex  sync.RWMutex
}

func newCrossTxCache(config crossTxCacheConfig) *crossTxCache {
	cache := crossTxCache{
		config: config,
	}

	cache.initializeChunks()
	return &cache
}

func (cache *crossTxCache) initializeChunks() {
	// Assignment is not an atomic operation, so we have to wrap this in a critical section
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	config := cache.config
	chunkConfig := config.getChunkConfig()

	cache.chunks = make([]*crossTxChunk, config.numChunks)
	for i := uint32(0); i < config.numChunks; i++ {
		cache.chunks[i] = newCrossTxChunk(chunkConfig)
	}
}

func (cache *crossTxCache) immunizeKeys(keys []string) {
	groups := cache.groupKeysByChunk(keys)

	for chunkIndex, chunkKeys := range groups {
		chunk := cache.getChunkByIndex(chunkIndex)
		chunk.immunizeKeys(chunkKeys)
	}
}

// AddItem adds the item in the map
func (cache *crossTxCache) AddItem(item *WrappedTransaction) {
	key := string(item.TxHash)
	chunk := cache.getChunkByKey(key)
	chunk.addItem(item)
}

// Get gets an item from the map
func (cache *crossTxCache) Get(key string) (*WrappedTransaction, bool) {
	chunk := cache.getChunkByKey(key)
	return chunk.getItem(key)
}

// Has returns whether the item is in the map
func (cache *crossTxCache) Has(key string) bool {
	chunk := cache.getChunkByKey(key)
	_, ok := chunk.getItem(key)
	return ok
}

// Remove removes an element from the map
func (cache *crossTxCache) Remove(key string) {
	chunk := cache.getChunkByKey(key)
	chunk.removeItem(key)
}

func (cache *crossTxCache) RemoveOldest(numToRemove int) {
}

func (cache *crossTxCache) getChunkByKey(key string) *crossTxChunk {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	chunkIndex := cache.getChunkIndexByKey(key)

	return cache.chunks[chunkIndex]
}

func (cache *crossTxCache) getChunkByIndex(index uint32) *crossTxChunk {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.chunks[index]
}

func (cache *crossTxCache) groupKeysByChunk(keys []string) map[uint32][]string {
	groups := make(map[uint32][]string)

	for _, key := range keys {
		chunkIndex := cache.getChunkIndexByKey(key)
		groups[chunkIndex] = append(groups[chunkIndex], key)
	}

	return groups
}

func (cache *crossTxCache) getChunkIndexByKey(key string) uint32 {
	return fnv32Hash(key) % cache.config.numChunks
}

// fnv32Hash implements https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function for 32 bits
func fnv32Hash(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

// Clear clears the map
func (cache *crossTxCache) Clear() {
	// There is no need to explicitly remove each item for each chunk
	// The garbage collector will remove the data from memory
	cache.initializeChunks()
}

// Count returns the number of elements within the map
func (cache *crossTxCache) Count() uint32 {
	count := uint32(0)
	for _, chunk := range cache.getChunks() {
		count += chunk.countItems()
	}
	return count
}

// Keys returns all keys as []string
func (cache *crossTxCache) Keys() []string {
	count := cache.Count()
	// count is not exact anymore, since we are in a different lock than the one aquired by Count() (but is a good approximation)
	keys := make([]string, 0, count)

	for _, chunk := range cache.getChunks() {
		keys = chunk.appendKeys(keys)
	}

	return keys
}

func (cache *crossTxCache) getChunks() []*crossTxChunk {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.chunks
}
