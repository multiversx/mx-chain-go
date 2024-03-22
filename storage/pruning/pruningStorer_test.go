package pruning_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/directoryhandler"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/storage/pathmanager"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("storage/pruning_test")

func getDummyConfig() (storageunit.CacheConfig, storageunit.DBConfig) {
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
	return cacheConf, dbConf
}

func getDefaultArgs() pruning.StorerArgs {
	cacheConf, dbConf := getDummyConfig()

	lockPersisterMap := sync.Mutex{}
	persistersMap := make(map[string]storage.Persister)
	persisterFactory := &mock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			lockPersisterMap.Lock()
			defer lockPersisterMap.Unlock()

			persister, exists := persistersMap[path]
			if !exists {
				persister = database.NewMemDB()
				persistersMap[path] = persister
			}

			return persister, nil
		},
	}

	epochsData := pruning.EpochArgs{
		NumOfEpochsToKeep:     2,
		NumOfActivePersisters: 2,
	}
	return pruning.StorerArgs{
		PruningEnabled:         true,
		Identifier:             "id",
		ShardCoordinator:       mock.NewShardCoordinatorMock(0, 2),
		PathManager:            &testscommon.PathManagerStub{},
		CacheConf:              cacheConf,
		DbPath:                 dbConf.FilePath,
		PersisterFactory:       persisterFactory,
		EpochsData:             epochsData,
		Notifier:               &mock.EpochStartNotifierStub{},
		OldDataCleanerProvider: &testscommon.OldDataCleanerProviderStub{},
		CustomDatabaseRemover:  &testscommon.CustomDatabaseRemoverStub{},
		MaxBatchSize:           10,
		PersistersTracker:      pruning.NewPersistersTracker(epochsData),
		StateStatsHandler:      disabled.NewStateStatistics(),
	}
}

func getDefaultArgsSerialDB() pruning.StorerArgs {
	cacheConf, dbConf := getDummyConfig()
	cacheConf.Capacity = 40
	persisterFactory := &mock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			return database.NewSerialDB(path, 1, 20, 10)
		},
	}
	pathManager := &testscommon.PathManagerStub{PathForEpochCalled: func(shardId string, epoch uint32, identifier string) string {
		return fmt.Sprintf("TestOnly-Epoch_%d/Shard_%s/%s", epoch, shardId, identifier)
	}}
	epochData := pruning.EpochArgs{
		NumOfEpochsToKeep:     3,
		NumOfActivePersisters: 2,
	}
	return pruning.StorerArgs{
		PruningEnabled:         true,
		Identifier:             "id",
		ShardCoordinator:       mock.NewShardCoordinatorMock(0, 2),
		PathManager:            pathManager,
		CacheConf:              cacheConf,
		DbPath:                 dbConf.FilePath,
		PersisterFactory:       persisterFactory,
		EpochsData:             epochData,
		Notifier:               &mock.EpochStartNotifierStub{},
		OldDataCleanerProvider: &testscommon.OldDataCleanerProviderStub{},
		CustomDatabaseRemover:  &testscommon.CustomDatabaseRemoverStub{},
		MaxBatchSize:           20,
		PersistersTracker:      pruning.NewPersistersTracker(epochData),
		StateStatsHandler:      disabled.NewStateStatistics(),
	}
}

func TestNewPruningStorer_InvalidNumberOfActivePersistersShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.EpochsData.NumOfActivePersisters = 0

	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrInvalidNumberOfPersisters, err)
}

func TestNewPruningStorer_NilPersistersTrackerShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.PersistersTracker = nil

	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrNilPersistersTracker, err)
}

func TestNewPruningStorer_NumEpochKeepLowerThanNumActiveShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.EpochsData.NumOfActivePersisters = 3
	args.EpochsData.NumOfEpochsToKeep = 2

	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrEpochKeepIsLowerThanNumActive, err)
}

func TestNewPruningStorer_NilEpochStartHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.Notifier = nil
	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrNilEpochStartNotifier, err)
}

func TestNewPruningStorer_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.ShardCoordinator = nil
	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrNilShardCoordinator, err)
}

func TestNewPruningStorer_NilPathManagerShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.PathManager = nil
	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrNilPathManager, err)
}

func TestNewPruningStorer_NilOldDataCleanerProviderShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.OldDataCleanerProvider = nil
	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrNilOldDataCleanerProvider, err)
}

func TestNewPruningStorer_NilCustomDatabaseRemoverProviderShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.CustomDatabaseRemover = nil
	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrNilCustomDatabaseRemover, err)
}

