package badgerdb

import (
	"github.com/dgraph-io/badger"
)

type batch struct {
	db    *badger.DB
	batch *badger.Txn
}

// NewBatch creates a batch
func NewBatch(db *DB) *batch {
	badgerDB := db.db

	b := badgerDB.NewTransaction(true)

	return &batch{
		db:    db.db,
		batch: b,
	}
}

// Put inserts one entry - key, value pair - into the batch
func (b *batch) Put(key []byte, val []byte) error {
	return b.batch.Set(key, val)
}

// Delete deletes the entry for the provided key from the batch
func (b *batch) Delete(key []byte) error {
	return b.batch.Delete(key)
}

// Reset clears the contents of the batch
func (b *batch) Reset() {
	b.batch = b.db.NewTransaction(true)
}
