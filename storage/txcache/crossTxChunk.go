package txcache

import (
	"container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

var emptyStruct struct{}

type crossTxChunkItem struct {
	payload     *WrappedTransaction
	listElement *list.Element
}

// crossTx is a chunk of the crossTxCache
type crossTxChunk struct {
	config         crossTxChunkConfig
	items          map[string]crossTxChunkItem
	itemsAsList    *list.List
	keysToImmunize map[string]struct{}
	numBytes       int
	mutex          sync.RWMutex
}

func newCrossTxChunk(config crossTxChunkConfig) *crossTxChunk {
	log.Trace("newCrossTxChunk", "config", config.String())

	return &crossTxChunk{
		config:         config,
		items:          make(map[string]crossTxChunkItem),
		itemsAsList:    list.New(),
		keysToImmunize: make(map[string]struct{}),
	}
}

func (chunk *crossTxChunk) ImmunizeKeys(keys [][]byte) (numNow, numFuture int) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	for _, key := range keys {
		item, ok := chunk.getItemNoLock(string(key))

		if ok {
			// Item exists, immunize on the spot
			item.ImmunizeAgainstEviction()
			numNow++
		} else {
			// Item not exists, will be immunized as it appears
			chunk.keysToImmunize[string(key)] = emptyStruct
			numFuture++
		}
	}

	return
}

func (chunk *crossTxChunk) getItem(key string) (*WrappedTransaction, bool) {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return chunk.getItemNoLock(key)
}

// This function should only be used in critical section (chunk.mutex)
func (chunk *crossTxChunk) getItemNoLock(key string) (*WrappedTransaction, bool) {
	item, ok := chunk.items[key]
	if !ok {
		return nil, false
	}

	return item.payload, true
}

func (chunk *crossTxChunk) addItem(item *WrappedTransaction) (ok bool, added bool) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	numRemoved, err := chunk.doEviction()
	chunk.monitorEviction(numRemoved, err)
	if err != nil {
		return false, false
	}

	key := string(item.TxHash)

	if _, exists := chunk.items[key]; exists {
		return true, false
	}

	// First, we insert (append) in the linked list; then in the map
	// We also need to hold a reference to the list element, to have O(1) removal.
	element := chunk.itemsAsList.PushBack(item)
	chunk.items[key] = crossTxChunkItem{payload: item, listElement: element}

	// Immunize if appropriate
	_, shouldImmunize := chunk.keysToImmunize[key]
	if shouldImmunize {
		item.ImmunizeAgainstEviction()
	}

	chunk.trackNumBytesOnAdd(item)
	return true, true
}

// This function should only be used in critical section (chunk.mutex)
func (chunk *crossTxChunk) doEviction() (numRemoved int, err error) {
	if !chunk.isCapacityExceeded() {
		return 0, nil
	}

	numToRemoveEachStep := int(chunk.config.numItemsToPreemptivelyEvict)

	// We perform the first step out of the loop (to detect and return error)
	numRemovedInStep := chunk.removeOldestNoLock(numToRemoveEachStep)
	numRemoved += numRemovedInStep

	if numRemovedInStep == 0 {
		return 0, errFailedCrossTxEviction
	}

	for chunk.isCapacityExceeded() && numRemovedInStep == numToRemoveEachStep {
		numRemovedInStep = chunk.removeOldestNoLock(numToRemoveEachStep)
		numRemoved += numRemovedInStep
	}

	return numRemoved, nil
}

func (chunk *crossTxChunk) monitorEviction(numRemoved int, err error) {
	if err != nil {
		log.Debug("crossTxChunk.doEviction()", "numRemoved", numRemoved, "err", err)
	} else if numRemoved > 0 {
		log.Trace("crossTxChunk.doEviction()", "numRemoved", numRemoved)
	}
}

// This function should only be used in critical section (chunk.mutex)
func (chunk *crossTxChunk) trackNumBytesOnAdd(item *WrappedTransaction) {
	chunk.numBytes += int(estimateTxSize(item))
}

// This function should only be used in critical section (chunk.mutex)
func (chunk *crossTxChunk) trackNumBytesOnRemove(item *WrappedTransaction) {
	chunk.numBytes -= int(estimateTxSize(item))
	chunk.numBytes = core.MaxInt(chunk.numBytes, 0)
}

// This function should only be used in critical section (chunk.mutex)
func (chunk *crossTxChunk) isCapacityExceeded() bool {
	tooManyItems := len(chunk.items) >= int(chunk.config.maxNumItems)
	tooManyBytes := chunk.numBytes >= int(chunk.config.maxNumBytes)
	return tooManyItems || tooManyBytes
}

func (chunk *crossTxChunk) removeItem(key string) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	item, ok := chunk.items[key]
	if !ok {
		return
	}

	// TODO: duplication
	delete(chunk.items, key)
	delete(chunk.keysToImmunize, key)
	chunk.itemsAsList.Remove(item.listElement)
	chunk.trackNumBytesOnRemove(item.payload)
}

func (chunk *crossTxChunk) RemoveOldest(numToRemove int) int {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()
	return chunk.removeOldestNoLock(numToRemove)
}

// This function should only be used in critical section (chunk.mutex)
func (chunk *crossTxChunk) removeOldestNoLock(numToRemove int) int {
	numRemoved := 0

	list := chunk.itemsAsList
	element := list.Front()

	for element != nil && numRemoved < numToRemove {
		item := element.Value.(*WrappedTransaction)
		key := string(item.TxHash)

		if item.IsImmuneToEviction() {
			element = element.Next()
			continue
		}

		// TODO: duplication
		elementToRemove := element
		element = element.Next()
		delete(chunk.items, key)
		list.Remove(elementToRemove)
		chunk.trackNumBytesOnRemove(item)
		numRemoved++
	}

	return numRemoved
}

func (chunk *crossTxChunk) CountItems() int {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return len(chunk.items)
}

func (chunk *crossTxChunk) NumBytes() int {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return chunk.numBytes
}

func (chunk *crossTxChunk) KeysInOrder() [][]byte {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()

	keys := make([][]byte, 0, chunk.itemsAsList.Len())

	for element := chunk.itemsAsList.Front(); element != nil; element = element.Next() {
		value := element.Value.(*WrappedTransaction)
		keys = append(keys, value.TxHash)
	}

	return keys
}

func (chunk *crossTxChunk) AppendKeys(keysAccumulator [][]byte) [][]byte {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()

	for key := range chunk.items {
		keysAccumulator = append(keysAccumulator, []byte(key))
	}

	return keysAccumulator
}