func TestNewPruningStorer_NilPersisterFactoryShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.PersisterFactory = nil
	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrNilPersisterFactory, err)
}

func TestNewPruningStorer_CacheSizeLowerThanBatchSizeShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.MaxBatchSize = 11
	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrCacheSizeIsLowerThanBatchSize, err)
}

func TestNewPruningStorer_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, err := pruning.NewPruningStorer(args)

	assert.NotNil(t, ps)
	assert.Nil(t, err)
	assert.False(t, ps.IsInterfaceNil())
}

func TestPruningStorer_PutAndGetShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, _ := pruning.NewPruningStorer(args)

	testKey, testVal := []byte("key"), []byte("value")
	err := ps.Put(testKey, testVal)
	assert.Nil(t, err)

	res, err := ps.Get(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)
}

func TestPruningStorer_Put_EpochWhichWasSetDoesNotExistShouldNotFind(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()

	persistersByPath := make(map[string]storage.Persister)
	persistersByPath["Epoch_0"] = database.NewMemDB()
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := database.NewMemDB()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	ps, _ := pruning.NewPruningStorer(args)

	expectedEpoch := uint32(37)
	ps.SetEpochForPutOperation(expectedEpoch)

	testKey, testVal := []byte("key1"), []byte("val1")
	_ = ps.Put(testKey, testVal) // the persister for epoch 37 does not exist - will put in storer 0

	res, _ := ps.Get(testKey)
	assert.Equal(t, testVal, res)
}

func TestPruningStorer_Put_ShouldPutInSpecifiedEpoch(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()

	persistersByPath := make(map[string]storage.Persister)
	persistersByPath["Epoch_0"] = database.NewMemDB()
	expectedEpoch := uint32(37)
	key := fmt.Sprintf("Epoch_%d", expectedEpoch)
	persistersByPath[key] = database.NewMemDB()
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := database.NewMemDB()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	ps, _ := pruning.NewPruningStorer(args)
	ps.SetEpochForPutOperation(expectedEpoch)

	_ = ps.ChangeEpochSimple(expectedEpoch) // only to add the persister to the map
	_ = ps.ChangeEpochSimple(0)             // set back

	testKey, testVal := []byte("key1"), []byte("val1")
	_ = ps.Put(testKey, testVal) // the persister for epoch 37 exists - will put it there

	ps.ClearCache()

	_, err := ps.GetFromEpoch(testKey, 0)
	assert.NotNil(t, err)

	res, err := ps.GetFromEpoch(testKey, expectedEpoch)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)
}

func TestPruningStorer_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, _ := pruning.NewPruningStorer(args)

	testKey, testVal := []byte("key"), []byte("value")
	err := ps.Put(testKey, testVal)
	assert.Nil(t, err)

	// make sure that the key is there
	res, err := ps.Get(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)

	// now remove it
	err = ps.Remove(testKey)
	assert.Nil(t, err)

	// it should have been removed from the persister and cache
	res, err = ps.Get(testKey)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestPruningStorer_DestroyUnitShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.EpochsData.NumOfEpochsToKeep = 3
	ps, _ := pruning.NewPruningStorer(args)

	// simulate the passing of 2 epochs in order to have more persisters.
	// we will store 3 epochs with 2 active. all 3 should be removed
	_ = ps.ChangeEpochSimple(1)
	_ = ps.ChangeEpochSimple(2)

	err := ps.DestroyUnit()
	assert.Nil(t, err)
}

func TestNewPruningStorer_Has_OnePersisterShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, _ := pruning.NewPruningStorer(args)

	testKey, testVal := []byte("key"), []byte("value")
	err := ps.Put(testKey, testVal)
	assert.Nil(t, err)

	err = ps.Has(testKey)
	assert.Nil(t, err)

	wrongKey := []byte("wrong_key")
	err = ps.Has(wrongKey)
	assert.NotNil(t, err)
}

