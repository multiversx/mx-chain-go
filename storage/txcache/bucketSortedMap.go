package txcache

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

// ScoredItem is
type ScoredItem interface {
	GetKey() string
	ComputeScore() uint32
	GetScoreChunk() *MapChunk
	SetScoreChunk(*MapChunk)
}

// MapChunk is
type MapChunk struct {
	items map[string]ScoredItem
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

	myMap := BucketSortedMap{
		nChunks:      nChunks,
		nScoreChunks: nScoreChunks,
		maxScore:     nScoreChunks - 1,
		chunks:       make([]*MapChunk, nChunks),
		scoreChunks:  make([]*MapChunk, nScoreChunks),
	}

	for i := uint32(0); i < nChunks; i++ {
		myMap.chunks[i] = &MapChunk{
			items: make(map[string]ScoredItem),
		}
	}

	for i := uint32(0); i < nScoreChunks; i++ {
		myMap.scoreChunks[i] = &MapChunk{
			items: make(map[string]ScoredItem),
		}
	}

	return &myMap
}

// Set puts the item in the map
// This doesn't add the item to the score chunks (not necessary)
func (myMap *BucketSortedMap) Set(item ScoredItem) {
	chunk := myMap.getChunk(item.GetKey())
	chunk.setItem(item)
}

// OnScoreChangeByKey moves or adds the item to the corresponding score chunk
func (myMap *BucketSortedMap) OnScoreChangeByKey(key string) {
	item, ok := myMap.Get(key)
	if ok {
		myMap.OnScoreChange(item)
	}
}

// OnScoreChange moves or adds the item to the corresponding score chunk
func (myMap *BucketSortedMap) OnScoreChange(item ScoredItem) {
	newScore := item.ComputeScore()
	if newScore > myMap.maxScore {
		newScore = myMap.maxScore
	}

	newScoreChunk := myMap.scoreChunks[newScore]
	if newScoreChunk != item.GetScoreChunk() {
		removeFromScoreChunk(item)
		newScoreChunk.setItem(item)
		item.SetScoreChunk(newScoreChunk)
	}
}

func removeFromScoreChunk(item ScoredItem) {
	currentScoreChunk := item.GetScoreChunk()
	if currentScoreChunk != nil {
		currentScoreChunk.removeItem(item)
	}
}

// Get retrieves an element from map under given key.
func (myMap *BucketSortedMap) Get(key string) (ScoredItem, bool) {
	chunk := myMap.getChunk(key)
	chunk.RLock()
	val, ok := chunk.items[key]
	chunk.RUnlock()
	return val, ok
}

// Count returns the number of elements within the map
func (myMap *BucketSortedMap) Count() uint32 {
	count := uint32(0)
	for _, chunk := range myMap.chunks {
		count += chunk.countItems()
	}
	return count
}

// CountSorted returns the number of sorted elements within the map
func (myMap *BucketSortedMap) CountSorted() uint32 {
	count := uint32(0)
	for _, chunk := range myMap.scoreChunks {
		count += chunk.countItems()
	}
	return count
}

// ChunksCounts returns the number of elements by chunk
func (myMap *BucketSortedMap) ChunksCounts() []uint32 {
	counts := make([]uint32, myMap.nChunks)
	for i, chunk := range myMap.chunks {
		counts[i] = chunk.countItems()
	}
	return counts
}

// ScoreChunksCounts returns the number of elements by chunk
func (myMap *BucketSortedMap) ScoreChunksCounts() []uint32 {
	counts := make([]uint32, myMap.nScoreChunks)
	for i, chunk := range myMap.scoreChunks {
		counts[i] = chunk.countItems()
	}
	return counts
}

// Has looks up an item under specified key
func (myMap *BucketSortedMap) Has(key string) bool {
	chunk := myMap.getChunk(key)
	chunk.RLock()
	_, ok := chunk.items[key]
	chunk.RUnlock()
	return ok
}

// Remove removes an element from the map
func (myMap *BucketSortedMap) Remove(key string) {
	chunk := myMap.getChunk(key)
	item := chunk.removeItemByKey(key)
	if item != nil {
		removeFromScoreChunk(item)
	}
}

// SortedMapIterCb is an iterator callback
type SortedMapIterCb func(key string, value ScoredItem)

// IterCb iterates over the elements in the map
func (myMap *BucketSortedMap) IterCb(callback SortedMapIterCb) {
	for idx := range myMap.chunks {
		chunk := (myMap.chunks)[idx]
		chunk.forEachItem(callback)
	}
}

// IterCbSortedAscending iterates over the sorted elements in the map
func (myMap *BucketSortedMap) IterCbSortedAscending(callback SortedMapIterCb) {
	for _, chunk := range myMap.scoreChunks {
		chunk.forEachItem(callback)
	}
}

// GetSnapshotAscending gets a snapshot of the items
// This applies a read lock on all chunks, so that they aren't mutated during snapshot
func (myMap *BucketSortedMap) GetSnapshotAscending() []ScoredItem {
	counter := uint32(0)

	for _, chunk := range myMap.scoreChunks {
		chunk.RLock()
		counter += uint32(len(chunk.items))
	}

	if counter == 0 {
		return make([]ScoredItem, 0)
	}

	snapshot := make([]ScoredItem, 0, counter)

	for _, chunk := range myMap.scoreChunks {
		for _, item := range chunk.items {
			snapshot = append(snapshot, item)
		}
	}

	for _, chunk := range myMap.scoreChunks {
		chunk.RUnlock()
	}

	return snapshot
}

// IterCbSortedDescending iterates over the sorted elements in the map
func (myMap *BucketSortedMap) IterCbSortedDescending(callback SortedMapIterCb) {
	chunks := myMap.scoreChunks
	for i := len(chunks) - 1; i >= 0; i-- {
		chunk := chunks[i]
		chunk.forEachItem(callback)
	}
}

// getChunk returns the chunk holding the given key.
func (myMap *BucketSortedMap) getChunk(key string) *MapChunk {
	return myMap.chunks[fnv32Hash(key)%myMap.nChunks]
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
func (myMap *BucketSortedMap) Clear() {
	// There is no need to explicitly remove each item for each shard
	// The garbage collector will remove the data from memory

	// Assignment is not an atomic operation, so we have to wrap this in a critical section
	myMap.globalMutex.Lock()
	myMap.chunks = make([]*MapChunk, myMap.nChunks)
	myMap.scoreChunks = make([]*MapChunk, myMap.nScoreChunks)
	myMap.globalMutex.Unlock()
}

// Keys returns all keys as []string
func (myMap *BucketSortedMap) Keys() []string {
	count := myMap.Count()
	keys := make([]string, 0, count)

	for _, chunk := range myMap.chunks {
		keys = chunk.appendKeys(keys)
	}

	return keys
}

// KeysSorted returns all keys of the sorted items as []string
func (myMap *BucketSortedMap) KeysSorted() []string {
	count := myMap.CountSorted()
	keys := make([]string, 0, count)

	for _, chunk := range myMap.scoreChunks {
		keys = chunk.appendKeys(keys)
	}

	return keys
}

func (chunk *MapChunk) removeItem(item ScoredItem) {
	chunk.Lock()
	defer chunk.Unlock()

	key := item.GetKey()
	delete(chunk.items, key)
}

func (chunk *MapChunk) removeItemByKey(key string) ScoredItem {
	chunk.Lock()
	defer chunk.Unlock()

	item := chunk.items[key]
	delete(chunk.items, key)
	return item
}

func (chunk *MapChunk) setItem(item ScoredItem) {
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
