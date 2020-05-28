package immunitycache

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ storage.Cacher = (*ImmunityCache)(nil)

var log = logger.GetOrCreate("ImmunityCache")

// ImmunityCache is a cache-like structure
type ImmunityCache struct {
	config CacheConfig
	chunks []*immunityChunk
	mutex  sync.RWMutex
}

// NewImmunityCache creates a new cache
func NewImmunityCache(config CacheConfig) (*ImmunityCache, error) {
	log.Debug("NewImmunityCache", "config", config.String())

	err := config.verify()
	if err != nil {
		return nil, err
	}

	cache := ImmunityCache{
		config: config,
	}

	cache.initializeChunksWithLock()
	return &cache, nil
}

func (cache *ImmunityCache) initializeChunksWithLock() {
	// Assignment is not an atomic operation, so we have to wrap this in a critical section
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	config := cache.config
	chunkConfig := config.getChunkConfig()

	cache.chunks = make([]*immunityChunk, config.NumChunks)
	for i := uint32(0); i < config.NumChunks; i++ {
		cache.chunks[i] = newImmunityChunk(chunkConfig)
	}
}

// ImmunizeKeys marks items as immune to eviction
func (cache *ImmunityCache) ImmunizeKeys(keys [][]byte) (numNowTotal, numFutureTotal int) {
	groups := cache.groupKeysByChunk(keys)

	for chunkIndex, chunkKeys := range groups {
		chunk := cache.getChunkByIndexWithLock(chunkIndex)

		numNow, numFuture := chunk.ImmunizeKeys(chunkKeys)
		numNowTotal += numNow
		numFutureTotal += numFuture
	}

	return
}

func (cache *ImmunityCache) groupKeysByChunk(keys [][]byte) map[uint32][][]byte {
	groups := make(map[uint32][][]byte)

	for _, key := range keys {
		chunkIndex := cache.getChunkIndexByKey(string(key))
		groups[chunkIndex] = append(groups[chunkIndex], key)
	}

	return groups
}

func (cache *ImmunityCache) getChunkIndexByKey(key string) uint32 {
	return fnv32Hash(key) % cache.config.NumChunks
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

func (cache *ImmunityCache) getChunkByIndexWithLock(index uint32) *immunityChunk {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.chunks[index]
}

// Add adds an item in the cache
func (cache *ImmunityCache) Add(item CacheItem) (ok bool, added bool) {
	key := string(item.GetKey())
	chunk := cache.getChunkByKeyWithLock(key)
	return chunk.AddItem(item)
}

func (cache *ImmunityCache) getChunkByKeyWithLock(key string) *immunityChunk {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	chunkIndex := cache.getChunkIndexByKey(key)
	return cache.chunks[chunkIndex]
}

// Get gets an item (payload) by key
func (cache *ImmunityCache) Get(key []byte) (value interface{}, ok bool) {
	item, ok := cache.GetItem(key)
	if ok {
		return item.Payload(), true
	}
	return nil, false
}

// GetItem gets an item by key
func (cache *ImmunityCache) GetItem(key []byte) (CacheItem, bool) {
	chunk := cache.getChunkByKeyWithLock(string(key))
	return chunk.GetItem(string(key))
}

// Has checks is an item exists
func (cache *ImmunityCache) Has(key []byte) bool {
	chunk := cache.getChunkByKeyWithLock(string(key))
	_, ok := chunk.GetItem(string(key))
	return ok
}

// Peek gets an item
func (cache *ImmunityCache) Peek(key []byte) (value interface{}, ok bool) {
	return cache.Get(key)
}

// HasOrAdd is not implemented
func (cache *ImmunityCache) HasOrAdd(_ []byte, _ interface{}) (ok, evicted bool) {
	log.Error("ImmunityCache.HasOrAdd is not implemented")
	return false, false
}

// Put is not implemented
func (cache *ImmunityCache) Put(_ []byte, _ interface{}) (evicted bool) {
	log.Error("ImmunityCache.Put is not implemented")
	return false
}

// Remove removes an item
func (cache *ImmunityCache) Remove(key []byte) {
	_ = cache.RemoveWithResult(key)
}

// RemoveWithResult removes an item
// TODO: In the future, add this method to the "storage.Cacher" interface
func (cache *ImmunityCache) RemoveWithResult(key []byte) bool {
	chunk := cache.getChunkByKeyWithLock(string(key))
	return chunk.RemoveItem(string(key))
}

// RemoveOldest is not implemented
func (cache *ImmunityCache) RemoveOldest() {
	log.Error("ImmunityCache.RemoveOldest is not implemented")
}

// Clear clears the map
func (cache *ImmunityCache) Clear() {
	// There is no need to explicitly remove each item for each chunk
	// The garbage collector will remove the data from memory
	cache.initializeChunksWithLock()
}

// MaxSize returns the capacity of the cache
func (cache *ImmunityCache) MaxSize() int {
	return int(cache.config.MaxNumItems)
}

// Len is an alias for Count
func (cache *ImmunityCache) Len() int {
	return cache.Count()
}

// Count returns the number of elements within the map
func (cache *ImmunityCache) Count() int {
	count := 0
	for _, chunk := range cache.getChunksWithLock() {
		count += chunk.Count()
	}
	return count
}

func (cache *ImmunityCache) getChunksWithLock() []*immunityChunk {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.chunks
}

// CountImmune returns the number of immunized (current or future) elements within the map
func (cache *ImmunityCache) CountImmune() int {
	count := 0
	for _, chunk := range cache.getChunksWithLock() {
		count += chunk.CountImmune()
	}
	return count
}

// NumBytes estimates the size of the cache, in bytes
func (cache *ImmunityCache) NumBytes() int {
	numBytes := 0
	for _, chunk := range cache.getChunksWithLock() {
		numBytes += chunk.NumBytes()
	}
	return numBytes
}

// Keys returns all keys
func (cache *ImmunityCache) Keys() [][]byte {
	count := cache.Count()
	// count is not exact anymore, since we are in a different lock than the one aquired by Count() (but is a good approximation)
	keys := make([][]byte, 0, count)

	for _, chunk := range cache.getChunksWithLock() {
		keys = chunk.AppendKeys(keys)
	}

	return keys
}

// RegisterHandler is not implemented
func (cache *ImmunityCache) RegisterHandler(func(key []byte, value interface{})) {
	log.Error("ImmunityCache.RegisterHandler is not implemented")
}

// ForEachItem iterates over the items in the cache
func (cache *ImmunityCache) ForEachItem(function ForEachItem) {
	for _, chunk := range cache.getChunksWithLock() {
		chunk.ForEachItem(function)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *ImmunityCache) IsInterfaceNil() bool {
	return cache == nil
}
