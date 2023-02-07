package ordering

import "sync"

type orderedCollection struct {
	itemsMap   map[string]int
	itemsArray [][]byte
	mut        sync.RWMutex
}

// NewOrderedCollection creates a new ordered collection
func NewOrderedCollection() *orderedCollection {
	return &orderedCollection{
		itemsMap:   make(map[string]int),
		itemsArray: make([][]byte, 0, 100),
		mut:        sync.RWMutex{},
	}
}

// Add adds a new item to the order collector
func (oc *orderedCollection) Add(item []byte) {
	oc.mut.Lock()
	defer oc.mut.Unlock()
	_, ok := oc.itemsMap[string(item)]
	if ok {
		return
	}
	oc.itemsMap[string(item)] = len(oc.itemsArray)
	oc.itemsArray = append(oc.itemsArray, item)
}

// GetItemAtIndex returns the item at the given index
func (oc *orderedCollection) GetItemAtIndex(index uint32) ([]byte, error) {
	oc.mut.RLock()
	defer oc.mut.RUnlock()

	if index >= uint32(len(oc.itemsArray)) {
		return nil, ErrIndexOutOfBounds
	}

	return oc.itemsArray[index], nil
}

// GetOrder returns the order of the item in the ordered collection
func (oc *orderedCollection) GetOrder(item []byte) (int, error) {
	oc.mut.RLock()
	defer oc.mut.RUnlock()
	order, ok := oc.itemsMap[string(item)]
	if !ok {
		return 0, ErrItemNotFound
	}

	return order, nil
}

// Remove removes an item from the order collector if it exists, adapting the order of the remaining items
func (oc *orderedCollection) Remove(item []byte) {
	oc.mut.Lock()
	defer oc.mut.Unlock()

	oc.removeOneUnprotected(item)
}

func (oc *orderedCollection) removeOneUnprotected(item []byte) {
	index, ok := oc.itemsMap[string(item)]
	if !ok {
		return
	}

	delete(oc.itemsMap, string(item))

	oc.itemsArray = append(oc.itemsArray[:index], oc.itemsArray[index+1:]...)
	for i := index; i < len(oc.itemsArray); i++ {
		oc.itemsMap[string(oc.itemsArray[i])]--
	}
}

// RemoveMultiple removes multiple items from the order collector if they exist, adapting the order of the remaining items
func (oc *orderedCollection) RemoveMultiple(items [][]byte) {
	oc.mut.Lock()
	defer oc.mut.Unlock()

	for _, item := range items {
		oc.removeOneUnprotected(item)
	}
}

// GetItems returns the items in the order they were added
func (oc *orderedCollection) GetItems() [][]byte {
	oc.mut.RLock()
	defer oc.mut.RUnlock()

	cpItems := make([][]byte, len(oc.itemsArray))
	copy(cpItems, oc.itemsArray)
	return cpItems
}

// Contains returns true if the item is in the ordered collection
func (oc *orderedCollection) Contains(item []byte) bool {
	oc.mut.RLock()
	defer oc.mut.RUnlock()
	_, ok := oc.itemsMap[string(item)]
	return ok
}

// Clear clears the ordered collection
func (oc *orderedCollection) Clear() {
	oc.mut.Lock()
	defer oc.mut.Unlock()
	oc.itemsArray = make([][]byte, 0, 100)
	oc.itemsMap = make(map[string]int)
}

// Len returns the number of items in the ordered collection
func (oc *orderedCollection) Len() int {
	oc.mut.RLock()
	defer oc.mut.RUnlock()
	return len(oc.itemsArray)
}

// IsInterfaceNil returns true if there is no value under the interface
func (oc *orderedCollection) IsInterfaceNil() bool {
	return oc == nil
}
