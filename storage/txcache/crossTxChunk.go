package txcache

import (
	"container/list"
	"sync"
)

type crossTxChunkItem struct {
	payload     *WrappedTransaction
	listElement *list.Element
}

// crossTx is a chunk of the crossTxCache
type crossTxChunk struct {
	items       map[string]crossTxChunkItem
	itemsAsList *list.List
	mutex       sync.RWMutex
}

func newCrossTxChunk() *crossTxChunk {
	return &crossTxChunk{
		items:       make(map[string]crossTxChunkItem),
		itemsAsList: list.New(),
	}
}

func (chunk *crossTxChunk) getItem(key string) (*WrappedTransaction, bool) {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()

	item, ok := chunk.items[key]
	if !ok {
		return nil, false
	}

	return item.payload, true
}

func (chunk *crossTxChunk) addItem(item *WrappedTransaction) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	key := string(item.TxHash)

	if _, ok := chunk.items[key]; ok {
		return
	}

	// First, we insert (append) in the linked list; then in the map
	// We also need to hold a reference to the list element, to have O(1) removal.
	element := chunk.itemsAsList.PushBack(item)
	chunk.items[key] = crossTxChunkItem{payload: item, listElement: element}
}

func (chunk *crossTxChunk) removeItem(key string) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	item, ok := chunk.items[key]
	if !ok {
		return
	}

	delete(chunk.items, key)
	chunk.itemsAsList.Remove(item.listElement)
}

func (chunk *crossTxChunk) removeOldest(numToRemove int) {
	chunk.mutex.Lock()
	defer chunk.mutex.Unlock()

	numRemoved := 0
	for element := chunk.itemsAsList.Front(); element != nil; element = element.Next() {
		item := element.Value.(*WrappedTransaction)
		key := string(item.TxHash)

		// TODO: if value.IsProtected(), skip.

		delete(chunk.items, key)
		chunk.itemsAsList.Remove(element)

		numRemoved++
		if numRemoved == numToRemove {
			break
		}
	}
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