func TestNewPruningStorer_OldDataHasToBeRemoved(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, _ := pruning.NewPruningStorer(args)

	// add a key and then make 2 epoch changes so the data won't be available anymore
	testKey, _ := json.Marshal([]byte("key"))
	testVal := []byte("value")
	err := ps.Put(testKey, testVal)
	assert.Nil(t, err)

	ps.ClearCache()

	// first check that data is available
	res, err := ps.Get(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)

	// now change the epoch once
	err = ps.ChangeEpochSimple(1)
	assert.Nil(t, err)

	ps.ClearCache()

	// check if data is still available
	res, err = ps.Get(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)

	// now change the epoch again
	err = ps.ChangeEpochSimple(2)
	assert.Nil(t, err)

	ps.ClearCache()

	// data shouldn't be available anymore
	res, err = ps.Get(testKey)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestNewPruningStorer_GetDataFromClosedPersister(t *testing.T) {
	t.Parallel()

	persistersByPath := make(map[string]storage.Persister)
	persistersByPath["Epoch_0"] = database.NewMemDB()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := database.NewMemDB()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.EpochsData.NumOfActivePersisters = 1
	ps, _ := pruning.NewPruningStorer(args)

	// add a key and then make 2 epoch changes so the data won't be available anymore
	testKey, _ := json.Marshal([]byte("key"))
	testVal := []byte("value")
	err := ps.Put(testKey, testVal)
	assert.Nil(t, err)

	ps.ClearCache()

	// first check that data is available
	res, err := ps.Get(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)

	// now change the epoch so the first persister will be closed as only one persister is active at a moment.
	err = ps.ChangeEpochSimple(1)
	assert.Nil(t, err)

	ps.ClearCache()

	// check if data is still available after searching in closed activePersisters
	res, err = ps.GetFromEpoch(testKey, 0)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)
}

func TestNewPruningStorer_GetBulkFromEpoch(t *testing.T) {
	t.Parallel()

	persistersByPath := make(map[string]storage.Persister)
	persistersByPath["Epoch_0"] = database.NewMemDB()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := database.NewMemDB()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.EpochsData.NumOfActivePersisters = 1
	ps, _ := pruning.NewPruningStorer(args)

	// add a key and then make 2 epoch changes so the data won't be available anymore
	testKey1, testKey2 := []byte("key1"), []byte("key2")
	testVal1, testVal2 := []byte("value1"), []byte("value2")
	err := ps.Put(testKey1, testVal1)
	assert.Nil(t, err)
	err = ps.Put(testKey2, testVal2)
	assert.Nil(t, err)

	ps.ClearCache()

	// now change the epoch so the first persister will be closed as only one persister is active at a moment.
	err = ps.ChangeEpochSimple(1)
	assert.Nil(t, err)

	ps.ClearCache()

	// check if data is still available after searching in closed activePersisters
	res, err := ps.GetBulkFromEpoch([][]byte{testKey1, testKey2}, 0)
	assert.Nil(t, err)
	assert.Equal(t, testVal1, res[0].Value)
	assert.Equal(t, testVal2, res[1].Value)
}

func TestNewPruningStorer_ChangeEpochDbsShouldNotBeDeletedIfPruningIsDisabled(t *testing.T) {
	t.Parallel()

	persistersByPath := make(map[string]storage.Persister)
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PruningEnabled = false
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := database.NewMemDB()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.EpochsData.NumOfActivePersisters = 1
	ps, _ := pruning.NewPruningStorer(args)

	// change the epoch multiple times
	_ = ps.ChangeEpochSimple(1)
	_ = ps.ChangeEpochSimple(2)
	_ = ps.ChangeEpochSimple(3)

	assert.Equal(t, 1, len(persistersByPath))
}

func TestNewPruningStorer_ChangeEpochConcurrentPut(t *testing.T) {
	t.Parallel()

	args := getDefaultArgsSerialDB()
	args.DbPath = "TestOnly-Epoch_0"
	args.PruningEnabled = true
	ps, err := pruning.NewPruningStorer(args)
	require.Nil(t, err)
	ps.SetEpochForPutOperation(0)
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		dr := directoryhandler.NewDirectoryReader()
		directories, errList := dr.ListDirectoriesAsString(".")
		assert.NoError(t, errList)
		for _, dir := range directories {
			if strings.HasPrefix(dir, "TestOnly-") {
				errList = os.RemoveAll(dir)
				assert.NoError(t, errList)
			}
		}
	}()

	go func(ctx context.Context) {
		cnt := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err1 := ps.Put([]byte("key"+strconv.Itoa(cnt)), []byte("val"+strconv.Itoa(cnt)))
				require.Nil(t, err1)
				cnt++
				time.Sleep(time.Millisecond)
			}
		}
	}(ctx)

	go func(ctx context.Context) {
		cnt := uint32(1)
		for {
			select {
			case <-ctx.Done():
				_ = ps.DestroyUnit()
				return
			default:
				err2 := ps.ChangeEpochSimple(cnt)
				ps.SetEpochForPutOperation(cnt)
				require.Nil(t, err2)
				cnt++
				time.Sleep(time.Millisecond)
			}
		}
	}(ctx)

	time.Sleep(time.Second * 4)
	cancel()
	time.Sleep(time.Second * 2)
}

