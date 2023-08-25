package disabled

import (
	"errors"
)

type orderedCollection struct {
}

var errDisabledComponent = errors.New("")

// NewOrderedCollection creates a new ordered collection
func NewOrderedCollection() *orderedCollection {
	return &orderedCollection{}
}

// Add adds a new item to the order collector
func (oc *orderedCollection) Add(_ []byte) {}

// GetItemAtIndex returns the item at the given index
func (oc *orderedCollection) GetItemAtIndex(_ uint32) ([]byte, error) {
	return nil, errDisabledComponent
}

// GetOrder returns the order of the item in the ordered collection
func (oc *orderedCollection) GetOrder(_ []byte) (int, error) {
	return 0, errDisabledComponent
}

// Remove removes an item from the order collector if it exists, adapting the order of the remaining items
func (oc *orderedCollection) Remove(_ []byte) {}

// RemoveMultiple removes multiple items from the order collector if they exist, adapting the order of the remaining items
func (oc *orderedCollection) RemoveMultiple(_ [][]byte) {}

// GetItems returns the items in the order they were added
func (oc *orderedCollection) GetItems() [][]byte {
	return make([][]byte, 0)
}

// Contains returns true if the item is in the ordered collection
func (oc *orderedCollection) Contains(_ []byte) bool {
	return false
}

// Clear clears the ordered collection
func (oc *orderedCollection) Clear() {}

// Len returns the number of items in the ordered collection
func (oc *orderedCollection) Len() int {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (oc *orderedCollection) IsInterfaceNil() bool {
	return oc == nil
}
