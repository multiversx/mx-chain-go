package immunitycache

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ storage.Cacher = (*ImmunityCache)(nil)

var log = logger.GetOrCreate("storage/immunitycache")

// ImmunityCache is a cache-like structure
type ImmunityCache struct {
	config CacheConfig
	chunks []*immunityChunk
	mutex  sync.RWMutex
}

// NewImmunityCache creates a new cache
func NewImmunityCache(config CacheConfig) (*ImmunityCache, error) {
	log.Debug("NewImmunityCache", "config", config.String())
	storage.MonitorNewCache(config.Name, uint64(config.MaxNumBytes))

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

func (ic *ImmunityCache) initializeChunksWithLock() {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	config := ic.config
	chunkConfig := config.getChunkConfig()

	ic.chunks = make([]*immunityChunk, config.NumChunks)
	for i := uint32(0); i < config.NumChunks; i++ {
		ic.chunks[i] = newImmunityChunk(chunkConfig)
	}
}

// ImmunizeKeys marks items as immune to eviction
func (ic *ImmunityCache) ImmunizeKeys(keys [][]byte) (numNowTotal, numFutureTotal int) {
	groups := ic.groupKeysByChunk(keys)

	for chunkIndex, chunkKeys := range groups {
		chunk := ic.getChunkByIndexWithLock(chunkIndex)

		numNow, numFuture := chunk.ImmunizeKeys(chunkKeys)
		numNowTotal += numNow
		numFutureTotal += numFuture
	}

	return
}

func (ic *ImmunityCache) groupKeysByChunk(keys [][]byte) map[uint32][][]byte {
	groups := make(map[uint32][][]byte)

	for _, key := range keys {
		chunkIndex := ic.getChunkIndexByKey(string(key))
		groups[chunkIndex] = append(groups[chunkIndex], key)
	}

	return groups
}

func (ic *ImmunityCache) getChunkIndexByKey(key string) uint32 {
	return fnv32Hash(key) % ic.config.NumChunks
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

func (ic *ImmunityCache) getChunkByIndexWithLock(index uint32) *immunityChunk {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()
	return ic.chunks[index]
}

// Add adds an item in the cache
func (ic *ImmunityCache) Add(item storage.CacheItem) (ok bool, added bool) {
	key := string(item.GetKey())
	chunk := ic.getChunkByKeyWithLock(key)
	return chunk.AddItem(item)
}

func (ic *ImmunityCache) getChunkByKeyWithLock(key string) *immunityChunk {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	chunkIndex := ic.getChunkIndexByKey(key)
	return ic.chunks[chunkIndex]
}

// Get gets an item (payload) by key
func (ic *ImmunityCache) Get(key []byte) (value interface{}, ok bool) {
	return ic.GetItem(key)
}

// GetItem gets an item by key
func (ic *ImmunityCache) GetItem(key []byte) (storage.CacheItem, bool) {
	chunk := ic.getChunkByKeyWithLock(string(key))
	return chunk.GetItem(string(key))
}

// Has checks is an item exists
func (ic *ImmunityCache) Has(key []byte) bool {
	chunk := ic.getChunkByKeyWithLock(string(key))
	_, ok := chunk.GetItem(string(key))
	return ok
}

// Peek gets an item
func (ic *ImmunityCache) Peek(key []byte) (value interface{}, ok bool) {
	return ic.Get(key)
}

// HasOrAdd adds an item in the cache
func (ic *ImmunityCache) HasOrAdd(_ []byte, value interface{}, _ int) (ok, evicted bool) {
	valueAsCacheItem, ok := value.(storage.CacheItem)
	if !ok {
		return false, false
	}

	ok, _ = ic.Add(valueAsCacheItem)
	return ok, false
}

// Put adds an item in the cache
func (ic *ImmunityCache) Put(key []byte, value interface{}, _ int) (evicted bool) {
	ic.HasOrAdd(key, value, 0)
	return false
}

// Remove removes an item
func (ic *ImmunityCache) Remove(key []byte) {
	_ = ic.RemoveWithResult(key)
}

// RemoveWithResult removes an item
// TODO: In the future, add this method to the "storage.Cacher" interface
func (ic *ImmunityCache) RemoveWithResult(key []byte) bool {
	chunk := ic.getChunkByKeyWithLock(string(key))
	return chunk.RemoveItem(string(key))
}

// RemoveOldest is not implemented
func (ic *ImmunityCache) RemoveOldest() {
	log.Error("ImmunityCache.RemoveOldest is not implemented")
}

// Clear clears the map
func (ic *ImmunityCache) Clear() {
	// There is no need to explicitly remove each item for each chunk
	// The garbage collector will remove the data from memory
	ic.initializeChunksWithLock()
}

// MaxSize returns the capacity of the cache
func (ic *ImmunityCache) MaxSize() int {
	return int(ic.config.MaxNumItems)
}

// Len is an alias for Count
func (ic *ImmunityCache) Len() int {
	return ic.Count()
}

// Count returns the number of elements within the map
func (ic *ImmunityCache) Count() int {
	count := 0
	for _, chunk := range ic.getChunksWithLock() {
		count += chunk.Count()
	}
	return count
}

func (ic *ImmunityCache) getChunksWithLock() []*immunityChunk {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()
	return ic.chunks
}

// CountImmune returns the number of immunized (current or future) elements within the map
func (ic *ImmunityCache) CountImmune() int {
	count := 0
	for _, chunk := range ic.getChunksWithLock() {
		count += chunk.CountImmune()
	}
	return count
}

// NumBytes estimates the size of the cache, in bytes
func (ic *ImmunityCache) NumBytes() int {
	numBytes := 0
	for _, chunk := range ic.getChunksWithLock() {
		numBytes += chunk.NumBytes()
	}
	return numBytes
}

// Keys returns all keys
func (ic *ImmunityCache) Keys() [][]byte {
	count := ic.Count()
	keys := make([][]byte, 0, count)

	for _, chunk := range ic.getChunksWithLock() {
		keys = chunk.AppendKeys(keys)
	}

	return keys
}

// RegisterHandler is not implemented
func (ic *ImmunityCache) RegisterHandler(func(key []byte, value interface{}), string) {
	log.Error("ImmunityCache.RegisterHandler is not implemented")
}

// UnRegisterHandler removes the handler from the list
func (ic *ImmunityCache) UnRegisterHandler(_ string) {
	log.Error("ImmunityCache.UnRegisterHandler is not implemented")
}

// ForEachItem iterates over the items in the cache
func (ic *ImmunityCache) ForEachItem(function storage.ForEachItem) {
	for _, chunk := range ic.getChunksWithLock() {
		chunk.ForEachItem(function)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ic *ImmunityCache) IsInterfaceNil() bool {
	return ic == nil
}
