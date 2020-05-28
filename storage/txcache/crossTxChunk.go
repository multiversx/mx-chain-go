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
	return &crossTxChunk{
		config:      config,
		items:       make(map[string]crossTxChunkItem),
		itemsAsList: list.New(),
	}
}

func (chunk *crossTxChunk) immunizeKeys(keys []string) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	for _, key := range keys {
		item, ok := chunk.getItemNoLock(key)

		if ok {
			// Item exists, immunize on the spot
			item.ImmunizeAgainstEviction()
		} else {
			// Item not exists, will be immunized as it appears
			chunk.keysToImmunize[key] = emptyStruct
		}
	}
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
	log.Trace("crossTxChunk.doEviction()", "numRemoved", numRemoved, "err", err)
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
func (chunk *crossTxChunk) doEviction() (numRemoved uint32, err error) {
	if !chunk.isCapacityExceeded() {
		return 0, nil
	}

	numToRemoveEachStep := chunk.config.numItemsToPreemptivelyEvict

	// Do first step out of the loop
	numRemovedInStep := chunk.removeOldest(numToRemoveEachStep)
	numRemoved += numRemovedInStep

	if numRemovedInStep == 0 {
		return 0, errFailedCrossTxEviction
	}

	for !chunk.isCapacityExceeded() && numRemovedInStep == numToRemoveEachStep {
		numRemovedInStep = chunk.removeOldest(numToRemoveEachStep)
		numRemoved += numRemovedInStep
	}

	return numRemoved, nil
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

// This function should only be used in critical section (chunk.mutex)
func (chunk *crossTxChunk) removeOldest(numToRemove uint32) uint32 {
	numRemoved := uint32(0)
	for element := chunk.itemsAsList.Front(); element != nil; element = element.Next() {
		item := element.Value.(*WrappedTransaction)
		key := string(item.TxHash)

		if item.IsImmuneToEviction() {
			continue
		}

		// TODO: duplication
		delete(chunk.items, key)
		chunk.itemsAsList.Remove(element)
		chunk.trackNumBytesOnRemove(item)

		numRemoved++
		if numRemoved == numToRemove {
			break
		}
	}

	return numRemoved
}

func (chunk *crossTxChunk) countItems() uint32 {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return uint32(len(chunk.items))
}

func (chunk *crossTxChunk) appendKeys(keysAccumulator []string) []string {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()

	for key := range chunk.items {
		keysAccumulator = append(keysAccumulator, key)
	}

	return keysAccumulator
}
