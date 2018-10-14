package leveldb

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type DB struct {
	db   *leveldb.DB
	path string
}

func NewDB(path string) (s *DB, err error) {
	db, err := leveldb.OpenFile(path, nil)

	if err != nil {
		return nil, err
	}

	dbStore := &DB{
		db:   db,
		path: path,
	}

	return dbStore, nil
}

// Add the value to the (key, val) storage medium
func (s *DB) Put(key, val []byte) error {
	return s.db.Put(key, val, nil)
}

// gets the value associated to the key
func (s *DB) Get(key []byte) ([]byte, error) {
	has, err := s.db.Has(key, nil)

	if err != nil || !has {
		return nil, errors.New(fmt.Sprintf("key: %s not found", string(key)))
	}

	data, err := s.db.Get(key, nil)

	if err == leveldb.ErrNotFound {
		return nil, errors.New(fmt.Sprintf("key: %s not found", string(key)))
	}

	return data, nil
}

// returns true if the given key is present in the persistance medium
func (s *DB) Has(key []byte) (bool, error) {
	return s.db.Has(key, nil)
}

// initialized the storage medium and prepares it for usage
func (s *DB) Init() error {
	// no special initialization needed
	return nil
}

// Closes the files/resources associated to the storage medium
func (s *DB) Close() error {
	return s.db.Close()
}

// Removes the data associated to the given key
func (s *DB) Remove(key []byte) error {
	return s.db.Delete(key, nil)
}

// Removes the storage medium stored data
func (s *DB) Destroy() error {
	s.db.Close()
	err := os.RemoveAll(s.path)
	return err
}
