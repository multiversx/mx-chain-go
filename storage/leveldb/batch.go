package leveldb

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

var _ storage.Batcher = (*batch)(nil)

type batch struct {
	batch       *leveldb.Batch
	cachedData  map[string][]byte
	removedData map[string]struct{}
	mutBatch    sync.RWMutex
}

// NewBatch creates a batch
func NewBatch() *batch {
	return &batch{
		batch:       &leveldb.Batch{},
		cachedData:  make(map[string][]byte),
		removedData: make(map[string]struct{}),
		mutBatch:    sync.RWMutex{},
	}
}

// Put inserts one entry - key, value pair - into the batch
func (b *batch) Put(key []byte, val []byte) error {
	b.mutBatch.Lock()
	b.batch.Put(key, val)
	b.cachedData[string(key)] = val
	delete(b.removedData, string(key))
	b.mutBatch.Unlock()
	return nil
}

// Delete deletes the entry for the provided key from the batch
func (b *batch) Delete(key []byte) error {
	b.mutBatch.Lock()
	b.batch.Delete(key)
	b.removedData[string(key)] = struct{}{}
	delete(b.cachedData, string(key))
	b.mutBatch.Unlock()
	return nil
}

// Reset clears the contents of the batch
func (b *batch) Reset() {
	b.mutBatch.Lock()
	b.batch.Reset()
	b.cachedData = make(map[string][]byte)
	b.removedData = make(map[string]struct{})
	b.mutBatch.Unlock()
}

// Get returns the value
func (b *batch) Get(key []byte) []byte {
	b.mutBatch.RLock()
	defer b.mutBatch.RUnlock()

	return b.cachedData[string(key)]
}

// IsRemoved returns true if the key is marked for removal
func (b *batch) IsRemoved(key []byte) bool {
	b.mutBatch.RLock()
	defer b.mutBatch.RUnlock()

	_, found := b.removedData[string(key)]

	return found
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *batch) IsInterfaceNil() bool {
	return b == nil
}
