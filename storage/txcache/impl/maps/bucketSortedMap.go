package maps

import (
	"sync"
)

// BucketSortedMap is
type BucketSortedMap struct {
	mutex        sync.RWMutex
	nChunks      uint32
	nScoreChunks uint32
	maxScore     uint32
	chunks       []*MapChunk
	scoreChunks  []*MapChunk
}

// MapChunk is
type MapChunk struct {
	items map[string]BucketSortedMapItem
	mutex sync.RWMutex
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
	// Assignment is not an atomic operation, so we have to wrap this in a critical section
	sortedMap.mutex.Lock()
	defer sortedMap.mutex.Unlock()

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

// NotifyScoreChange moves or adds the item to the corresponding score chunk
func (sortedMap *BucketSortedMap) NotifyScoreChange(item BucketSortedMapItem, newScore uint32) {
	if newScore > sortedMap.maxScore {
		newScore = sortedMap.maxScore
	}

	newScoreChunk := sortedMap.getScoreChunks()[newScore]
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
	chunk.mutex.RLock()
	val, ok := chunk.items[key]
	chunk.mutex.RUnlock()
	return val, ok
}

// Has looks up an item under specified key
func (sortedMap *BucketSortedMap) Has(key string) bool {
	chunk := sortedMap.getChunk(key)
	chunk.mutex.RLock()
	_, ok := chunk.items[key]
	chunk.mutex.RUnlock()
	return ok
}

// Remove removes an element from the map
func (sortedMap *BucketSortedMap) Remove(key string) (interface{}, bool) {
	chunk := sortedMap.getChunk(key)
	item := chunk.removeItemByKey(key)
	if item != nil {
		removeFromScoreChunk(item)
	}

	return item, item != nil
}

// getChunk returns the chunk holding the given key.
func (sortedMap *BucketSortedMap) getChunk(key string) *MapChunk {
	sortedMap.mutex.RLock()
	defer sortedMap.mutex.RUnlock()
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
	// There is no need to explicitly remove each item for each chunk
	// The garbage collector will remove the data from memory
	sortedMap.initializeChunks()
}

// Count returns the number of elements within the map
func (sortedMap *BucketSortedMap) Count() uint32 {
	count := uint32(0)
	for _, chunk := range sortedMap.getChunks() {
		count += chunk.countItems()
	}
	return count
}

// CountSorted returns the number of sorted elements within the map
func (sortedMap *BucketSortedMap) CountSorted() uint32 {
	count := uint32(0)
	for _, chunk := range sortedMap.getScoreChunks() {
		count += chunk.countItems()
	}
	return count
}

// ChunksCounts returns the number of elements by chunk
func (sortedMap *BucketSortedMap) ChunksCounts() []uint32 {
	counts := make([]uint32, sortedMap.nChunks)
	for i, chunk := range sortedMap.getChunks() {
		counts[i] = chunk.countItems()
	}
	return counts
}

// ScoreChunksCounts returns the number of elements by chunk
func (sortedMap *BucketSortedMap) ScoreChunksCounts() []uint32 {
	counts := make([]uint32, sortedMap.nScoreChunks)
	for i, chunk := range sortedMap.getScoreChunks() {
		counts[i] = chunk.countItems()
	}
	return counts
}

// SortedMapIterCb is an iterator callback
type SortedMapIterCb func(key string, value BucketSortedMapItem)

// GetSnapshotAscending gets a snapshot of the items
func (sortedMap *BucketSortedMap) GetSnapshotAscending() []BucketSortedMapItem {
	return sortedMap.getSortedSnapshot(sortedMap.fillSnapshotAscending)
}

// GetSnapshotDescending gets a snapshot of the items
func (sortedMap *BucketSortedMap) GetSnapshotDescending() []BucketSortedMapItem {
	return sortedMap.getSortedSnapshot(sortedMap.fillSnapshotDescending)
}

// This applies a read lock on all chunks, so that they aren't mutated during snapshot
func (sortedMap *BucketSortedMap) getSortedSnapshot(fillSnapshot func(scoreChunks []*MapChunk, snapshot []BucketSortedMapItem)) []BucketSortedMapItem {
	counter := uint32(0)
	scoreChunks := sortedMap.getScoreChunks()

	for _, chunk := range scoreChunks {
		chunk.mutex.RLock()
		counter += uint32(len(chunk.items))
	}

	snapshot := make([]BucketSortedMapItem, counter)
	fillSnapshot(scoreChunks, snapshot)

	for _, chunk := range scoreChunks {
		chunk.mutex.RUnlock()
	}

	return snapshot
}

// This function should only be called under already read-locked score chunks
func (sortedMap *BucketSortedMap) fillSnapshotAscending(scoreChunks []*MapChunk, snapshot []BucketSortedMapItem) {
	i := 0
	for _, chunk := range scoreChunks {
		for _, item := range chunk.items {
			snapshot[i] = item
			i++
		}
	}
}

// This function should only be called under already read-locked score chunks
func (sortedMap *BucketSortedMap) fillSnapshotDescending(scoreChunks []*MapChunk, snapshot []BucketSortedMapItem) {
	i := 0
	for chunkIndex := len(scoreChunks) - 1; chunkIndex >= 0; chunkIndex-- {
		chunk := scoreChunks[chunkIndex]
		for _, item := range chunk.items {
			snapshot[i] = item
			i++
		}
	}
}

// IterCbSortedAscending iterates over the sorted elements in the map
func (sortedMap *BucketSortedMap) IterCbSortedAscending(callback SortedMapIterCb) {
	for _, chunk := range sortedMap.getScoreChunks() {
		chunk.forEachItem(callback)
	}
}

// IterCbSortedDescending iterates over the sorted elements in the map
func (sortedMap *BucketSortedMap) IterCbSortedDescending(callback SortedMapIterCb) {
	chunks := sortedMap.getScoreChunks()
	for i := len(chunks) - 1; i >= 0; i-- {
		chunk := chunks[i]
		chunk.forEachItem(callback)
	}
}

// Keys returns all keys as []string
func (sortedMap *BucketSortedMap) Keys() []string {
	count := sortedMap.Count()
	// count is not exact anymore, since we are in a different lock than the one aquired by Count() (but is a good approximation)
	keys := make([]string, 0, count)

	for _, chunk := range sortedMap.getChunks() {
		keys = chunk.appendKeys(keys)
	}

	return keys
}

// KeysSorted returns all keys of the sorted items as []string
func (sortedMap *BucketSortedMap) KeysSorted() []string {
	count := sortedMap.CountSorted()
	// count is not exact anymore, since we are in a different lock than the one aquired by CountSorted() (but is a good approximation)
	keys := make([]string, 0, count)

	for _, chunk := range sortedMap.getScoreChunks() {
		keys = chunk.appendKeys(keys)
	}

	return keys
}

func (sortedMap *BucketSortedMap) getChunks() []*MapChunk {
	sortedMap.mutex.RLock()
	defer sortedMap.mutex.RUnlock()
	return sortedMap.chunks
}

func (sortedMap *BucketSortedMap) getScoreChunks() []*MapChunk {
	sortedMap.mutex.RLock()
	defer sortedMap.mutex.RUnlock()
	return sortedMap.scoreChunks
}

func (chunk *MapChunk) removeItem(item BucketSortedMapItem) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	key := item.GetKey()
	delete(chunk.items, key)
}

func (chunk *MapChunk) removeItemByKey(key string) BucketSortedMapItem {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	item := chunk.items[key]
	delete(chunk.items, key)
	return item
}

func (chunk *MapChunk) setItem(item BucketSortedMapItem) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	key := item.GetKey()
	chunk.items[key] = item
}

func (chunk *MapChunk) countItems() uint32 {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()

	return uint32(len(chunk.items))
}

func (chunk *MapChunk) forEachItem(callback SortedMapIterCb) {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()

	for key, value := range chunk.items {
		callback(key, value)
	}
}

func (chunk *MapChunk) appendKeys(keysAccumulator []string) []string {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()

	for key := range chunk.items {
		keysAccumulator = append(keysAccumulator, key)
	}

	return keysAccumulator
}
