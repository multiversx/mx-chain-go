package database

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-storage-go/leveldb"
)

var log = logger.GetOrCreate("storage/database")

type persisterIDProvider interface {
	ComputeIdFromBytes(address []byte) uint32
}

type shardedPersister struct {
	persistersMut    sync.RWMutex
	persisters       map[uint32]storage.Persister
	shardCoordinator sharding.Coordinator
}

func NewShardedPersister(path string, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int, shardCoordinator sharding.Coordinator) (*shardedPersister, error) {
	persisters := make(map[uint32]storage.Persister)
	for _, shardID := range getShardIDs(shardCoordinator) {
		newPath := updatePathWithShardID(path, shardID)
		db, err := leveldb.NewDB(newPath, batchDelaySeconds, maxBatchSize, maxOpenFiles)
		if err != nil {
			return nil, err
		}
		persisters[shardID] = db
	}

	return &shardedPersister{
		persisters:       persisters,
		shardCoordinator: shardCoordinator,
	}, nil
}

func updatePathWithShardID(path string, shardID uint32) string {
	if shardID == core.MetachainShardId {
		return fmt.Sprintf("%s_%s", path, "Meta")
	}

	return fmt.Sprintf("%s_%d", path, shardID)
}

func getShardIDs(shardCoordinator sharding.Coordinator) []uint32 {
	shardIDs := make([]uint32, shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		shardIDs[i] = i
	}
	shardIDs[shardCoordinator.NumberOfShards()] = core.MetachainShardId

	return shardIDs
}

func (s *shardedPersister) computeID(key []byte) uint32 {
	return s.shardCoordinator.ComputeId(key)
}

// Put add the value to the (key, val) persistence medium
func (s *shardedPersister) Put(key []byte, val []byte) error {
	return s.persisters[s.computeID(key)].Put(key, val)
}

// Get gets the value associated to the key
func (s *shardedPersister) Get(key []byte) ([]byte, error) {
	return s.persisters[s.computeID(key)].Get(key)
}

// Has returns true if the given key is present in the persistence medium
func (s *shardedPersister) Has(key []byte) error {
	return s.persisters[s.computeID(key)].Has(key)
}

// Close closes the files/resources associated to the persistence medium
func (s *shardedPersister) Close() error {
	closedSuccessfully := true
	for _, persister := range s.persisters {
		err := persister.Close()
		if err != nil {
			closedSuccessfully = false
		}
	}

	if !closedSuccessfully {
		return storage.ErrClosingPersisters
	}

	return nil
}

// Remove removes the data associated to the given key
func (s *shardedPersister) Remove(key []byte) error {
	return s.persisters[s.computeID(key)].Remove(key)
}

// Destroy removes the persistence medium stored data
func (s *shardedPersister) Destroy() error {
	for _, persister := range s.persisters {
		err := persister.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// DestroyClosed removes the already closed persistence medium stored data
func (s *shardedPersister) DestroyClosed() error {
	for _, persister := range s.persisters {
		err := persister.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *shardedPersister) RangeKeys(handler func(key []byte, val []byte) bool) {
	panic("not implemented") // TODO: Implement
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *shardedPersister) IsInterfaceNil() bool {
	panic("not implemented") // TODO: Implement
}
