package state

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageMock "github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/ElrondNetwork/elrond-go/storage/storageunit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
)

// CreateTestingTriePruningStorer creates a new trie pruning storer that is used for testing
func CreateTestingTriePruningStorer(coordinator sharding.Coordinator, notifier pruning.EpochStartNotifier) (storage.Storer, *persisterMap, error) {
	cacheConf := storageunit.CacheConfig{
		Capacity: 10,
		Type:     "LRU",
		Shards:   3,
	}
	dbConf := storageunit.DBConfig{
		FilePath:          "path/Epoch_0/Shard_1",
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 500,
		MaxBatchSize:      1,
		MaxOpenFiles:      1000,
	}

	persistersMap := NewPersistersMap()
	persisterFactory := &storageMock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			return persistersMap.GetPersister(path), nil
		},
	}
	args := &pruning.StorerArgs{
		PruningEnabled:         true,
		Identifier:             "id",
		ShardCoordinator:       coordinator,
		PathManager:            &testscommon.PathManagerStub{},
		CacheConf:              cacheConf,
		DbPath:                 dbConf.FilePath,
		PersisterFactory:       persisterFactory,
		NumOfEpochsToKeep:      4,
		NumOfActivePersisters:  4,
		Notifier:               notifier,
		OldDataCleanerProvider: &testscommon.OldDataCleanerProviderStub{},
		CustomDatabaseRemover:  &testscommon.CustomDatabaseRemoverStub{},
		MaxBatchSize:           10,
	}

	tps, err := pruning.NewTriePruningStorer(args)
	return tps, persistersMap, err
}

type persisterMap struct {
	persisters map[string]storage.Persister
	mutex      sync.Mutex
}

// NewPersistersMap returns a new persisterMap
func NewPersistersMap() *persisterMap {
	return &persisterMap{
		persisters: make(map[string]storage.Persister),
		mutex:      sync.Mutex{},
	}
}

// GetPersister returns the persister for the given path, or creates a new persister if it does not exist
func (pm *persisterMap) GetPersister(path string) storage.Persister {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	persister, exists := pm.persisters[path]
	if !exists {
		persister = memorydb.New()
		pm.persisters[path] = persister
	}

	return persister
}
