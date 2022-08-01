package mock

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/storage"
)

// StorerMock -
type StorerMock struct {
	mut  sync.Mutex
	data map[string][]byte
}

// NewStorerMock -
func NewStorerMock() *StorerMock {
	return &StorerMock{
		data: make(map[string][]byte),
	}
}

// Close -
func (sm *StorerMock) Close() error {
	return nil
}

// Put -
func (sm *StorerMock) Put(key, data []byte) error {
	sm.mut.Lock()
	defer sm.mut.Unlock()
	sm.data[string(key)] = data

	return nil
}

// PutInEpoch -
func (sm *StorerMock) PutInEpoch(key, data []byte, _ uint32) error {
	return sm.Put(key, data)
}

// Get -
func (sm *StorerMock) Get(key []byte) ([]byte, error) {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	val, ok := sm.data[string(key)]
	if !ok {
		return nil, fmt.Errorf("key: %s not found", base64.StdEncoding.EncodeToString(key))
	}

	return val, nil
}

// GetFromEpoch -
func (sm *StorerMock) GetFromEpoch(key []byte, _ uint32) ([]byte, error) {
	return sm.Get(key)
}

// GetBulkFromEpoch -
func (sm *StorerMock) GetBulkFromEpoch(keys [][]byte, _ uint32) ([]storage.KeyValuePair, error) {
	return nil, errors.New("not implemented")
}

// SearchFirst -
func (sm *StorerMock) SearchFirst(_ []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Has -
func (sm *StorerMock) Has(_ []byte) error {
	return errors.New("not implemented")
}

// RemoveFromCurrentEpoch -
func (sm *StorerMock) RemoveFromCurrentEpoch(key []byte) error {
	return sm.Remove(key)
}

// Remove -
func (sm *StorerMock) Remove(key []byte) error {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	delete(sm.data, string(key))

	return nil
}

// GetOldestEpoch -
func (sm *StorerMock) GetOldestEpoch() (uint32, error) {
	return 0, nil
}

// ClearCache -
func (sm *StorerMock) ClearCache() {
}

// DestroyUnit -
func (sm *StorerMock) DestroyUnit() error {
	return nil
}

// RangeKeys -
func (sm *StorerMock) RangeKeys(_ func(key []byte, val []byte) bool) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *StorerMock) IsInterfaceNil() bool {
	return sm == nil
}
