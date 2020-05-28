package immunitycache

import (
	"container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

var emptyStruct struct{}

type immunityChunk struct {
	config               immunityChunkConfig
	items                map[string]immunityChunkItem
	itemsAsList          *list.List
	keysToImmunizeFuture map[string]struct{}
	numBytes             int
	mutex                sync.RWMutex
}

type immunityChunkItem struct {
	item        CacheItem
	listElement *list.Element
}

func newImmunityChunk(config immunityChunkConfig) *immunityChunk {
	log.Trace("newImmunityChunk", "config", config.String())

	return &immunityChunk{
		config:               config,
		items:                make(map[string]immunityChunkItem),
		itemsAsList:          list.New(),
		keysToImmunizeFuture: make(map[string]struct{}),
	}
}

func (chunk *immunityChunk) ImmunizeKeys(keys [][]byte) (numNow, numFuture int) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	for _, key := range keys {
		item, ok := chunk.getItemNoLock(string(key))

		if ok {
			// Item exists, immunize now!
			item.ImmunizeAgainstEviction()
			numNow++
		} else {
			// Item not yet in cache, will be immunized in the future
			chunk.keysToImmunizeFuture[string(key)] = emptyStruct
			numFuture++
		}
	}

	return
}

func (chunk *immunityChunk) getItemNoLock(key string) (CacheItem, bool) {
	item, ok := chunk.items[key]
	if !ok {
		return nil, false
	}

	return item.item, true
}

func (chunk *immunityChunk) addItemWithLock(item CacheItem) (ok bool, added bool) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	err := chunk.evictItemsIfCapacityExceededNoLock()
	if err != nil {
		return false, false
	}

	key := string(item.GetKey())

	if _, exists := chunk.items[key]; exists {
		return true, false
	}

	// First, we insert (append) in the linked list; then in the map
	// We also need to hold a reference to the list element, to have O(1) removal.
	element := chunk.itemsAsList.PushBack(item)
	chunk.items[key] = immunityChunkItem{item: item, listElement: element}

	// Immunize if appropriate
	_, shouldImmunize := chunk.keysToImmunizeFuture[key]
	if shouldImmunize {
		item.ImmunizeAgainstEviction()
	}

	chunk.trackNumBytesOnAddNoLock(item)
	return true, true
}

func (chunk *immunityChunk) evictItemsIfCapacityExceededNoLock() error {
	if !chunk.isCapacityExceededNoLock() {
		return nil
	}

	numRemoved, err := chunk.evictItemsNoLock()
	chunk.monitorEvictionNoLock(numRemoved, err)
	return err
}

func (chunk *immunityChunk) isCapacityExceededNoLock() bool {
	tooManyItems := len(chunk.items) >= int(chunk.config.maxNumItems)
	tooManyBytes := chunk.numBytes >= int(chunk.config.maxNumBytes)
	return tooManyItems || tooManyBytes
}

func (chunk *immunityChunk) evictItemsNoLock() (numRemoved int, err error) {
	numToRemoveEachStep := int(chunk.config.numItemsToPreemptivelyEvict)

	// We perform the first step out of the loop in order to detect & return error
	numRemovedInStep := chunk.removeOldestNoLock(numToRemoveEachStep)
	numRemoved += numRemovedInStep

	if numRemovedInStep == 0 {
		return 0, errFailedEviction
	}

	for chunk.isCapacityExceededNoLock() && numRemovedInStep == numToRemoveEachStep {
		numRemovedInStep = chunk.removeOldestNoLock(numToRemoveEachStep)
		numRemoved += numRemovedInStep
	}

	return numRemoved, nil
}

func (chunk *immunityChunk) removeOldestNoLock(numToRemove int) int {
	numRemoved := 0

	list := chunk.itemsAsList
	element := list.Front()

	for element != nil && numRemoved < numToRemove {
		item := element.Value.(CacheItem)
		key := string(item.GetKey())

		if item.IsImmuneToEviction() {
			element = element.Next()
			continue
		}

		// TODO: duplication
		elementToRemove := element
		element = element.Next()
		delete(chunk.items, key)
		list.Remove(elementToRemove)
		chunk.trackNumBytesOnRemoveNoLock(item)
		numRemoved++
	}

	return numRemoved
}

func (chunk *immunityChunk) monitorEvictionNoLock(numRemoved int, err error) {
	cacheName := chunk.config.cacheName

	if err != nil {
		log.Debug("immunityChunk.monitorEviction()", "name", cacheName, "numRemoved", numRemoved, "err", err)
	} else if numRemoved > 0 {
		log.Trace("immunityChunk.monitorEviction()", "name", cacheName, "numRemoved", numRemoved)
	}
}

func (chunk *immunityChunk) trackNumBytesOnAddNoLock(item CacheItem) {
	chunk.numBytes += item.Size()
}

func (chunk *immunityChunk) getItemWithLock(key string) (CacheItem, bool) {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return chunk.getItemNoLock(key)
}

func (chunk *immunityChunk) removeItemWithLock(key string) bool {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	item, ok := chunk.items[key]
	if !ok {
		return false
	}

	// TODO: duplication
	delete(chunk.items, key)
	delete(chunk.keysToImmunizeFuture, key)
	chunk.itemsAsList.Remove(item.listElement)
	chunk.trackNumBytesOnRemoveNoLock(item.item)
	return true
}

func (chunk *immunityChunk) trackNumBytesOnRemoveNoLock(item CacheItem) {
	chunk.numBytes -= item.Size()
	chunk.numBytes = core.MaxInt(chunk.numBytes, 0)
}

func (chunk *immunityChunk) RemoveOldest(numToRemove int) int {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()
	return chunk.removeOldestNoLock(numToRemove)
}

func (chunk *immunityChunk) CountItems() int {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return len(chunk.items)
}

func (chunk *immunityChunk) CountImmunized() int {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return len(chunk.keysToImmunizeFuture)
}

func (chunk *immunityChunk) NumBytes() int {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return chunk.numBytes
}

func (chunk *immunityChunk) KeysInOrder() [][]byte {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()

	keys := make([][]byte, 0, chunk.itemsAsList.Len())

	for element := chunk.itemsAsList.Front(); element != nil; element = element.Next() {
		item := element.Value.(CacheItem)
		keys = append(keys, item.GetKey())
	}

	return keys
}

func (chunk *immunityChunk) AppendKeys(keysAccumulator [][]byte) [][]byte {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()

	for key := range chunk.items {
		keysAccumulator = append(keysAccumulator, []byte(key))
	}

	return keysAccumulator
}

// ForEachItem iterates over the items in the chunk
func (chunk *immunityChunk) ForEachItem(function ForEachItem) {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()

	for key, value := range chunk.items {
		function([]byte(key), value.item)
	}
}
