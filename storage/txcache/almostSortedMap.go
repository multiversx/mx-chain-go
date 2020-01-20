package txcache

import (
	"sync"
)

// AlmostSortedMap is
type AlmostSortedMap struct {
	globalMutex  sync.Mutex
	nChunks      uint32
	nScoreChunks uint32
	maxScore     uint32
	chunks       []*MapChunk
	scoreChunks  []*MapChunk
}

// MapItem is
type MapItem interface {
	GetKey() string
	ComputeScore() uint32
	GetScoreChunk() *MapChunk
	SetScoreChunk(*MapChunk)
}

// MapChunk is
type MapChunk struct {
	items map[string]MapItem
	sync.RWMutex
}

// NewAlmostSortedMap creates a new map.
func NewAlmostSortedMap(nChunks uint32, nScoreChunks uint32) *AlmostSortedMap {
	if nChunks == 0 {
		nChunks = 1
	}
	if nScoreChunks == 0 {
		nScoreChunks = 1
	}

	myMap := AlmostSortedMap{
		nChunks:      nChunks,
		nScoreChunks: nScoreChunks,
		maxScore:     nScoreChunks - 1,
		chunks:       make([]*MapChunk, nChunks),
		scoreChunks:  make([]*MapChunk, nScoreChunks),
	}

	for i := uint32(0); i < nChunks; i++ {
		myMap.chunks[i] = &MapChunk{
			items: make(map[string]MapItem),
		}
	}

	for i := uint32(0); i < nScoreChunks; i++ {
		myMap.scoreChunks[i] = &MapChunk{
			items: make(map[string]MapItem),
		}
	}

	return &myMap
}

// Set puts the item in the map
// This doesn't add the item to the score chunks (not necessary)
func (myMap *AlmostSortedMap) Set(item MapItem) {
	chunk := myMap.getChunk(item.GetKey())
	chunk.setItem(item)
}

// OnScoreChangeByKey moves or adds the item to the corresponding score chunk
func (myMap *AlmostSortedMap) OnScoreChangeByKey(key string) {
	item, ok := myMap.Get(key)
	if ok {
		myMap.OnScoreChange(item)
	}
}

// OnScoreChange moves or adds the item to the corresponding score chunk
func (myMap *AlmostSortedMap) OnScoreChange(item MapItem) {
	removeFromScoreChunk(item)

	newScore := item.ComputeScore()
	if newScore > myMap.maxScore {
		newScore = myMap.maxScore
	}

	newScoreChunk := myMap.scoreChunks[newScore]
	newScoreChunk.setItem(item)
}

func removeFromScoreChunk(item MapItem) {
	currentScoreChunk := item.GetScoreChunk()
	if currentScoreChunk != nil {
		currentScoreChunk.removeItem(item)
	}
}

// Get retrieves an element from map under given key.
func (myMap *AlmostSortedMap) Get(key string) (MapItem, bool) {
	chunk := myMap.getChunk(key)
	chunk.RLock()
	val, ok := chunk.items[key]
	chunk.RUnlock()
	return val, ok
}

// Count returns the number of elements within the map
func (myMap *AlmostSortedMap) Count() uint32 {
	count := uint32(0)
	for _, chunk := range myMap.chunks {
		count += chunk.countItems()
	}
	return count
}

// CountSorted returns the number of sorted elements within the map
func (myMap *AlmostSortedMap) CountSorted() uint32 {
	count := uint32(0)
	for _, chunk := range myMap.scoreChunks {
		count += chunk.countItems()
	}
	return count
}

// CountScoreChunk returns the number of sorted elements within a score chunk
func (myMap *AlmostSortedMap) CountScoreChunk(score uint32) uint32 {
	if score > myMap.maxScore {
		return 0
	}
	return myMap.scoreChunks[score].countItems()
}

