package mock

import (
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
)

var _ storage.Persister = (*countingDB)(nil)

type countingDB struct {
	db      *database.MemDB
	nrOfPut int
}

// NewCountingDB returns a new instance of countingDB
func NewCountingDB() *countingDB {
	return &countingDB{database.NewMemDB(), 0}
}

// Put will add the given key-value pair in the db
func (cdb *countingDB) Put(key, val []byte) error {
	_ = cdb.db.Put(key, val)
	cdb.nrOfPut++
	return nil
}

// Get will return the value for the given key, if exists
func (cdb *countingDB) Get(key []byte) ([]byte, error) {
	return cdb.db.Get(key)
}

// Has will return true if the db has the given key stored
func (cdb *countingDB) Has(key []byte) error {
	return cdb.db.Has(key)
}

// Close will close the db
func (cdb *countingDB) Close() error {
	return cdb.db.Close()
}

// Remove will remove the key-value pair for the given key, if found in the db
func (cdb *countingDB) Remove(key []byte) error {
	return cdb.db.Remove(key)
}

// Destroy will destroy the db
func (cdb *countingDB) Destroy() error {
	return cdb.db.Destroy()
}

// DestroyClosed will destroy an already closed db
func (cdb *countingDB) DestroyClosed() error {
	return cdb.Destroy()
}

// Reset will reset the number of time the Put method was called
func (cdb *countingDB) Reset() {
	cdb.nrOfPut = 0
}

// GetCounter will return the number of times the Put method was called
func (cdb *countingDB) GetCounter() int {
	return cdb.nrOfPut
}

// RangeKeys will call the handler on all (key, value) pairs
func (cdb *countingDB) RangeKeys(handler func(key []byte, value []byte) bool) {
	cdb.db.RangeKeys(handler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (cdb *countingDB) IsInterfaceNil() bool {
	return cdb == nil
}
