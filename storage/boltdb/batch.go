package boltdb

import (
	"github.com/boltdb/bolt"
)

type batch struct {
	db *DB
}

func NewBatch(db *DB) *batch {
	return &batch{
		db: db,
	}
}

func (b *batch) Put(key []byte, val []byte) error {
	err := b.db.db.Batch(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(b.db.parentFolder)).Put(key, val)
	})

	return err
}

func (b *batch) Delete(key []byte) error {
	err := b.db.db.Batch(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(b.db.parentFolder)).Delete(key)
	})

	return err
}

func (b *batch) Reset() {
	// nothing to do
}
