package database

import (
	"errors"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-storage-go/leveldb"
)

var log = logger.GetOrCreate("storage/database")

// ErrNilIDProvider signals that a nil id provider was provided
var ErrNilIDProvider = errors.New("nil id provider")

type persisterIDProvider interface {
	ComputeId(key []byte) uint32
	NumberOfShards() uint32
	GetShardIDs() []uint32
	IsInterfaceNil() bool
}

type shardedPersister struct {
	persisters map[uint32]storage.Persister
	idProvider persisterIDProvider
}

// NewShardedPersister will created a new sharded persister
func NewShardedPersister(path string, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int, idProvider persisterIDProvider) (*shardedPersister, error) {
	if check.IfNil(idProvider) {
		return nil, ErrNilIDProvider
	}

	persisters := make(map[uint32]storage.Persister)
	for _, shardID := range idProvider.GetShardIDs() {
		newPath := updatePathWithShardID(path, shardID)
		db, err := leveldb.NewDB(newPath, batchDelaySeconds, maxBatchSize, maxOpenFiles)
		if err != nil {
			return nil, err
		}
		persisters[shardID] = db
	}

	return &shardedPersister{
		persisters: persisters,
		idProvider: idProvider,
	}, nil
}

func updatePathWithShardID(path string, shardID uint32) string {
	return fmt.Sprintf("%s_%d", path, shardID)
}

func (s *shardedPersister) computeID(key []byte) uint32 {
	return s.idProvider.ComputeId(key)
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
			log.Error("failed to close persister", "error", err)
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
		err := persister.Destroy()
		if err != nil {
			return err
		}
	}

	return nil
}

// DestroyClosed removes the already closed persistence medium stored data
func (s *shardedPersister) DestroyClosed() error {
	for _, persister := range s.persisters {
		err := persister.DestroyClosed()
		if err != nil {
			return err
		}
	}

	return nil
}

// RangeKeys will iterate over all contained pairs, in all persisters, calling te provided handler
func (s *shardedPersister) RangeKeys(handler func(key []byte, val []byte) bool) {
	for _, persister := range s.persisters {
		persister.RangeKeys(handler)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *shardedPersister) IsInterfaceNil() bool {
	return s == nil
}