func TestPruningStorer_SearchFirst(t *testing.T) {
	t.Parallel()

	persistersByPath := make(map[string]storage.Persister)
	persistersByPath["Epoch_0"] = database.NewMemDB()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := database.NewMemDB()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.EpochsData.NumOfActivePersisters = 3
	args.EpochsData.NumOfEpochsToKeep = 4

	ps, _ := pruning.NewPruningStorer(args)

	// add a key and then make 2 epoch changes so the data won't be available anymore
	testKey, _ := json.Marshal([]byte("key"))
	testVal := []byte("value")
	err := ps.Put(testKey, testVal)
	assert.Nil(t, err)

	ps.ClearCache()

	// check the SearchFirst method works for only one active persister
	res, _ := ps.SearchFirst(testKey)
	assert.Equal(t, testVal, res)

	// now skip 1 epoch and data should still be available
	_ = ps.ChangeEpochSimple(1)
	ps.ClearCache()
	res, _ = ps.SearchFirst(testKey)
	assert.Equal(t, testVal, res)

	// skip one more epoch and data should still be available
	_ = ps.ChangeEpochSimple(2)
	ps.ClearCache()
	res, _ = ps.SearchFirst(testKey)
	assert.Equal(t, testVal, res)

	// when we skip one more epoch, the number of active persisters is exceeded and data shouldn't be available anymore
	_ = ps.ChangeEpochSimple(3)
	ps.ClearCache()
	res, err = ps.SearchFirst(testKey)
	assert.Nil(t, res)
	assert.True(t, errors.Is(err, storage.ErrKeyNotFound))
}

func TestPruningStorer_ChangeEpochWithKeepingFromOldestEpochInMetaBlock(t *testing.T) {
	t.Parallel()

	persistersByPath := make(map[string]storage.Persister)
	persistersByPath["Epoch_0"] = database.NewMemDB()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := database.NewMemDB()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.EpochsData.NumOfActivePersisters = 2
	args.EpochsData.NumOfEpochsToKeep = 4

	ps, _ := pruning.NewPruningStorer(args)
	_ = ps.ChangeEpochSimple(1)
	_ = ps.ChangeEpochSimple(2)
	_ = ps.ChangeEpochSimple(3)

	metaBlock := &block.MetaBlock{
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{{Epoch: 1}},
			Economics:            block.Economics{},
		},
		Epoch: 4,
	}

	err := ps.ChangeEpoch(metaBlock)
	assert.NoError(t, err)

	epochs := ps.GetActivePersistersEpochs()
	assert.Equal(t, 4, len(epochs))

	metaBlock = &block.MetaBlock{
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{{Epoch: 1}},
			Economics:            block.Economics{},
		},
		Epoch: 5,
	}

	err = ps.ChangeEpoch(metaBlock)
	assert.NoError(t, err)

	epochs = ps.GetActivePersistersEpochs()
	assert.Equal(t, 5, len(epochs))

	// the limit of 5 is exceeded, next epoch the num of active persisters will return to 2
	metaBlock = &block.MetaBlock{
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{{Epoch: 1}},
			Economics:            block.Economics{},
		},
		Epoch: 6,
	}

	err = ps.ChangeEpoch(metaBlock)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(ps.GetActivePersistersEpochs()))
}

func TestPruningStorer_ChangeEpochShouldUseMetaBlockFromEpochPrepare(t *testing.T) {
	t.Parallel()

	persistersByPath := make(map[string]storage.Persister)
	persistersByPath["Epoch_0"] = database.NewMemDB()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := database.NewMemDB()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.EpochsData.NumOfActivePersisters = 2
	args.EpochsData.NumOfEpochsToKeep = 4

	ps, _ := pruning.NewPruningStorer(args)
	_ = ps.ChangeEpochSimple(1)
	_ = ps.ChangeEpochSimple(2)
	_ = ps.ChangeEpochSimple(3)
	assert.Equal(t, 2, len(ps.GetActivePersistersEpochs()))

	metaBlock := &block.MetaBlock{
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{{Epoch: 2}},
			Economics:            block.Economics{},
		},
		Epoch: 7,
	}

	err := ps.PrepareChangeEpoch(metaBlock)
	assert.NoError(t, err)

	_ = ps.ChangeEpoch(&block.Header{Epoch: 4})
	assert.Equal(t, 3, len(ps.GetActivePersistersEpochs()))
}

