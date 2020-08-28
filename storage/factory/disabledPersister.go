package factory

import (
	"fmt"
)

type disabledPersister struct {
}

// Put returns error
func (dp *disabledPersister) Put(key, val []byte) error {
	return fmt.Errorf("disabledPersister.Put")
}

// Get returns error
func (dp *disabledPersister) Get(key []byte) ([]byte, error) {
	return nil, fmt.Errorf("disabledPersister.Get")
}

// Has returns error
func (dp *disabledPersister) Has(key []byte) error {
	return fmt.Errorf("disabledPersister.Has")
}

// Init returns error
func (dp *disabledPersister) Init() error {
	return fmt.Errorf("disabledPersister.Init")
}

// Close returns error
func (dp *disabledPersister) Close() error {
	return fmt.Errorf("disabledPersister.Close")
}

// Remove returns error
func (dp *disabledPersister) Remove(key []byte) error {
	return fmt.Errorf("disabledPersister.Remove")
}

// Destroy does nothing
func (dp *disabledPersister) Destroy() error {
	return fmt.Errorf("disabledPersister.Destroy")
}

// DestroyClosed returns error
func (dp *disabledPersister) DestroyClosed() error {
	return fmt.Errorf("disabledPersister.DestroyClosed")
}

// RangeKeys does nothing
func (dp *disabledPersister) RangeKeys(handler func(key []byte, val []byte) bool) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (dp *disabledPersister) IsInterfaceNil() bool {
	return dp == nil
}
