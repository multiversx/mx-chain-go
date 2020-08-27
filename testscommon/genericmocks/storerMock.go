package genericmocks

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/marshal"
)

// StorerMock -
type StorerMock struct {
	mutex sync.RWMutex
	Data  map[string][]byte
}

// NewStorerMock -
func NewStorerMock() *StorerMock {
	return &StorerMock{
		Data: make(map[string][]byte),
	}
}

// GetFromEpoch -
func (sm *StorerMock) GetFromEpoch(key []byte, _ uint32) ([]byte, error) {
	return sm.Get(key)
}

// GetBulkFromEpoch -
func (sm *StorerMock) GetBulkFromEpoch(keys [][]byte, _ uint32) (map[string][]byte, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := map[string][]byte{}

	for _, key := range keys {
		value, ok := sm.Data[string(key)]
		if ok {
			result[string(key)] = value
		}
	}

	return result, nil
}

// HasInEpoch -
func (sm *StorerMock) HasInEpoch(key []byte, _ uint32) error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	_, ok := sm.Data[string(key)]
	if ok {
		return nil
	}

	return fmt.Errorf("not in epoch")
}

// Put -
func (sm *StorerMock) Put(key, data []byte) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.Data[string(key)] = data
	return nil
}

// PutWithMarshalizer -
func (sm *StorerMock) PutWithMarshalizer(key []byte, obj interface{}, marshalizer marshal.Marshalizer) error {
	data, err := marshalizer.Marshal(obj)
	if err != nil {
		return err
	}

	return sm.Put(key, data)
}

// Get -
func (sm *StorerMock) Get(key []byte) ([]byte, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	value, ok := sm.Data[string(key)]
	if !ok {
		return nil, fmt.Errorf("key: %s not found", hex.EncodeToString(key))
	}

	return value, nil
}

// SearchFirst -
func (sm *StorerMock) SearchFirst(key []byte) ([]byte, error) {
	return sm.Get(key)
}

// Close -
func (sm *StorerMock) Close() error {
	return nil
}

// Has -
func (sm *StorerMock) Has(key []byte) error {
	_, err := sm.Get(key)
	return err
}

// Remove -
func (sm *StorerMock) Remove(_ []byte) error {
	return errors.New("not implemented")
}

// ClearCache -
func (sm *StorerMock) ClearCache() {
}

// DestroyUnit -
func (sm *StorerMock) DestroyUnit() error {
	return nil
}

// RangeKeys -
func (sm *StorerMock) RangeKeys(handler func(key []byte, val []byte) bool) {
	if handler == nil {
		return
	}

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for k, v := range sm.Data {
		shouldContinue := handler([]byte(k), v)
		if !shouldContinue {
			return
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *StorerMock) IsInterfaceNil() bool {
	return sm == nil
}
