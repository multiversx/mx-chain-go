package disabled

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
)

// Storer -
type Storer struct {
	mut  sync.Mutex
	data map[string][]byte
}

// NewDisabledStorer -
func NewDisabledStorer() *Storer {
	return &Storer{
		data: make(map[string][]byte),
	}
}

// Put -
func (sm *Storer) Put(key, data []byte) error {
	sm.mut.Lock()
	defer sm.mut.Unlock()
	sm.data[string(key)] = data

	return nil
}

// Get -
func (sm *Storer) Get(key []byte) ([]byte, error) {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	val, ok := sm.data[string(key)]
	if !ok {
		return nil, fmt.Errorf("key: %s not found", base64.StdEncoding.EncodeToString(key))
	}

	return val, nil
}

// GetFromEpoch -
func (sm *Storer) GetFromEpoch(key []byte, _ uint32) ([]byte, error) {
	return sm.Get(key)
}

// HasInEpoch -
func (sm *Storer) HasInEpoch(key []byte, epoch uint32) error {
	return errors.New("not implemented")
}

// SearchFirst -
func (sm *Storer) SearchFirst(key []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Close -
func (sm *Storer) Close() error {
	return nil
}

// Has -
func (sm *Storer) Has(key []byte) error {
	return errors.New("not implemented")
}

// Remove -
func (sm *Storer) Remove(key []byte) error {
	return errors.New("not implemented")
}

// ClearCache -
func (sm *Storer) ClearCache() {
}

// DestroyUnit -
func (sm *Storer) DestroyUnit() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *Storer) IsInterfaceNil() bool {
	return sm == nil
}