func TestPruningStorer_ChangeEpochWithExisting(t *testing.T) {
	t.Parallel()

	persistersByPath := make(map[string]storage.Persister)
	persistersByPath["Epoch_0/Shard_0/id"] = database.NewMemDB()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := database.NewMemDB()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.EpochsData.NumOfActivePersisters = 2
	args.EpochsData.NumOfEpochsToKeep = 3

	ps, _ := pruning.NewPruningStorer(args)
	key0 := []byte("key_ep0")
	val0 := []byte("value_key_ep0")
	key1 := []byte("key_ep1")
	val1 := []byte("value_key_ep1")
	key2 := []byte("key_ep2")
	val2 := []byte("value_key_ep2")

	err := ps.Put(key0, val0)
	require.Nil(t, err)

	_ = ps.ChangeEpochSimple(1)
	ps.ClearCache()
	err = ps.Put(key1, val1)
	require.Nil(t, err)

	_ = ps.ChangeEpochSimple(2)
	ps.ClearCache()
	err = ps.Put(key2, val2)
	require.Nil(t, err)

	err = ps.ChangeEpochSimple(1)
	require.Nil(t, err)
	ps.ClearCache()

	err = ps.ChangeEpochSimple(1)
	require.Nil(t, err)
	ps.ClearCache()
	restoredVal0, err := ps.Get(key0)
	require.Nil(t, err)
	require.Equal(t, val0, restoredVal0)

	restoredVal1, err := ps.Get(key1)
	require.Nil(t, err)
	require.Equal(t, val1, restoredVal1)

	err = ps.ChangeEpochSimple(2)
	require.Nil(t, err)
	ps.ClearCache()
	restoredVal2, err := ps.Get(key2)
	require.Nil(t, err)
	require.Equal(t, val2, restoredVal2)
}

func TestPruningStorer_ClosePersisters(t *testing.T) {
	t.Parallel()

	t.Run("should remove old databases from map", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		args.OldDataCleanerProvider = &testscommon.OldDataCleanerProviderStub{
			ShouldCleanCalled: func() bool {
				return true
			},
		}
		args.EpochsData.NumOfActivePersisters = 2
		args.EpochsData.NumOfEpochsToKeep = 3

		ps, _ := pruning.NewPruningStorer(args)
		ps.ClearPersisters()

		ps.AddMockActivePersisters([]uint32{0, 1}, true, true)
		err := ps.ClosePersisters(1)
		require.NoError(t, err)
		require.Equal(t, []uint32{0, 1}, ps.PersistersMapByEpochToSlice())

		ps.AddMockActivePersisters([]uint32{2, 3}, true, true)
		err = ps.ClosePersisters(3)
		require.NoError(t, err)
		require.Equal(t, []uint32{1, 2, 3}, ps.PersistersMapByEpochToSlice())

		ps.AddMockActivePersisters([]uint32{4, 5, 6}, true, true)
		err = ps.ClosePersisters(6)
		require.NoError(t, err)
		require.Equal(t, []uint32{4, 5, 6}, ps.PersistersMapByEpochToSlice())
	})

	t.Run("should remove old databases from map + destroy them", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		args.OldDataCleanerProvider = &testscommon.OldDataCleanerProviderStub{
			ShouldCleanCalled: func() bool {
				return true
			},
		}
		args.CustomDatabaseRemover = &testscommon.CustomDatabaseRemoverStub{
			ShouldRemoveCalled: func(_ string, _ uint32) bool {
				return true
			},
		}
		args.EpochsData.NumOfActivePersisters = 2
		args.EpochsData.NumOfEpochsToKeep = 3

		ps, _ := pruning.NewPruningStorer(args)

		destroyClosedWasCalled0 := false
		mockPersister0 := &mock.PersisterStub{
			DestroyClosedCalled: func() error {
				destroyClosedWasCalled0 = true
				return nil
			},
		}
		destroyClosedWasCalled1 := false
		mockPersister1 := &mock.PersisterStub{
			DestroyClosedCalled: func() error {
				destroyClosedWasCalled1 = true
				return nil
			},
		}
		ps.AddMockActivePersister(0, mockPersister0)
		ps.AddMockActivePersister(1, mockPersister1)
		err := ps.ClosePersisters(1)
		require.NoError(t, err)
		require.Equal(t, []uint32{0, 1}, ps.PersistersMapByEpochToSlice())
		require.False(t, destroyClosedWasCalled0)
		require.False(t, destroyClosedWasCalled1)

		destroyClosedWasCalled2 := false
		mockPersister2 := &mock.PersisterStub{
			DestroyClosedCalled: func() error {
				destroyClosedWasCalled2 = true
				return nil
			},
		}
		destroyClosedWasCalled3 := false
		mockPersister3 := &mock.PersisterStub{
			DestroyClosedCalled: func() error {
				destroyClosedWasCalled3 = true
				return nil
			},
		}
		ps.AddMockActivePersister(2, mockPersister2)
		ps.AddMockActivePersister(3, mockPersister3)
		err = ps.ClosePersisters(3)
		require.NoError(t, err)
		require.Equal(t, []uint32{1, 2, 3}, ps.PersistersMapByEpochToSlice())
		require.True(t, destroyClosedWasCalled0)
		require.False(t, destroyClosedWasCalled1)
		require.False(t, destroyClosedWasCalled2)
		require.False(t, destroyClosedWasCalled3)
	})
}

