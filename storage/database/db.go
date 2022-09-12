package database

import (
	"github.com/ElrondNetwork/elrond-go-storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// DB represents the memory database storage. It holds a map of key value pairs
// and a mutex to handle concurrent accesses to the map
type DB = memorydb.DB

// NewMemDB creates a new memorydb object
func NewMemDB() *DB {
	return memorydb.New()
}

// NewlruDB creates a lruDB according to size
func NewlruDB(size uint32) (storage.Persister, error) {
	return memorydb.NewlruDB(size)
}
