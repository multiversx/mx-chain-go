package state

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	storageMock "github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
)

// CreateTriePruningStorer creates a new trie pruning storer that is used for testing
func CreateTriePruningStorer(coordinator sharding.Coordinator, notifier pruning.EpochStartNotifier) (storage.Storer, map[string]storage.Persister, error) {
	cacheConf := storageUnit.CacheConfig{
		Capacity: 10,
		Type:     "LRU",
		Shards:   3,
	}
	dbConf := storageUnit.DBConfig{
		FilePath:          "path/Epoch_0/Shard_1",
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 500,
		MaxBatchSize:      1,
		MaxOpenFiles:      1000,
	}

	lockPersisterMap := sync.Mutex{}
	persistersMap := make(map[string]storage.Persister)
	persisterFactory := &storageMock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			lockPersisterMap.Lock()
			defer lockPersisterMap.Unlock()

			persister, exists := persistersMap[path]
			if !exists {
				persister = memorydb.New()
				persistersMap[path] = persister
			}

			return persister, nil
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
