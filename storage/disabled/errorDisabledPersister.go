package disabled

import (
	"fmt"
)

type errorDisabledPersister struct {
}

// NewErrorDisabledPersister returns a new instance of this disabled persister that errors on all operations
func NewErrorDisabledPersister() *errorDisabledPersister {
	return &errorDisabledPersister{}
}

// Put returns error
func (disabled *errorDisabledPersister) Put(_, _ []byte) error {
	return fmt.Errorf("disabledPersister.Put")
}

// Get returns error
func (disabled *errorDisabledPersister) Get(_ []byte) ([]byte, error) {
	return nil, fmt.Errorf("disabledPersister.Get")
}

// Has returns error
func (disabled *errorDisabledPersister) Has(_ []byte) error {
	return fmt.Errorf("disabledPersister.Has")
}

// Close returns error
func (disabled *errorDisabledPersister) Close() error {
	return fmt.Errorf("disabledPersister.Close")
}

// Remove returns error
func (disabled *errorDisabledPersister) Remove(_ []byte) error {
	return fmt.Errorf("disabledPersister.Remove")
}

// Destroy returns error
func (disabled *errorDisabledPersister) Destroy() error {
	return fmt.Errorf("disabledPersister.Destroy")
}

// DestroyClosed returns error
func (disabled *errorDisabledPersister) DestroyClosed() error {
	return fmt.Errorf("disabledPersister.DestroyClosed")
}

// RangeKeys does nothing
func (disabled *errorDisabledPersister) RangeKeys(_ func(key []byte, val []byte) bool) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (disabled *errorDisabledPersister) IsInterfaceNil() bool {
	return disabled == nil
}
