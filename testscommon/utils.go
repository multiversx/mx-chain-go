package testscommon

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"

	"github.com/multiversx/mx-chain-core-go/core"
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

func ReadInitialAccounts(filePath string) ([]genesis.InitialAccountHandler, error) {
	initialAccounts := make([]*data.InitialAccount, 0)
	err := core.LoadJsonFile(&initialAccounts, filePath)
	if err != nil {
		return nil, err
	}

	var accounts []genesis.InitialAccountHandler
	for _, ia := range initialAccounts {
		accounts = append(accounts, ia)
	}

	return accounts, nil
}
