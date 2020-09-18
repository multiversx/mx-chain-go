package mock

import (
	"errors"
	"sync"
)

var errKeyNotFound = errors.New("key not found")

type persisterMock struct {
	values    map[string][]byte
	mutValues sync.RWMutex
}

// NewPersisterMock -
func NewPersisterMock() *persisterMock {
	return &persisterMock{
		values: make(map[string][]byte),
	}
}

// Put -
func (p *persisterMock) Put(key, val []byte) error {
	p.mutValues.Lock()
	p.values[string(key)] = val
	p.mutValues.Unlock()

	return nil
}

// Get -
func (p *persisterMock) Get(key []byte) ([]byte, error) {
	p.mutValues.RLock()
	val, ok := p.values[string(key)]
	p.mutValues.RUnlock()

	if !ok {
		return nil, errKeyNotFound
	}

	return val, nil
}

// Has -
func (p *persisterMock) Has(key []byte) error {
	_, err := p.Get(key)
	return err
}

// Init -
func (p *persisterMock) Init() error {
	return nil
}

// Close -
func (p *persisterMock) Close() error {
	return nil
}

// Remove -
func (p *persisterMock) Remove(key []byte) error {
	p.mutValues.Lock()
	delete(p.values, string(key))
	p.mutValues.Unlock()

	return nil
}

// Destroy -
func (p *persisterMock) Destroy() error {
	p.mutValues.Lock()
	p.values = make(map[string][]byte)
	p.mutValues.Unlock()

	return nil
}

// DestroyClosed -
func (p *persisterMock) DestroyClosed() error {
	return p.Destroy()
}

// RangeKeys -
func (p *persisterMock) RangeKeys(handler func(key []byte, val []byte) bool) {
	p.mutValues.RLock()
	for key, val := range p.values {
		handler([]byte(key), val)
	}
	p.mutValues.RUnlock()
}

// IsInterfaceNil -
func (p *persisterMock) IsInterfaceNil() bool {
	return p == nil
}
