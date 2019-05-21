package boltdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
)

// DB holds a pointer to the boltdb database and the path to where it is stored.
type DB struct {
	db           *bolt.DB
	path         string
	parentFolder string
}

// NewDB is a constructor for the boltdb persister
// It creates the files in the location given as parameter
func NewDB(path string) (s *DB, err error) {
	os.Mkdir(path, 0777)

	if err != nil {
		return nil, err
	}
	// create a filename
	file := fmt.Sprintf("%s/%s", path, "data.db")
	db, err := bolt.Open(file, 0666, nil)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(file)
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

	return dbStore, nil
}

// Put adds the value to the (key, val) storage medium
func (s *DB) Put(key, val []byte) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.parentFolder))
		err := b.Put(key, val)
		return err
	})

	return err
}

// Get returns the value associated to the key
func (s *DB) Get(key []byte) ([]byte, error) {
	var val []byte

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.parentFolder))
		v := b.Get(key)

		if v == nil {
			return errors.New("Key not found")
		}

		val = append([]byte{}, v...)
		return nil
	})

	return val, err
}

// Has returns true if the given key is present in the persistance medium
func (s *DB) Has(key []byte) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.parentFolder))
		v := b.Get(key)

		if v == nil {
			return errors.New("Key not found")
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
