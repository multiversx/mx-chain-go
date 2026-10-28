package testscommon

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-go/config"
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

// NewNTPGoogleConfig creates an NTPConfig object that configures NTP to use a predefined list of hosts. This is useful
// for tests, for example, to avoid loading a configuration file just to have a NTPConfig
func NewNTPGoogleConfig() config.NTPConfig {
	return config.NTPConfig{
		Hosts:                []string{"time.google.com", "time.cloudflare.com", "time.apple.com", "time.windows.com"},
		Port:                 123,
		Version:              0,
		TimeoutMilliseconds:  100,
		SyncPeriodSeconds:    3600,
		OutOfBoundsThreshold: 120,
	}
}
