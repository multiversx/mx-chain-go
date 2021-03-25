package genericMocks

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/container"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// StorerMock -
type StorerMock struct {
	mutex        sync.RWMutex
	Name         string
	DataByEpoch  map[uint32]*container.MutexMap
	currentEpoch atomic.Uint32
}

// NewStorerMock -
func NewStorerMock(name string, currentEpoch uint32) *StorerMock {
	sm := &StorerMock{
		Name:        name,
		DataByEpoch: make(map[uint32]*container.MutexMap),
	}

	sm.SetCurrentEpoch(currentEpoch)
	return sm
}

// SetCurrentEpoch -
func (sm *StorerMock) SetCurrentEpoch(epoch uint32) {
	sm.currentEpoch.Set(epoch)
}

// GetCurrentEpochData -
func (sm *StorerMock) GetCurrentEpochData() *container.MutexMap {
	return sm.GetEpochData(sm.currentEpoch.Get())
}

// GetEpochData -
func (sm *StorerMock) GetEpochData(epoch uint32) *container.MutexMap {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	data, ok := sm.DataByEpoch[epoch]
	if ok {
		return data
	}

	data = container.NewMutexMap()
	sm.DataByEpoch[epoch] = data

	return data
}

// GetFromEpoch -
func (sm *StorerMock) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	data := sm.GetEpochData(epoch)
	value, ok := data.Get(string(key))
	if !ok {
		return nil, sm.newErrNotFound(key, epoch)
	}

	return value.([]byte), nil
}

// GetBulkFromEpoch -
func (sm *StorerMock) GetBulkFromEpoch(keys [][]byte, epoch uint32) (map[string][]byte, error) {
	data := sm.GetEpochData(epoch)
	result := map[string][]byte{}

	for _, key := range keys {
		value, ok := data.Get(string(key))
		if ok {
			result[string(key)] = value.([]byte)
		}
	}

	return result, nil
}

// HasInEpoch -
func (sm *StorerMock) HasInEpoch(key []byte, epoch uint32) error {
	data := sm.GetEpochData(epoch)

	_, ok := data.Get(string(key))
	if ok {
		return nil
	}

	return sm.newErrNotFound(key, epoch)
}

// Put -
func (sm *StorerMock) Put(key, value []byte) error {
	data := sm.GetCurrentEpochData()
	data.Set(string(key), value)
	return nil
}

// PutInEpoch -
func (sm *StorerMock) PutInEpoch(key, value []byte, epoch uint32) error {
	data := sm.GetEpochData(epoch)
	data.Set(string(key), value)
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
	data := sm.GetCurrentEpochData()
	value, ok := data.Get(string(key))
	if !ok {
		return nil, sm.newErrNotFound(key, sm.currentEpoch.Get())
	}

	return value.([]byte), nil
}

// GetFromEpochWithMarshalizer -
func (sm *StorerMock) GetFromEpochWithMarshalizer(key []byte, epoch uint32, obj interface{}, marshalizer marshal.Marshalizer) error {
	data, err := sm.GetFromEpoch(key, epoch)
	if err != nil {
		return err
	}

	err = marshalizer.Unmarshal(obj, data)
	if err != nil {
		return err
	}

	return nil
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
	return sm.HasInEpoch(key, sm.currentEpoch.Get())
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
func (sm *StorerMock) RangeKeys(handler func(key []byte, value []byte) bool) {
	if handler == nil {
		return
	}

	data := sm.GetCurrentEpochData()

	for _, key := range data.Keys() {
		value, ok := data.Get(key)
		if !ok {
			continue
		}

		shouldContinueRange := handler([]byte(key.(string)), value.([]byte))
		if !shouldContinueRange {
			return
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *StorerMock) IsInterfaceNil() bool {
	return sm == nil
}

func (sm *StorerMock) newErrNotFound(key []byte, epoch uint32) error {
	return fmt.Errorf("StorerMock: not found in %s: key = %s, epoch = %d", sm.Name, hex.EncodeToString(key), epoch)
}
