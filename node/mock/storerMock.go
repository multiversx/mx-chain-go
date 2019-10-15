package mock

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
)

type StorerMock struct {
	mut  sync.Mutex
	data map[string][]byte
}

func NewStorerMock() *StorerMock {
	return &StorerMock{
		data: make(map[string][]byte),
	}
}

func (sm *StorerMock) Put(key, data []byte) error {
	sm.mut.Lock()
	defer sm.mut.Unlock()
	sm.data[string(key)] = data

	return nil
}

func (sm *StorerMock) Get(key []byte) ([]byte, error) {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	val, ok := sm.data[string(key)]
	if !ok {
		return nil, errors.New(fmt.Sprintf("key: %s not found", base64.StdEncoding.EncodeToString(key)))
	}

	return val, nil
}

func (sm *StorerMock) Has(key []byte) error {
	return errors.New("not implemented")
}

func (sm *StorerMock) Remove(key []byte) error {
	return errors.New("not implemented")
}

func (sm *StorerMock) ClearCache() {
}

func (sm *StorerMock) DestroyUnit() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *StorerMock) IsInterfaceNil() bool {
	if sm == nil {
		return true
	}
	return false
}
