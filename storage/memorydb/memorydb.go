package memorydb

import (
	"errors"
	"fmt"
	"sync"
)

type DB struct {
	db   map[string][]byte
	mutx sync.RWMutex
}

func New() (*DB, error) {
	return &DB{
		db:   make(map[string][]byte),
		mutx: sync.RWMutex{},
	}, nil
}

// Add the value to the (key, val) storage medium
func (s *DB) Put(key, val []byte) error {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	s.db[string(key)] = val

	return nil
}

// gets the value associated to the key
func (s *DB) Get(key []byte) ([]byte, error) {
	s.mutx.RLock()
	defer s.mutx.RUnlock()

	val, ok := s.db[string(key)]

	if !ok {
		return nil, errors.New(fmt.Sprintf("key %s not found", key))
	}

	return val, nil
}

// returns true if the given key is present in the persistance medium
func (s *DB) Has(key []byte) (bool, error) {
	s.mutx.RLock()
	defer s.mutx.RUnlock()

	_, ok := s.db[string(key)]

	return ok, nil
}

// initialized the storage medium and prepares it for usage
func (s *DB) Init() error {
	// no special initialization needed
	return nil
}

// Closes the files/resources associated to the storage medium
func (s *DB) Close() error {
	// nothing to do
	return nil
}

// Removes the data associated to the given key
func (s *DB) Remove(key []byte) error {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	delete(s.db, string(key))

	return nil
}

// Removes the storage medium stored data
func (s *DB) Destroy() error {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	s.db = make(map[string][]byte)

	return nil
}
