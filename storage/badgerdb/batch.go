package badgerdb

import (
	"github.com/dgraph-io/badger"
)

type batch struct {
	db    *badger.DB
	batch *badger.Txn
}

func NewBatch(db *DB) *batch {
	badgerDB := db.db

	b := badgerDB.NewTransaction(true)

	return &batch{
		db:    db.db,
		batch: b,
	}
}

func (b *batch) Put(key []byte, val []byte) error {
	return b.batch.Set(key, val)
}

func (b *batch) Delete(key []byte) error {
	return b.batch.Delete(key)
}

func (b *batch) Reset() {
	b.batch = b.db.NewTransaction(true)
}
