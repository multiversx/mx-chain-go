package maps

import (
	"sync"
)

// BucketSortedMap is
type BucketSortedMap struct {
	globalMutex  sync.Mutex
	nChunks      uint32
	nScoreChunks uint32
	maxScore     uint32
	chunks       []*MapChunk
	scoreChunks  []*MapChunk
}

// MapChunk is
type MapChunk struct {
	items map[string]BucketSortedMapItem
	sync.RWMutex
}

// NewBucketSortedMap creates a new map.
func NewBucketSortedMap(nChunks uint32, nScoreChunks uint32) *BucketSortedMap {
	if nChunks == 0 {
		nChunks = 1
	}
	if nScoreChunks == 0 {
		nScoreChunks = 1
	}

	sortedMap := BucketSortedMap{
		nChunks:      nChunks,
		nScoreChunks: nScoreChunks,
		maxScore:     nScoreChunks - 1,
	}

	sortedMap.initializeChunks()

	return &sortedMap
}

func (sortedMap *BucketSortedMap) initializeChunks() {
	sortedMap.chunks = make([]*MapChunk, sortedMap.nChunks)
	sortedMap.scoreChunks = make([]*MapChunk, sortedMap.nScoreChunks)

	for i := uint32(0); i < sortedMap.nChunks; i++ {
		sortedMap.chunks[i] = &MapChunk{
			items: make(map[string]BucketSortedMapItem),
		}
	}

	for i := uint32(0); i < sortedMap.nScoreChunks; i++ {
		sortedMap.scoreChunks[i] = &MapChunk{
			items: make(map[string]BucketSortedMapItem),
		}
	}
}

// Set puts the item in the map
// This doesn't add the item to the score chunks (not necessary)
func (sortedMap *BucketSortedMap) Set(item BucketSortedMapItem) {
	chunk := sortedMap.getChunk(item.GetKey())
	chunk.setItem(item)
}

// OnScoreChangeByKey moves or adds the item to the corresponding score chunk
func (sortedMap *BucketSortedMap) OnScoreChangeByKey(key string) {
	item, ok := sortedMap.Get(key)
	if ok {
		sortedMap.OnScoreChange(item)
	}
}

// OnScoreChange moves or adds the item to the corresponding score chunk
func (sortedMap *BucketSortedMap) OnScoreChange(item BucketSortedMapItem) {
	newScore := item.ComputeScore()
	if newScore > sortedMap.maxScore {
		newScore = sortedMap.maxScore
	}

	newScoreChunk := sortedMap.scoreChunks[newScore]
	if newScoreChunk != item.GetScoreChunk() {
		removeFromScoreChunk(item)
		newScoreChunk.setItem(item)
		item.SetScoreChunk(newScoreChunk)
	}
}

func removeFromScoreChunk(item BucketSortedMapItem) {
	currentScoreChunk := item.GetScoreChunk()
	if currentScoreChunk != nil {
		currentScoreChunk.removeItem(item)
	}
}

// Get retrieves an element from map under given key.
func (sortedMap *BucketSortedMap) Get(key string) (BucketSortedMapItem, bool) {
	chunk := sortedMap.getChunk(key)
	chunk.RLock()
	val, ok := chunk.items[key]
	chunk.RUnlock()
	return val, ok
}

// Count returns the number of elements within the map
func (sortedMap *BucketSortedMap) Count() uint32 {
	count := uint32(0)
	for _, chunk := range sortedMap.chunks {
		count += chunk.countItems()
	}
	return count
}

// CountSorted returns the number of sorted elements within the map
func (sortedMap *BucketSortedMap) CountSorted() uint32 {
	count := uint32(0)
	for _, chunk := range sortedMap.scoreChunks {
		count += chunk.countItems()
	}
	return count
}

// ChunksCounts returns the number of elements by chunk
func (sortedMap *BucketSortedMap) ChunksCounts() []uint32 {
	counts := make([]uint32, sortedMap.nChunks)
	for i, chunk := range sortedMap.chunks {
		counts[i] = chunk.countItems()
	}
	return counts
}

// ScoreChunksCounts returns the number of elements by chunk
func (sortedMap *BucketSortedMap) ScoreChunksCounts() []uint32 {
	counts := make([]uint32, sortedMap.nScoreChunks)
	for i, chunk := range sortedMap.scoreChunks {
		counts[i] = chunk.countItems()
	}
	return counts
}