// Has looks up an item under specified key
func (myMap *AlmostSortedMap) Has(key string) bool {
	chunk := myMap.getChunk(key)
	chunk.RLock()
	_, ok := chunk.items[key]
	chunk.RUnlock()
	return ok
}

// Remove removes an element from the map
func (myMap *AlmostSortedMap) Remove(key string) {
	chunk := myMap.getChunk(key)
	item := chunk.removeItemByKey(key)
	if item != nil {
		removeFromScoreChunk(item)
	}
}

// SortedMapIterCb is an iterator callback
type SortedMapIterCb func(key string, value MapItem)

// IterCb iterates over the elements in the map
func (myMap *AlmostSortedMap) IterCb(callback SortedMapIterCb) {
	for idx := range myMap.chunks {
		chunk := (myMap.chunks)[idx]
		chunk.RLock()
		for key, value := range chunk.items {
			callback(key, value)
		}
		chunk.RUnlock()
	}
}

// IterCbSortedAscending iterates over the sorted elements in the map
func (myMap *AlmostSortedMap) IterCbSortedAscending(callback SortedMapIterCb) {
	for _, chunk := range myMap.scoreChunks {
		chunk.RLock()
		for key, value := range chunk.items {
			callback(key, value)
		}
		chunk.RUnlock()
	}
}

// IterCbSortedDescending iterates over the sorted elements in the map
func (myMap *AlmostSortedMap) IterCbSortedDescending(callback SortedMapIterCb) {
	chunks := myMap.scoreChunks
	for i := len(chunks) - 1; i >= 0; i-- {
		chunk := chunks[i]
		chunk.RLock()
		for key, value := range chunk.items {
			callback(key, value)
		}
		chunk.RUnlock()
	}
}

// getChunk returns the chunk holding the given key.
func (myMap *AlmostSortedMap) getChunk(key string) *MapChunk {
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
func (myMap *AlmostSortedMap) Clear() {
	// There is no need to explicitly remove each item for each shard
	// The garbage collector will remove the data from memory

	// Assignment is not an atomic operation, so we have to wrap this in a critical section
	myMap.globalMutex.Lock()
	myMap.chunks = make([]*MapChunk, myMap.nChunks)
	myMap.scoreChunks = make([]*MapChunk, myMap.nScoreChunks)
	myMap.globalMutex.Unlock()
}

// Keys returns all keys as []string
func (myMap *AlmostSortedMap) Keys() []string {
	count := myMap.Count()
	chunks := myMap.chunks
	channel := make(chan string, count)

	go func() {
		// Foreach chunk.
		waitGroup := sync.WaitGroup{}
		waitGroup.Add(len(chunks))
		for _, chunk := range chunks {
			go func(chunk *MapChunk) {
				// Foreach key, value pair.
				chunk.RLock()
				for key := range chunk.items {
					channel <- key
				}
				chunk.RUnlock()
				waitGroup.Done()
			}(chunk)
		}
		waitGroup.Wait()
		close(channel)
	}()

	// Generate keys
	keys := make([]string, 0, count)
	for key := range channel {
		keys = append(keys, key)
	}
	return keys
}

// KeysSorted returns all keys of the sorted items as []string
func (myMap *AlmostSortedMap) KeysSorted() []string {
	count := myMap.CountSorted()
	keys := make([]string, 0, count)

	for _, chunk := range myMap.scoreChunks {
		chunk.RLock()
		for key := range chunk.items {
			keys = append(keys, key)
		}
		chunk.RUnlock()
	}

	return keys
}

func (chunk *MapChunk) removeItem(item MapItem) {
	chunk.Lock()
	defer chunk.Unlock()

	key := item.GetKey()
	delete(chunk.items, key)
}

func (chunk *MapChunk) removeItemByKey(key string) MapItem {
	chunk.Lock()
	defer chunk.Unlock()

	item, _ := chunk.items[key]
	delete(chunk.items, key)
	return item
}

func (chunk *MapChunk) setItem(item MapItem) {
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
