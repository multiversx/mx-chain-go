package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type batch struct {
	batch *leveldb.Batch
}

func NewBatch() *batch {
	return &batch{
		batch: &leveldb.Batch{},
	}
}

func (b *batch) Put(key []byte, val []byte) error {
	b.batch.Put(key, val)

	return nil
}

func (b *batch) Delete(key []byte) error {
	b.batch.Delete(key)

	return nil
}

func (b *batch) Reset() {
	b.batch.Reset()
}
