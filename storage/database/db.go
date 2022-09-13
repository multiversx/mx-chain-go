package database

import (
	"github.com/ElrondNetwork/elrond-go-storage/leveldb"
	"github.com/ElrondNetwork/elrond-go-storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// MemDB represents the memory database storage. It holds a map of key value pairs
// and a mutex to handle concurrent accesses to the map
type MemDB = memorydb.DB

// LevelDB holds a pointer to the leveldb database and the path to where it is stored.
type LevelDB = leveldb.DB

// SerialLevelDB holds a pointer to the leveldb database and the path to where it is stored.
type SerialLevelDB = leveldb.SerialDB

// NewMemDB creates a new memorydb object
func NewMemDB() *MemDB {
	return memorydb.New()
}

// NewlruDB creates a lruDB according to size
func NewlruDB(size uint32) (storage.Persister, error) {
	return memorydb.NewlruDB(size)
}

// NewLevelDB is a constructor for the leveldb persister
// It creates the files in the location given as parameter
func NewLevelDB(path string, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int) (s *LevelDB, err error) {
	return leveldb.NewDB(path, batchDelaySeconds, maxBatchSize, maxOpenFiles)
}

// NewSerialDB is a constructor for the leveldb persister
// It creates the files in the location given as parameter
func NewSerialDB(path string, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int) (s *SerialLevelDB, err error) {
	return leveldb.NewSerialDB(path, batchDelaySeconds, maxBatchSize, maxOpenFiles)
}
