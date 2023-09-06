package disabled

import (
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

const defaultCapacity = 10000
const defaultNumShards = 1
const zeroSize = 0

// CreateMemUnit creates an in-memory storer unit using maps
func CreateMemUnit() storage.Storer {
	cache, err := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: defaultCapacity, Shards: defaultNumShards, SizeInBytes: zeroSize})
	if err != nil {
		return nil
	}

	unit, err := storageunit.NewStorageUnit(cache, database.NewMemDB())
	if err != nil {
		return nil
	}

	return unit
}

// storerWithStats -
type storerWithStats struct {
	storage.Storer
}

// GetWithStats will trigger get operation with statistics
func (ss *storerWithStats) GetWithStats(key []byte) ([]byte, bool, error) {
	v, err := ss.Get(key)
	return v, false, err
}

// CreateMemStorerWithStats -
func CreateMemStorerWithStats() storage.StorerWithStats {
	storerUnit := CreateMemUnit()
	return &storerWithStats{storerUnit}
}
