package txcache

import (
	"sync"
)

var _ txCache = (*crossTxCache)(nil)

// crossTxCache holds cross-shard transactions (where destination == me)
type crossTxCache struct {
	config ConfigDestinationMe
	chunks []*crossTxChunk
	mutex  sync.RWMutex
}

// NewCrossTxCache creates a new transactions cache
func NewCrossTxCache(config ConfigDestinationMe) (*crossTxCache, error) {
	// TODO: Remove logic
	if config.NumChunks == 0 {
		config.NumChunks = 1
	}

	log.Debug("NewCrossTxCache", "config", config.String())

	cache := crossTxCache{
		config: config,
	}

	cache.initializeChunks()
	return &cache, nil
}

func (cache *crossTxCache) initializeChunks() {
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

func (cache *crossTxCache) immunizeKeys(keys []string) {
	groups := cache.groupKeysByChunk(keys)

	for chunkIndex, chunkKeys := range groups {
		chunk := cache.getChunkByIndex(chunkIndex)
		chunk.immunizeKeys(chunkKeys)
	}
}

// AddTx adds a transaction in the cache
func (cache *crossTxCache) AddTx(tx *WrappedTransaction) (ok bool, added bool) {
	return cache.AddItem(tx)
}

// AddItem adds the item in the map
func (cache *crossTxCache) AddItem(item *WrappedTransaction) (ok bool, added bool) {
	key := string(item.TxHash)
	chunk := cache.getChunkByKey(key)
	return chunk.addItem(item)
}

// Get gets an item from the map
func (cache *crossTxCache) GetItem(key string) (*WrappedTransaction, bool) {
	chunk := cache.getChunkByKey(key)
	return chunk.getItem(key)
}

// Has returns whether the item is in the map
func (cache *crossTxCache) HasItem(key string) bool {
	chunk := cache.getChunkByKey(key)
	_, ok := chunk.getItem(key)
	return ok
}

// Remove removes an element from the map
func (cache *crossTxCache) RemoveItem(key string) {
	chunk := cache.getChunkByKey(key)
	chunk.removeItem(key)
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

// Len is an alias for CountTx
func (cache *crossTxCache) Len() int {
	return int(cache.Count())
}

// Keys returns all keys
func (cache *crossTxCache) Keys() [][]byte {
	count := cache.Count()
	// count is not exact anymore, since we are in a different lock than the one aquired by Count() (but is a good approximation)
	keys := make([][]byte, 0, count)

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

// ForEachTransaction iterates over the transactions in the cache
func (cache *crossTxCache) ForEachTransaction(function ForEachTransaction) {
	chunks := cache.getChunks()

	for _, chunk := range chunks {
		// TODO: do not lock private mutex. Call chunk.IterCb()
		chunk.mutex.RLock()
		for key, value := range chunk.items {
			tx := value.payload
			function([]byte(key), tx)
		}
		chunk.mutex.RUnlock()
	}
}

// Get gets a transaction by hash
func (cache *crossTxCache) Get(key []byte) (value interface{}, ok bool) {
	tx, ok := cache.GetItem(string(key))
	if ok {
		return tx.Tx, true
	}
	return nil, false
}

// GetByTxHash gets the transaction by hash
func (cache *crossTxCache) GetByTxHash(txHash []byte) (*WrappedTransaction, bool) {
	return cache.GetItem(string(txHash))
}

// Has checks is a transaction exists
func (cache *crossTxCache) Has(key []byte) bool {
	return cache.HasItem(string(key))
}

// Peek gets a transaction by hash
func (cache *crossTxCache) Peek(key []byte) (value interface{}, ok bool) {
	tx, ok := cache.GetByTxHash(key)
	if ok {
		return tx.Tx, true
	}
	return nil, false
}

// HasOrAdd is not implemented
func (cache *crossTxCache) HasOrAdd(_ []byte, _ interface{}) (ok, evicted bool) {
	log.Error("crossTxCache.HasOrAdd is not implemented")
	return false, false
}

// ImmunizeTxsAgainstEviction does nothing for this type of cache
func (cache *crossTxCache) ImmunizeTxsAgainstEviction(keys [][]byte) {
	// TODO: implement
}

// MaxSize returns the capacity of the cache
func (cache *crossTxCache) MaxSize() int {
	return int(cache.config.MaxNumItems)
}

// Put is not implemented
func (cache *crossTxCache) Put(_ []byte, _ interface{}) (evicted bool) {
	log.Error("crossTxCache.Put is not implemented")
	return false
}

// RegisterHandler is not implemented
func (cache *crossTxCache) RegisterHandler(func(key []byte, value interface{})) {
	log.Error("crossTxCache.RegisterHandler is not implemented")
}

// Remove removes tx by hash
func (cache *crossTxCache) Remove(key []byte) {
	cache.RemoveItem(string(key))
}

// RemoveTxByHash removes tx by hash
func (cache *crossTxCache) RemoveTxByHash(txHash []byte) error {
	cache.RemoveItem(string(txHash))
	return nil
}

// RemoveOldest is not implemented
func (cache *crossTxCache) RemoveOldest() {
	log.Error("TxCache.RemoveOldest is not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *crossTxCache) IsInterfaceNil() bool {
	return cache == nil
}
