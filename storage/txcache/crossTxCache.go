package txcache

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ storage.Cacher = (*CrossTxCache)(nil)

// CrossTxCache holds cross-shard transactions (where destination == me)
type CrossTxCache struct {
	config ConfigDestinationMe
	chunks []*crossTxChunk
	mutex  sync.RWMutex
}

// NewCrossTxCache creates a new transactions cache
func NewCrossTxCache(config ConfigDestinationMe) (*CrossTxCache, error) {
	log.Debug("NewCrossTxCache", "config", config.String())

	err := config.verify()
	if err != nil {
		return nil, err
	}

	cache := CrossTxCache{
		config: config,
	}

	cache.initializeChunks()
	return &cache, nil
}

func (cache *CrossTxCache) initializeChunks() {
	// Assignment is not an atomic operation, so we have to wrap this in a critical section
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	config := cache.config
	chunkConfig := config.getChunkConfig()

	cache.chunks = make([]*crossTxChunk, config.NumChunks)
	for i := uint32(0); i < config.NumChunks; i++ {
		cache.chunks[i] = newCrossTxChunk(chunkConfig)
	}
}

// ImmunizeTxsAgainstEviction marks items as non-evictable
func (cache *CrossTxCache) ImmunizeTxsAgainstEviction(keys [][]byte) {
	numNow, numFuture := cache.doImmunizeTxsAgainstEviction(keys)
	log.Debug("CrossTxCache.ImmunizeTxsAgainstEviction()", "name", cache.config.Name, "len(keys)", len(keys), "numNow", numNow, "numFuture", numFuture)
	cache.diagnose()
}

func (cache *CrossTxCache) doImmunizeTxsAgainstEviction(keys [][]byte) (numNowTotal, numFutureTotal int) {
	groups := cache.groupKeysByChunk(keys)

	for chunkIndex, chunkKeys := range groups {
		chunk := cache.getChunkByIndex(chunkIndex)
		numNow, numFuture := chunk.ImmunizeKeys(chunkKeys)

		numNowTotal += numNow
		numFutureTotal += numFuture
	}

	return
}

func (cache *CrossTxCache) groupKeysByChunk(keys [][]byte) map[uint32][][]byte {
	groups := make(map[uint32][][]byte)

	for _, key := range keys {
		chunkIndex := cache.getChunkIndexByKey(string(key))
		groups[chunkIndex] = append(groups[chunkIndex], key)
	}

	return groups
}

func (cache *CrossTxCache) getChunkByKey(key string) *crossTxChunk {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	chunkIndex := cache.getChunkIndexByKey(key)

	return cache.chunks[chunkIndex]
}

func (cache *CrossTxCache) getChunkByIndex(index uint32) *crossTxChunk {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.chunks[index]
}

func (cache *CrossTxCache) getChunkIndexByKey(key string) uint32 {
	return fnv32Hash(key) % cache.config.NumChunks
}

func (cache *CrossTxCache) diagnose() {
	count := cache.Count()
	countImmunized := cache.CountImmunized()
	numBytes := cache.NumBytes()
	log.Debug("CrossTxCache.diagnose()", "name", cache.config.Name, "count", count, "countImmunized", countImmunized, "numBytes", numBytes)
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

// AddTx adds a transaction in the cache
func (cache *CrossTxCache) AddTx(tx *WrappedTransaction) (ok bool, added bool) {
	key := string(tx.TxHash)
	chunk := cache.getChunkByKey(key)
	return chunk.addItem(tx)
}

// Clear clears the map
func (cache *CrossTxCache) Clear() {
	// There is no need to explicitly remove each item for each chunk
	// The garbage collector will remove the data from memory
	cache.initializeChunks()
}

// MaxSize returns the capacity of the cache
func (cache *CrossTxCache) MaxSize() int {
	return int(cache.config.MaxNumItems)
}

// Len is an alias for Count
func (cache *CrossTxCache) Len() int {
	return cache.Count()
}

// Count returns the number of elements within the map
func (cache *CrossTxCache) Count() int {
	count := 0
	for _, chunk := range cache.getChunks() {
		count += chunk.CountItems()
	}
	return count
}

// CountImmunized returns the number of immunized (current or future) elements within the map
func (cache *CrossTxCache) CountImmunized() int {
	count := 0
	for _, chunk := range cache.getChunks() {
		count += chunk.CountItems()
	}
	return count
}

// NumBytes estimates the size of the cache, in bytes
func (cache *CrossTxCache) NumBytes() int {
	numBytes := 0
	for _, chunk := range cache.getChunks() {
		numBytes += chunk.NumBytes()
	}
	return numBytes
}

// Keys returns all keys
func (cache *CrossTxCache) Keys() [][]byte {
	count := cache.Count()
	// count is not exact anymore, since we are in a different lock than the one aquired by Count() (but is a good approximation)
	keys := make([][]byte, 0, count)

	for _, chunk := range cache.getChunks() {
		keys = chunk.AppendKeys(keys)
	}

	return keys
}

func (cache *CrossTxCache) getChunks() []*crossTxChunk {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.chunks
}

// Get gets a transaction by hash
func (cache *CrossTxCache) Get(key []byte) (value interface{}, ok bool) {
	tx, ok := cache.GetByTxHash(key)
	if ok {
		return tx.Tx, true
	}
	return nil, false
}

// GetByTxHash gets the transaction by hash
func (cache *CrossTxCache) GetByTxHash(txHash []byte) (*WrappedTransaction, bool) {
	chunk := cache.getChunkByKey(string(txHash))
	return chunk.getItem(string(txHash))
}

// Has checks is a transaction exists
func (cache *CrossTxCache) Has(key []byte) bool {
	chunk := cache.getChunkByKey(string(key))
	_, ok := chunk.getItem(string(key))
	return ok
}

// Peek gets a transaction by hash
func (cache *CrossTxCache) Peek(key []byte) (value interface{}, ok bool) {
	tx, ok := cache.GetByTxHash(key)
	if ok {
		return tx.Tx, true
	}
	return nil, false
}

// HasOrAdd is not implemented
func (cache *CrossTxCache) HasOrAdd(_ []byte, _ interface{}) (ok, evicted bool) {
	log.Error("CrossTxCache.HasOrAdd is not implemented")
	return false, false
}

// Put is not implemented
func (cache *CrossTxCache) Put(_ []byte, _ interface{}) (evicted bool) {
	log.Error("CrossTxCache.Put is not implemented")
	return false
}

// Remove removes tx by hash
func (cache *CrossTxCache) Remove(key []byte) {
	_ = cache.RemoveTxByHash(key)
}

// RemoveTxByHash removes tx by hash
func (cache *CrossTxCache) RemoveTxByHash(txHash []byte) bool {
	key := string(txHash)
	chunk := cache.getChunkByKey(key)
	return chunk.removeItem(key)
}

// RemoveOldest is not implemented
func (cache *CrossTxCache) RemoveOldest() {
	log.Error("CrossTxCache.RemoveOldest is not implemented")
}

// RegisterHandler is not implemented
func (cache *CrossTxCache) RegisterHandler(func(key []byte, value interface{})) {
	log.Error("CrossTxCache.RegisterHandler is not implemented")
}

// ForEachTransaction iterates over the transactions in the cache
func (cache *CrossTxCache) ForEachTransaction(function ForEachTransaction) {
	for _, chunk := range cache.getChunks() {
		chunk.ForEachTransaction(function)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *CrossTxCache) IsInterfaceNil() bool {
	return cache == nil
}
