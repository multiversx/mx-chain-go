package boltdb

import (
	"github.com/boltdb/bolt"
)

type Batch struct {
	db *DB
	ch chan error
}

func NewBatch(db *DB) *Batch {
	return &Batch{
		db: db,
	}
}

func (b *Batch) Put(key []byte, val []byte) error {
	err := b.db.db.Batch(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(b.db.parentFolder)).Put(key, val)
	})

	return err
}

func (b *Batch) Delete(key []byte) error {
	err := b.db.db.Batch(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(b.db.parentFolder)).Delete(key)
	})

	return err
}

func (b *Batch) Reset() {
	_ = b.db.db.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.db.parentFolder))

		err := bucket.ForEach(func(k, v []byte) error {
			return bucket.Delete(k)
		})

		return err
	})
}
