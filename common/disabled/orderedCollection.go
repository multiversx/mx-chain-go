package disabled

import (
	"errors"
)

type orderedCollection struct {
}

var errDisabledComponent = errors.New("disabled component")

// NewOrderedCollection creates a new ordered collection
func NewOrderedCollection() *orderedCollection {
	return &orderedCollection{}
}

// Add does nothing
func (oc *orderedCollection) Add(_ []byte) {}

// GetItemAtIndex does nothing
func (oc *orderedCollection) GetItemAtIndex(_ uint32) ([]byte, error) {
	return nil, errDisabledComponent
}

// GetOrder does nothing
func (oc *orderedCollection) GetOrder(_ []byte) (int, error) {
	return 0, errDisabledComponent
}

// Remove does nothing
func (oc *orderedCollection) Remove(_ []byte) {}

// RemoveMultiple does nothing
func (oc *orderedCollection) RemoveMultiple(_ [][]byte) {}

// GetItems does nothing
func (oc *orderedCollection) GetItems() [][]byte {
	return make([][]byte, 0)
}

// Contains does nothing
func (oc *orderedCollection) Contains(_ []byte) bool {
	return false
}

// Clear does nothing
func (oc *orderedCollection) Clear() {}

// Len does nothing
func (oc *orderedCollection) Len() int {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (oc *orderedCollection) IsInterfaceNil() bool {
	return oc == nil
}