func TestPruningStorer_CleanCustomDatabase(t *testing.T) {
	t.Parallel()

	t.Run("should not destroy old database", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		args.CustomDatabaseRemover = &testscommon.CustomDatabaseRemoverStub{
			ShouldRemoveCalled: func(_ string, _ uint32) bool {
				return false
			},
		}
		args.OldDataCleanerProvider = &testscommon.OldDataCleanerProviderStub{
			ShouldCleanCalled: func() bool {
				return true
			},
		}
	})

}

func TestRegex(t *testing.T) {
	t.Parallel()

	expectedRes := "db/Epoch_7/Shard_2"
	replacementEpoch := "Epoch_7"

	var testPaths []string
	testPaths = append(testPaths, "db/Epoch_22282493984354/Shard_2")
	testPaths = append(testPaths, "db/Epoch_0/Shard_2")
	testPaths = append(testPaths, "db/Epoch_02/Shard_2")
	testPaths = append(testPaths, "db/Epoch_99999999999999999999999999999999999999999999/Shard_2")

	rg := regexp.MustCompile(`Epoch_\d+`)

	for _, path := range testPaths {
		assert.Equal(t, expectedRes, rg.ReplaceAllString(path, replacementEpoch))
	}
}

func TestPruningStorer_processPersistersToClose(t *testing.T) {
	t.Run("edge case - epochs in reversed order", func(t *testing.T) {
		ps := pruning.NewEmptyPruningStorer()
		ps.SetNumActivePersistersParameter(3)
		ps.AddMockActivePersisters([]uint32{10, 9, 8, 7, 6}, false, false)
		persistersToCloseEpochs := ps.ProcessPersistersToClose(8)
		assert.Equal(t, []uint32{7, 6}, persistersToCloseEpochs)
		assert.Equal(t, []uint32{10, 9, 8}, ps.GetActivePersistersEpochs())
		assert.Equal(t, []uint32{6, 7}, ps.PersistersMapByEpochToSlice())
	})

	t.Run("normal operations", func(t *testing.T) {
		ps := pruning.NewEmptyPruningStorer()
		ps.SetNumActivePersistersParameter(3)
		ps.AddMockActivePersisters([]uint32{6, 7, 8, 9, 10}, true, false)
		persistersToCloseEpochs := ps.ProcessPersistersToClose(8)
		assert.Equal(t, []uint32{7, 6}, persistersToCloseEpochs)
		assert.Equal(t, []uint32{10, 9, 8}, ps.GetActivePersistersEpochs())
		assert.Equal(t, []uint32{6, 7}, ps.PersistersMapByEpochToSlice())
	})

	t.Run("normal operations - older last epoch needed", func(t *testing.T) {
		ps := pruning.NewEmptyPruningStorer()
		ps.SetNumActivePersistersParameter(3)
		ps.AddMockActivePersisters([]uint32{6, 7, 8, 9, 10}, true, false)
		persistersToCloseEpochs := ps.ProcessPersistersToClose(6)
		assert.Equal(t, []uint32{}, persistersToCloseEpochs)
		assert.Equal(t, []uint32{10, 9, 8, 7, 6}, ps.GetActivePersistersEpochs())
		assert.Equal(t, []uint32{}, ps.PersistersMapByEpochToSlice())
	})

	t.Run("normal operations - newer last epoch needed", func(t *testing.T) {
		ps := pruning.NewEmptyPruningStorer()
		ps.SetNumActivePersistersParameter(3)
		ps.AddMockActivePersisters([]uint32{6, 7, 8, 9, 10}, true, false)
		persistersToCloseEpochs := ps.ProcessPersistersToClose(10)
		assert.Equal(t, []uint32{7, 6}, persistersToCloseEpochs)
		assert.Equal(t, []uint32{10, 9, 8}, ps.GetActivePersistersEpochs())
		assert.Equal(t, []uint32{6, 7}, ps.PersistersMapByEpochToSlice())
	})
}

