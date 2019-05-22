package leveldb

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const maxOpenFilesPerTable = 50

// read + write + execute for owner only
const rwxOwner = 0700

// DB holds a pointer to the leveldb database and the path to where it is stored.
type DB struct {
	db   *leveldb.DB
	path string
}

// NewDB is a constructor for the leveldb persister
// It creates the files in the location given as parameter
func NewDB(path string) (s *DB, err error) {
	err = os.MkdirAll(path, rwxOwner)
	if err != nil {
		return nil, err
	}

	options := &opt.Options{
		OpenFilesCacheCapacity: maxOpenFilesPerTable,
	}

	db, err := leveldb.OpenFile(path, options)
	if err != nil {
		return nil, err
	}

	dbStore := &DB{
		db:   db,
		path: path,
	}

	return dbStore, nil
}

// Put adds the value to the (key, val) storage medium
func (s *DB) Put(key, val []byte) error {
	return s.db.Put(key, val, nil)
}

// Get returns the value associated to the key
func (s *DB) Get(key []byte) ([]byte, error) {
	has, err := s.db.Has(key, nil)
	if err != nil || !has {
		return nil, errors.New(fmt.Sprintf("key: %s not found", base64.StdEncoding.EncodeToString(key)))
	}

	data, err := s.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, errors.New(fmt.Sprintf("key: %s not found", base64.StdEncoding.EncodeToString(key)))
	}

	return data, nil
}

// Has returns true if the given key is present in the persistance medium
func (s *DB) Has(key []byte) error {
	has, err := s.db.Has(key, nil)
	if err != nil {
		return err
	}

	if has {
		return nil
	}

	return errors.New("Key not found")
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
	return s.db.Delete(key, nil)
}

// Destroy removes the storage medium stored data
func (s *DB) Destroy() error {
	_ = s.db.Close()
	err := os.RemoveAll(s.path)
	return err
}
