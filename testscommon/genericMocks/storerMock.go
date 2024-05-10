package genericMocks

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/container"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
)

// StorerMock -
type StorerMock struct {
	mutex                      sync.RWMutex
	Name                       string
	DataByEpoch                map[uint32]*container.MutexMap
	shouldReturnErrKeyNotFound bool
	currentEpoch               atomic.Uint32
}

// NewStorerMock -
func NewStorerMock() *StorerMock {
	return NewStorerMockWithEpoch(0)
}

// NewStorerMockWithEpoch -
func NewStorerMockWithEpoch(currentEpoch uint32) *StorerMock {
	sm := &StorerMock{
		Name:        "",
		DataByEpoch: make(map[uint32]*container.MutexMap),
	}

	sm.SetCurrentEpoch(currentEpoch)
	return sm
}

// NewStorerMockWithErrKeyNotFound -
func NewStorerMockWithErrKeyNotFound(currentEpoch uint32) *StorerMock {
	sm := NewStorerMockWithEpoch(currentEpoch)
	sm.shouldReturnErrKeyNotFound = true

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

	value, ok := sm.DataByEpoch[epoch]
	if ok {
		return value
	}

	value = container.NewMutexMap()
	sm.DataByEpoch[epoch] = value

	return value
}

// GetFromEpoch -
func (sm *StorerMock) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	epochData := sm.GetEpochData(epoch)
	value, ok := epochData.Get(string(key))
	if !ok {
		return nil, sm.newErrNotFound(key, epoch)
	}

	return value.([]byte), nil
}

// GetBulkFromEpoch -
func (sm *StorerMock) GetBulkFromEpoch(keys [][]byte, epoch uint32) ([]data.KeyValuePair, error) {
	epochData := sm.GetEpochData(epoch)
	results := make([]data.KeyValuePair, 0, len(keys))

	for _, key := range keys {
		value, ok := epochData.Get(string(key))
		if ok {
			keyValue := data.KeyValuePair{Key: key, Value: value.([]byte)}
			results = append(results, keyValue)
		}
	}

	return results, nil
}

// hasInEpoch -
func (sm *StorerMock) hasInEpoch(key []byte, epoch uint32) error {
	epochData := sm.GetEpochData(epoch)

	_, ok := epochData.Get(string(key))
	if ok {
		return nil
	}

	return sm.newErrNotFound(key, epoch)
}

// Put -
func (sm *StorerMock) Put(key, value []byte) error {
	epochData := sm.GetCurrentEpochData()
	epochData.Set(string(key), value)
	return nil
}

// PutInEpoch -
func (sm *StorerMock) PutInEpoch(key, value []byte, epoch uint32) error {
	epochData := sm.GetEpochData(epoch)
	epochData.Set(string(key), value)
	return nil
}

// PutWithMarshalizer -
func (sm *StorerMock) PutWithMarshalizer(key []byte, obj interface{}, marshalizer marshal.Marshalizer) error {
	value, err := marshalizer.Marshal(obj)
	if err != nil {
		return err
	}

	return sm.Put(key, value)
}

// Get -
func (sm *StorerMock) Get(key []byte) ([]byte, error) {
	epochData := sm.GetCurrentEpochData()
	value, ok := epochData.Get(string(key))
	if !ok {
		return nil, sm.newErrNotFound(key, sm.currentEpoch.Get())
	}

	return value.([]byte), nil
}

// GetFromEpochWithMarshalizer -
func (sm *StorerMock) GetFromEpochWithMarshalizer(key []byte, epoch uint32, obj interface{}, marshalizer marshal.Marshalizer) error {
	value, err := sm.GetFromEpoch(key, epoch)
	if err != nil {
		return err
	}

	err = marshalizer.Unmarshal(obj, value)
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
	return sm.hasInEpoch(key, sm.currentEpoch.Get())
}

// RemoveFromCurrentEpoch -
func (sm *StorerMock) RemoveFromCurrentEpoch(_ []byte) error {
	return errors.New("not implemented")
}

// Remove -
func (sm *StorerMock) Remove(_ []byte) error {
	return errors.New("not implemented")
}

// ClearAll removes all data from the mock (useful in unit tests)
func (sm *StorerMock) ClearAll() {
	sm.DataByEpoch = make(map[uint32]*container.MutexMap)
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

	epochData := sm.GetCurrentEpochData()

	for _, key := range epochData.Keys() {
		value, ok := epochData.Get(key)
		if !ok {
			continue
		}

		shouldContinueRange := handler([]byte(key.(string)), value.([]byte))
		if !shouldContinueRange {
			return
		}
	}
}

// GetOldestEpoch -
func (sm *StorerMock) GetOldestEpoch() (uint32, error) {
	return 0, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *StorerMock) IsInterfaceNil() bool {
	return sm == nil
}

func (sm *StorerMock) newErrNotFound(key []byte, epoch uint32) error {
	if sm.shouldReturnErrKeyNotFound {
		return storage.ErrKeyNotFound
	}

	return fmt.Errorf("StorerMock: not found; key = %s, epoch = %d", hex.EncodeToString(key), epoch)
}
