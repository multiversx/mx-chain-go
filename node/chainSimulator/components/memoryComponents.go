package components

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

// CreateMemUnit creates a new in-memory storage unit
func CreateMemUnit() storage.Storer {
	capacity := uint32(10)
	shards := uint32(1)
	sizeInBytes := uint64(0)
	cache, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: capacity, Shards: shards, SizeInBytes: sizeInBytes})
	persist, _ := database.NewlruDB(100000)
	unit, _ := storageunit.NewStorageUnit(cache, persist)

	return unit
}

type trieStorage struct {
	storage.Storer
}

// SetEpochForPutOperation does nothing
func (store *trieStorage) SetEpochForPutOperation(_ uint32) {
}

// GetFromOldEpochsWithoutAddingToCache tries to get directly the key
func (store *trieStorage) GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, core.OptionalUint32, error) {
	value, err := store.Get(key)

	return value, core.OptionalUint32{}, err
}

// GetFromLastEpoch tries to get directly the key
func (store *trieStorage) GetFromLastEpoch(key []byte) ([]byte, error) {
	return store.Get(key)
}

// PutInEpoch will put the key directly
func (store *trieStorage) PutInEpoch(key []byte, data []byte, _ uint32) error {
	return store.Put(key, data)
}

// PutInEpochWithoutCache will put the key directly
func (store *trieStorage) PutInEpochWithoutCache(key []byte, data []byte, _ uint32) error {
	return store.Put(key, data)
}

// GetLatestStorageEpoch returns 0
func (store *trieStorage) GetLatestStorageEpoch() (uint32, error) {
	return 0, nil
}

// GetFromCurrentEpoch tries to get directly the key
func (store *trieStorage) GetFromCurrentEpoch(key []byte) ([]byte, error) {
	return store.Get(key)
}

// GetFromEpoch tries to get directly the key
func (store *trieStorage) GetFromEpoch(key []byte, _ uint32) ([]byte, error) {
	return store.Get(key)
}

// RemoveFromCurrentEpoch removes directly the key
func (store *trieStorage) RemoveFromCurrentEpoch(key []byte) error {
	return store.Remove(key)
}

// RemoveFromAllActiveEpochs removes directly the key
func (store *trieStorage) RemoveFromAllActiveEpochs(key []byte) error {
	return store.Remove(key)
}

// CreateMemUnitForTries returns a special type of storer used on tries instances
func CreateMemUnitForTries() storage.Storer {
	return &trieStorage{
		Storer: CreateMemUnit(),
	}
}
