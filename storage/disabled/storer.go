package disabled

import (
	storageCore "github.com/multiversx/mx-chain-core-go/storage"
	"github.com/multiversx/mx-chain-storage-go/common"
)

type storer struct{}

// NewStorer returns a new instance of this disabled storer
func NewStorer() *storer {
	return &storer{}
}

// Put returns nil
func (s *storer) Put(_, _ []byte) error {
	return nil
}

// PutInEpoch returns nil
func (s *storer) PutInEpoch(_, _ []byte, _ uint32) error {
	return nil
}

// Get returns nil
func (s *storer) Get(_ []byte) ([]byte, error) {
	return nil, nil
}

// Has returns nil
func (s *storer) Has(_ []byte) error {
	return nil
}

// SearchFirst returns nil
func (s *storer) SearchFirst(_ []byte) ([]byte, error) {
	return nil, nil
}

// RemoveFromCurrentEpoch returns nil
func (s *storer) RemoveFromCurrentEpoch(_ []byte) error {
	return nil
}

// Remove returns nil
func (s *storer) Remove(_ []byte) error {
	return nil
}

// ClearCache does nothing
func (s *storer) ClearCache() {
}

// DestroyUnit returns nil
func (s *storer) DestroyUnit() error {
	return nil
}

// GetFromEpoch returns nil
func (s *storer) GetFromEpoch(_ []byte, _ uint32) ([]byte, error) {
	return nil, nil
}

// GetBulkFromEpoch returns nil
func (s *storer) GetBulkFromEpoch(_ [][]byte, _ uint32) ([]storageCore.KeyValuePair, error) {
	return nil, nil
}

// GetOldestEpoch returns 0
func (s *storer) GetOldestEpoch() (uint32, error) {
	return 0, common.ErrOldestEpochNotAvailable
}

// RangeKeys does nothing
func (s *storer) RangeKeys(_ func(key []byte, val []byte) bool) {
}

// Close returns nil
func (s *storer) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *storer) IsInterfaceNil() bool {
	return s == nil
}