func TestPruningStorer_ConcurrentOperations(t *testing.T) {
	numOperations := 100 // increase this to 5000 when troubleshooting pruning storer concurrent operations

	startTime := time.Now()

	_ = logger.SetLogLevel("*:DEBUG")

	dbName := "db-concurrent-test"
	testDir := t.TempDir()

	fmt.Println(testDir)
	args := getDefaultArgs()

	persisterFactory, err := factory.NewPersisterFactory(config.DBConfig{
		FilePath:          filepath.Join(testDir, dbName),
		Type:              "LvlDBSerial",
		MaxBatchSize:      100,
		MaxOpenFiles:      10,
		BatchDelaySeconds: 2,
	})
	require.Nil(t, err)

	args.PersisterFactory = persisterFactory
	args.PathManager, err = pathmanager.NewPathManager(testDir+"/epoch_[E]/shard_[S]/[I]", "shard_[S]/[I]", "db")
	require.NoError(t, err)

	ps, _ := pruning.NewPruningStorer(args)
	require.NotNil(t, ps)
	defer func() {
		_ = ps.Close()
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	rnd := random.ConcurrentSafeIntRandomizer{}
	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	mut := sync.Mutex{}
	currentEpoch := 0

	chanChangeEpoch := make(chan struct{}, 10000)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(ctx context.Context) {
		for {
			select {
			case <-chanChangeEpoch:
				buff := make([]byte, 1)
				_, _ = rand.Read(buff)
				mut.Lock()
				if buff[0] > 200 && currentEpoch > 0 {
					currentEpoch--
				} else {
					currentEpoch++
				}
				newEpoch := currentEpoch
				mut.Unlock()

				_ = ps.ChangeEpochSimple(uint32(newEpoch))
				log.Debug("called ChangeEpochSimple", "epoch", newEpoch)
				wg.Done()
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	numTestedOperations := 7
	for idx := 0; idx < numOperations; idx++ {
		if idx%numTestedOperations != 0 {
			chanChangeEpoch <- struct{}{}
			continue
		}

		go func(index int) {
			time.Sleep(time.Duration(index) * 1 * time.Millisecond)
			switch index % numTestedOperations {
			case 1:
				_, _ = ps.GetFromEpoch([]byte("key"), uint32(index-1))
				log.Debug("called GetFromEpoch", "epoch", index-1)
			case 2:
				_ = ps.Put([]byte("key"), []byte("value"))
				log.Debug("called Put")
			case 3:
				_, _ = ps.Get([]byte("key"))
				log.Debug("called Get")
			case 4:
				epoch := uint32(rnd.Intn(100))
				_, _ = ps.GetFromEpoch([]byte("key"), epoch)
				log.Debug("called GetFromEpoch", "epoch", epoch)
			case 5:
				epoch := uint32(rnd.Intn(100))
				_, _ = ps.GetBulkFromEpoch([][]byte{[]byte("key")}, epoch)
				log.Debug("called GetBulkFromEpoch", "epoch", epoch)
			case 6:
				time.Sleep(time.Millisecond * 10)
				_ = ps.Close()
			}
			wg.Done()
		}(idx)
	}

	wg.Wait()

	log.Info("test done")

	elapsedTime := time.Since(startTime)
	// if the "resource temporary unavailable" occurs, this test will take longer than this to execute
	require.True(t, elapsedTime < 100*time.Second)
}

func TestPruningStorer_RangeKeys(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, _ := pruning.NewPruningStorer(args)

	t.Run("should not panic with nil handler", func(t *testing.T) {
		t.Parallel()

		assert.NotPanics(t, func() {
			ps.RangeKeys(nil)
		})
	})
	t.Run("should not call handler", func(t *testing.T) {
		t.Parallel()

		ps.RangeKeys(func(key []byte, val []byte) bool {
			assert.Fail(t, "should not have called handler")
			return false
		})
	})
}

func TestPruningStorer_GetOldestEpoch(t *testing.T) {
	t.Parallel()

	t.Run("should return error if no persisters are found", func(t *testing.T) {
		t.Parallel()

		epochsData := pruning.EpochArgs{
			NumOfEpochsToKeep:     0,
			NumOfActivePersisters: 0,
		}

		args := getDefaultArgs()
		args.PersistersTracker = pruning.NewPersistersTracker(epochsData)
		ps, _ := pruning.NewPruningStorer(args)

		epoch, err := ps.GetOldestEpoch()
		assert.NotNil(t, err)
		assert.Zero(t, epoch)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		epochsData := pruning.EpochArgs{
			NumOfEpochsToKeep:     2,
			NumOfActivePersisters: 2,
			StartingEpoch:         5,
		}

		args := getDefaultArgs()
		args.PersistersTracker = pruning.NewPersistersTracker(epochsData)
		args.EpochsData = epochsData
		ps, _ := pruning.NewPruningStorer(args)

		epoch, err := ps.GetOldestEpoch()
		assert.Nil(t, err)
		expectedEpoch := uint32(4) // 5 and 4 are the active epochs
		assert.Equal(t, expectedEpoch, epoch)
	})
}

func TestPruningStorer_PutInEpoch(t *testing.T) {
	t.Parallel()

	epochsData := pruning.EpochArgs{
		NumOfEpochsToKeep:     2,
		NumOfActivePersisters: 2,
		StartingEpoch:         5,
	}
	args := getDefaultArgs()
	args.PersistersTracker = pruning.NewPersistersTracker(epochsData)
	args.EpochsData = epochsData
	ps, _ := pruning.NewPruningStorer(args)

	t.Run("if the epoch is not handled, should error", func(t *testing.T) {
		t.Parallel()

		err := ps.PutInEpoch([]byte("key"), []byte("value"), 3) // only 4 and 5 are handled
		expectedErrorString := "put in epoch: persister for epoch 3 not found"
		assert.Equal(t, expectedErrorString, err.Error())
	})
	t.Run("put in existing epochs", func(t *testing.T) {
		t.Parallel()

		key4 := []byte("key4")
		value4 := []byte("value4")
		key5 := []byte("key5")
		value5 := []byte("value5")

		err := ps.PutInEpoch(key4, value4, 4)
		assert.Nil(t, err)

		err = ps.PutInEpoch(key5, value5, 5)
		assert.Nil(t, err)

		t.Run("get from their respective epochs should work", func(t *testing.T) {
			ps.ClearCache()
			recovered4, errGet := ps.GetFromEpoch(key4, 4)
			assert.Nil(t, errGet)
			assert.Equal(t, value4, recovered4)

			ps.ClearCache()
			recovered5, errGet := ps.GetFromEpoch(key5, 5)
			assert.Nil(t, errGet)
			assert.Equal(t, value5, recovered5)
		})
		t.Run("get from wrong epochs should error", func(t *testing.T) {
			ps.ClearCache()
			result, errGet := ps.GetFromEpoch(key4, 3)
			expectedErrorString := fmt.Sprintf("key %x not found in id", key4)
			assert.Equal(t, expectedErrorString, errGet.Error())
			assert.Nil(t, result)

			ps.ClearCache()
			result, errGet = ps.GetFromEpoch(key4, 5)
			expectedErrorString = fmt.Sprintf("key %x not found in id", key4)
			assert.Equal(t, expectedErrorString, errGet.Error())
			assert.Nil(t, result)
		})
	})
}

func TestPruningStorer_RemoveFromCurrentEpoch(t *testing.T) {
	t.Parallel()

	epochsData := pruning.EpochArgs{
		NumOfEpochsToKeep:     2,
		NumOfActivePersisters: 2,
		StartingEpoch:         5,
	}
	args := getDefaultArgs()
	args.PersistersTracker = pruning.NewPersistersTracker(epochsData)
	args.EpochsData = epochsData
	ps, _ := pruning.NewPruningStorer(args)

	// current epoch is 5
	key := []byte("key")
	value := []byte("value")

	// put in epoch 4
	_ = ps.PutInEpoch(key, value, 4)
	// put in epoch 5
	_ = ps.PutInEpoch(key, value, 5)

	// remove from epoch 5
	err := ps.RemoveFromCurrentEpoch(key)
	assert.Nil(t, err)

	// get from epoch 5 should error
	ps.ClearCache()
	result, errGet := ps.GetFromEpoch(key, 5)
	expectedErrorString := fmt.Sprintf("key %x not found in id", key)
	assert.Equal(t, expectedErrorString, errGet.Error())
	assert.Nil(t, result)

	// get from epoch 4 should work
	ps.ClearCache()
	recovered, errGet := ps.GetFromEpoch(key, 4)
	assert.Nil(t, errGet)
	assert.Equal(t, value, recovered)
}

func TestPruningStorer_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ps *pruning.PruningStorer
	require.True(t, ps.IsInterfaceNil())

	args := getDefaultArgs()
	ps, _ = pruning.NewPruningStorer(args)
	require.False(t, ps.IsInterfaceNil())
}
