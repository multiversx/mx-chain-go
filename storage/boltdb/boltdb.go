package boltdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/boltdb/bolt"
)

// read + write + execute for owner only
const rwxOwner = 0700

// read + write for owner
const rwOwner = 0600

// DB holds a pointer to the boltdb database and the path to where it is stored.
type DB struct {
	db           *bolt.DB
	path         string
	parentFolder string
	batch        storage.Batcher
}

// NewDB is a constructor for the boltdb persister
// It creates the files in the location given as parameter
func NewDB(path string, batchDelaySeconds int, maxBatchSize int) (s *DB, err error) {
	err = os.MkdirAll(path, rwxOwner)
	if err != nil {
		return nil, err
	}
	// create a filename
	fPath := filepath.Join(path, "data.db")
	db, err := bolt.Open(fPath, rwOwner, nil)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(fPath)
	parentFolder := filepath.Base(dir)
	fmt.Println("Parent Folder: ", parentFolder)

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(parentFolder))
		if err != nil {
			return errors.New(fmt.Sprintf("create bucket: %s", err))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	dbStore := &DB{
		db:           db,
		path:         path,
		parentFolder: parentFolder,
	}

	dbStore.db.MaxBatchDelay = time.Duration(batchDelaySeconds) * time.Second
	dbStore.db.MaxBatchSize = maxBatchSize
	dbStore.batch = dbStore.createBatch()

	return dbStore, nil
}

// Put adds the value to the (key, val) storage medium
func (s *DB) Put(key, val []byte) error {
	return s.batch.Put(key, val)
}

// Get returns the value associated to the key
func (s *DB) Get(key []byte) ([]byte, error) {
	var val []byte

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.parentFolder))
		v := b.Get(key)
		if v == nil {
			return storage.ErrKeyNotFound
		}

		val = append([]byte{}, v...)

		return nil
	})

	return val, err
}

// CreateBatch returns a batcher to be used for batch writing data to the database
func (s *DB) createBatch() storage.Batcher {
	return NewBatch(s)
}

// Has returns true if the given key is present in the persistence medium
func (s *DB) Has(key []byte) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.parentFolder))
		v := b.Get(key)
		if v == nil {
			return storage.ErrKeyNotFound
		}

		return nil
	})
}

// Init initializes the storage medium and prepares it for usage
func (s *DB) Init() error {
	// no special initialization needed
	return nil
}

// Close closes the files/resources associated to the storage medium
func (s *DB) Close() error {
	return s.db.Close()
}

// Remove removes the data associated to the given key
func (s *DB) Remove(key []byte) error {
	_ = s.batch.Delete(key)

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.parentFolder))
		err := b.Delete(key)
		return err
	})

	return err
}

// Destroy removes the storage medium stored data
func (s *DB) Destroy() error {
	_ = s.db.Close()
	err := os.RemoveAll(s.path)
	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *DB) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}
