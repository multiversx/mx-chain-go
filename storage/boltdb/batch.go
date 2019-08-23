package boltdb

import (
	"github.com/boltdb/bolt"
)

type batch struct {
	db *DB
}

// NewBatch creates a batch
func NewBatch(db *DB) *batch {
	return &batch{
		db: db,
	}
}

// Put inserts one entry - key, value pair - into the batch
func (b *batch) Put(key []byte, val []byte) error {
	err := b.db.db.Batch(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(b.db.parentFolder)).Put(key, val)
	})

	return err
}

// Delete deletes the entry for the provided key from the batch
func (b *batch) Delete(key []byte) error {
	err := b.db.db.Batch(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(b.db.parentFolder)).Delete(key)
	})

	return err
}

// Reset clears the contents of the batch
func (b *batch) Reset() {
	// nothing to do
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *batch) IsInterfaceNil() bool {
	if b == nil {
		return true
	}
	return false
}
