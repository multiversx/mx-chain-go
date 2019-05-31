package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type batch struct {
	batch *leveldb.Batch
}

// NewBatch creates a batch
func NewBatch() *batch {
	return &batch{
		batch: &leveldb.Batch{},
	}
}

// Put inserts one entry - key, value pair - into the batch
func (b *batch) Put(key []byte, val []byte) error {
	b.batch.Put(key, val)

	return nil
}

// Delete deletes the entry for the provided key from the batch
func (b *batch) Delete(key []byte) error {
	b.batch.Delete(key)

	return nil
}

// Reset clears the contents of the batch
func (b *batch) Reset() {
	b.batch.Reset()
}
