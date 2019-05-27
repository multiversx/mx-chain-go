package leveldb

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

type Batch struct {
	batch         *leveldb.Batch
	mutBatchWrite sync.RWMutex
}

func NewBatch() *Batch {
	return &Batch{
		batch: &leveldb.Batch{},
	}
}

func (b *Batch) Put(key []byte, val []byte) error {
	b.mutBatchWrite.Lock()
	b.batch.Put(key, val)
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
	b.batch.Reset()
	b.mutBatchWrite.Unlock()
}
