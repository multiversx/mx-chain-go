package disabled

import (
	"github.com/multiversx/mx-chain-go/storage"
)

type persister struct{}

// NewPersister returns a new instance of this disabled persister
func NewPersister() *persister {
	return &persister{}
}

// Put returns nil
func (p *persister) Put(_, _ []byte) error {
	return nil
}

// Get returns nil and ErrKeyNotFound
func (p *persister) Get(_ []byte) ([]byte, error) {
	return nil, storage.ErrKeyNotFound
}

// Has returns ErrKeyNotFound
func (p *persister) Has(_ []byte) error {
	return storage.ErrKeyNotFound
}

// Close returns nil
func (p *persister) Close() error {
	return nil
}

// Remove returns nil
func (p *persister) Remove(_ []byte) error {
	return nil
}

// Destroy returns nil
func (p *persister) Destroy() error {
	return nil
}

// DestroyClosed returns nil
func (p *persister) DestroyClosed() error {
	return nil
}

// RangeKeys does nothing
func (p *persister) RangeKeys(_ func(key []byte, val []byte) bool) {}

// IsInterfaceNil returns true if there is no value under the interface
func (p *persister) IsInterfaceNil() bool {
	return p == nil
}