// Has looks up an item under specified key
func (sortedMap *BucketSortedMap) Has(key string) bool {
	chunk := sortedMap.getChunk(key)
	chunk.RLock()
	_, ok := chunk.items[key]
	chunk.RUnlock()
	return ok
}

// Remove removes an element from the map
func (sortedMap *BucketSortedMap) Remove(key string) {
	chunk := sortedMap.getChunk(key)
	item := chunk.removeItemByKey(key)
	if item != nil {
		removeFromScoreChunk(item)
	}
}

// SortedMapIterCb is an iterator callback
type SortedMapIterCb func(key string, value BucketSortedMapItem)

// GetSnapshotAscending gets a snapshot of the items
// This applies a read lock on all chunks, so that they aren't mutated during snapshot
func (sortedMap *BucketSortedMap) GetSnapshotAscending() []BucketSortedMapItem {
	counter := uint32(0)

	for _, chunk := range sortedMap.scoreChunks {
		chunk.RLock()
		counter += uint32(len(chunk.items))
	}

	if counter == 0 {
		return make([]BucketSortedMapItem, 0)
	}

	snapshot := make([]BucketSortedMapItem, 0, counter)

	for _, chunk := range sortedMap.scoreChunks {
		for _, item := range chunk.items {
			snapshot = append(snapshot, item)
		}
	}

	for _, chunk := range sortedMap.scoreChunks {
		chunk.RUnlock()
	}

	return snapshot
}

// IterCbSortedDescending iterates over the sorted elements in the map
func (sortedMap *BucketSortedMap) IterCbSortedDescending(callback SortedMapIterCb) {
	chunks := sortedMap.scoreChunks
	for i := len(chunks) - 1; i >= 0; i-- {
		chunk := chunks[i]
		chunk.forEachItem(callback)
	}
}

// getChunk returns the chunk holding the given key.
func (sortedMap *BucketSortedMap) getChunk(key string) *MapChunk {
	return sortedMap.chunks[fnv32Hash(key)%sortedMap.nChunks]
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
func (sortedMap *BucketSortedMap) Clear() {
	// There is no need to explicitly remove each item for each shard
	// The garbage collector will remove the data from memory

	// Assignment is not an atomic operation, so we have to wrap this in a critical section
	sortedMap.globalMutex.Lock()
	sortedMap.initializeChunks()
	sortedMap.globalMutex.Unlock()
}

// Keys returns all keys as []string
func (sortedMap *BucketSortedMap) Keys() []string {
	count := sortedMap.Count()
	keys := make([]string, 0, count)

	for _, chunk := range sortedMap.chunks {
		keys = chunk.appendKeys(keys)
	}

	return keys
}

// KeysSorted returns all keys of the sorted items as []string
func (sortedMap *BucketSortedMap) KeysSorted() []string {
	count := sortedMap.CountSorted()
	keys := make([]string, 0, count)

	for _, chunk := range sortedMap.scoreChunks {
		keys = chunk.appendKeys(keys)
	}

	return keys
}

func (chunk *MapChunk) removeItem(item BucketSortedMapItem) {
	chunk.Lock()
	defer chunk.Unlock()

	key := item.GetKey()
	delete(chunk.items, key)
}

func (chunk *MapChunk) removeItemByKey(key string) BucketSortedMapItem {
	chunk.Lock()
	defer chunk.Unlock()

	item := chunk.items[key]
	delete(chunk.items, key)
	return item
}

func (chunk *MapChunk) setItem(item BucketSortedMapItem) {
	chunk.Lock()
	defer chunk.Unlock()

	key := item.GetKey()
	chunk.items[key] = item
}

func (chunk *MapChunk) countItems() uint32 {
	chunk.RLock()
	defer chunk.RUnlock()

	return uint32(len(chunk.items))
}

func (chunk *MapChunk) forEachItem(callback SortedMapIterCb) {
	chunk.RLock()
	defer chunk.RUnlock()

	for key, value := range chunk.items {
		callback(key, value)
	}
}

func (chunk *MapChunk) appendKeys(keysAccumulator []string) []string {
	chunk.RLock()
	defer chunk.RUnlock()

	for key := range chunk.items {
		keysAccumulator = append(keysAccumulator, key)
	}

	return keysAccumulator
}
