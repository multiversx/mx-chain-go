package mock

import (
	"github.com/ElrondNetwork/elrond-go/common"
)

// MockDB -
type MockDB struct {
}

// Put -
func (MockDB) Put(_, _ []byte, _ common.StorageAccessType) error {
	return nil
}

// Get -
func (MockDB) Get(_ []byte, _ common.StorageAccessType) ([]byte, error) {
	return []byte{}, nil
}

// Has -
func (MockDB) Has(_ []byte, _ common.StorageAccessType) error {
	return nil
}

// Close -
func (MockDB) Close() error {
	return nil
}

// Remove -
func (MockDB) Remove(_ []byte, _ common.StorageAccessType) error {
	return nil
}

// Destroy -
func (MockDB) Destroy() error {
	return nil
}

// DestroyClosed -
func (MockDB) DestroyClosed() error {
	return nil
}

// RangeKeys -
func (MockDB) RangeKeys(_ func(key []byte, val []byte) bool) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (s MockDB) IsInterfaceNil() bool {
	return false
}
