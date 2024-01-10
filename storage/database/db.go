package database

import (
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-storage-go/leveldb"
	"github.com/multiversx/mx-chain-storage-go/memorydb"
	"github.com/multiversx/mx-chain-storage-go/sharded"
)

// MemDB represents the memory database storage. It holds a map of key value pairs
// and a mutex to handle concurrent accesses to the map
type MemDB = memorydb.DB

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
func NewLevelDB(path string, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int) (s *leveldb.DB, err error) {
	return leveldb.NewDB(path, batchDelaySeconds, maxBatchSize, maxOpenFiles)
}

// NewSerialDB is a constructor for the leveldb persister
// It creates the files in the location given as parameter
func NewSerialDB(path string, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int) (s *leveldb.SerialDB, err error) {
	return leveldb.NewSerialDB(path, batchDelaySeconds, maxBatchSize, maxOpenFiles)
}

// NewShardIDProvider is a constructor for shard id provider
func NewShardIDProvider(numShards int32) (storage.ShardIDProvider, error) {
	return sharded.NewShardIDProvider(numShards)
}

// NewShardedPersister is a constructor for sharded persister based on provided db type
func NewShardedPersister(path string, persisterCreator storage.BasePersisterCreator, idPersister storage.ShardIDProvider) (s storage.Persister, err error) {
	return sharded.NewShardedPersister(path, persisterCreator, idPersister)
}
