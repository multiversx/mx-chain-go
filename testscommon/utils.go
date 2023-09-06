package testscommon

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

// HashSize holds the size of a typical hash used by the protocol
const HashSize = 32

// AddTimestampSuffix -
func AddTimestampSuffix(tag string) string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s_%s", tag, timestamp)
}

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

// StorerWithStats -
type StorerWithStats struct {
	storage.Storer
}

// GetWithStats -
func (ss *StorerWithStats) GetWithStats(key []byte) ([]byte, bool, error) {
	v, err := ss.Get(key)
	return v, false, err
}

// CreateDefaultMemStorerWithStats will create a new in-memory storer with stats component
func CreateDefaultMemStorerWithStats() storage.StorerWithStats {
	storerUnit := CreateMemUnit()
	return &StorerWithStats{storerUnit}
}

// CreateMemStorerWithStats -
func CreateMemStorerWithStats(storer storage.Storer) storage.StorerWithStats {
	return &StorerWithStats{storer}
}
