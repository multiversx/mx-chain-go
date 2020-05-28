package immunitycache

import (
	"container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

var emptyStruct struct{}

type immunityChunk struct {
	config               immunityChunkConfig
	items                map[string]chunkItemWrapper
	itemsAsList          *list.List
	keysToImmunizeFuture map[string]struct{}
	numBytes             int
	mutex                sync.RWMutex
}

type chunkItemWrapper struct {
	item        CacheItem
	listElement *list.Element
}

func newImmunityChunk(config immunityChunkConfig) *immunityChunk {
	log.Trace("newImmunityChunk", "config", config.String())

	return &immunityChunk{
		config:               config,
		items:                make(map[string]chunkItemWrapper),
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
	wrapper, ok := chunk.items[key]
	if !ok {
		return nil, false
	}

	return wrapper.item, true
}

func (chunk *immunityChunk) AddItem(item CacheItem) (ok bool, added bool) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	err := chunk.evictItemsIfCapacityExceededNoLock()
	if err != nil {
		// No more room for the new item
		return false, false
	}

	// Discard duplicates
	if chunk.itemExistsNoLock(item) {
		return true, false
	}

	chunk.addItemNoLock(item)
	chunk.immunizeItemOnAddNoLock(item)
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
	element := chunk.itemsAsList.Front()

	for element != nil && numRemoved < numToRemove {
		item := element.Value.(CacheItem)

		if item.IsImmuneToEviction() {
			element = element.Next()
			continue
		}

		elementToRemove := element
		element = element.Next()

		chunk.removeNoLock(elementToRemove)
		numRemoved++
	}

	return numRemoved
}

func (chunk *immunityChunk) removeNoLock(element *list.Element) {
	item := element.Value.(CacheItem)
	delete(chunk.items, string(item.GetKey()))
	chunk.itemsAsList.Remove(element)
	chunk.trackNumBytesOnRemoveNoLock(item)
}

func (chunk *immunityChunk) monitorEvictionNoLock(numRemoved int, err error) {
	cacheName := chunk.config.cacheName

	if err != nil {
		log.Debug("immunityChunk.monitorEviction()", "name", cacheName, "numRemoved", numRemoved, "err", err)
	} else if numRemoved > 0 {
		log.Trace("immunityChunk.monitorEviction()", "name", cacheName, "numRemoved", numRemoved)
	}
}

func (chunk *immunityChunk) itemExistsNoLock(item CacheItem) bool {
	key := string(item.GetKey())
	_, exists := chunk.items[key]
	return exists
}

func (chunk *immunityChunk) addItemNoLock(item CacheItem) {
	key := string(item.GetKey())

	// First, we insert (append) in the linked list; then in the map.
	// In the map, we also need to hold a reference to the list element, to have O(1) removal.
	element := chunk.itemsAsList.PushBack(item)
	chunk.items[key] = chunkItemWrapper{item: item, listElement: element}
}

func (chunk *immunityChunk) immunizeItemOnAddNoLock(item CacheItem) {
	key := string(item.GetKey())

	if _, immunize := chunk.keysToImmunizeFuture[key]; immunize {
		item.ImmunizeAgainstEviction()
		delete(chunk.keysToImmunizeFuture, key)
	}
}

func (chunk *immunityChunk) trackNumBytesOnAddNoLock(item CacheItem) {
	chunk.numBytes += item.Size()
}

func (chunk *immunityChunk) GetItem(key string) (CacheItem, bool) {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return chunk.getItemNoLock(key)
}

func (chunk *immunityChunk) RemoveItem(key string) bool {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	// In order to improve the robustness of the cache, we'll also remove from "keysToImmunizeFuture",
	// even if the item does not actually exist in the cache - to allow un-doing immunization intent (perhaps useful for rollbacks).
	delete(chunk.keysToImmunizeFuture, key)

	wrapper, ok := chunk.items[key]
	if !ok {
		return false
	}

	chunk.removeNoLock(wrapper.listElement)
	return true
}

func (chunk *immunityChunk) trackNumBytesOnRemoveNoLock(item CacheItem) {
	chunk.numBytes -= item.Size()
	chunk.numBytes = core.MaxInt(chunk.numBytes, 0)
}

// RemoveOldest removes a number of old items
func (chunk *immunityChunk) RemoveOldest(numToRemove int) int {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()
	return chunk.removeOldestNoLock(numToRemove)
}

// Count counts the items
func (chunk *immunityChunk) Count() int {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return len(chunk.items)
}

// CountImmune counts the immune items
func (chunk *immunityChunk) CountImmune() int {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return len(chunk.keysToImmunizeFuture)
}

// NumBytes gets the number of bytes stored
func (chunk *immunityChunk) NumBytes() int {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return chunk.numBytes
}

// KeysInOrder gets the keys, in order
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

// AppendKeys accumulates keys in a given slice
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
