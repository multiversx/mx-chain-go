package badgerdb

import (
	"sync"

	"github.com/dgraph-io/badger"
)

type Batch struct {
	batch         *badger.WriteBatch
	mutBatchWrite sync.RWMutex
}

func NewBatch() *Batch {
	return &Batch{
		batch: &badger.WriteBatch{},
	}
}

func (b *Batch) Put(key []byte, val []byte) error {
	b.mutBatchWrite.Lock()
	b.batch.Set(key, val, 0)
	b.mutBatchWrite.Unlock()

	return nil
}

func (b *Batch) Delete(key []byte) error {
	b.mutBatchWrite.Lock()
	b.batch.Delete(key)
	b.mutBatchWrite.Unlock()

	return nil
}

func (b *Batch) Reset() {
	b.mutBatchWrite.Lock()
	b.batch = &badger.WriteBatch{}
	b.mutBatchWrite.Unlock()
}
