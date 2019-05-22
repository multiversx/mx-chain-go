package badgerdb

import (
	"os"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
)

// DB holds a pointer to the badger database and the path to where it is stored.
type DB struct {
	db   *badger.DB
	path string
}

// NewDB is a constructor for the badger persister
// It creates the files in the location given as parameter
func NewDB(path string) (s *DB, err error) {
	os.MkdirAll(path, 0777)
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	opts.ValueLogLoadingMode = options.FileIO
	db, err := badger.Open(opts)

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
	err := s.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, val)
		return err
	})

	return err
}

// Get returns the value associated to the key
func (s *DB) Get(key []byte) ([]byte, error) {
	var value []byte

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)

		if err != nil {
			return err
		}

		value, err = item.ValueCopy(nil)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

// Has returns true if the given key is present in the persistance medium
func (s *DB) Has(key []byte) error {
	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)

		if err != nil {
			return err
		}

		return nil
	})

	return err
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
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Destroy removes the storage medium stored data
func (s *DB) Destroy() error {
	err := os.RemoveAll(s.path)
	return err
}
