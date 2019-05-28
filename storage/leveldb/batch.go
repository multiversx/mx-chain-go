package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type Batch struct {
	batch *leveldb.Batch
}

func NewBatch() *Batch {
	return &Batch{
		batch: &leveldb.Batch{},
	}
}

func (b *Batch) Put(key []byte, val []byte) error {
	b.batch.Put(key, val)

	return nil
}

func (b *Batch) Delete(key []byte) error {
	b.batch.Delete(key)

	return nil
}

func (b *Batch) Reset() {
	b.batch.Reset()
}
