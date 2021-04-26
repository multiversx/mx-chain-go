package pruning_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/leveldb"
	"github.com/stretchr/testify/require"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func getDummyConfig() (storageUnit.CacheConfig, storageUnit.DBConfig, storageUnit.BloomConfig) {
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
	blConf := storageUnit.BloomConfig{}
	return cacheConf, dbConf, blConf
}

func getDefaultArgs() *pruning.StorerArgs {
	cacheConf, dbConf, blConf := getDummyConfig()

	lockPersisterMap := sync.Mutex{}
	persistersMap := make(map[string]storage.Persister)
	persisterFactory := &mock.PersisterFactoryStub{
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
	return &pruning.StorerArgs{
		PruningEnabled:        true,
		Identifier:            "id",
		CleanOldEpochsData:    false,
		ShardCoordinator:      mock.NewShardCoordinatorMock(0, 2),
		PathManager:           &mock.PathManagerStub{},
		CacheConf:             cacheConf,
		DbPath:                dbConf.FilePath,
		PersisterFactory:      persisterFactory,
		BloomFilterConf:       blConf,
		NumOfEpochsToKeep:     2,
		NumOfActivePersisters: 2,
		Notifier:              &mock.EpochStartNotifierStub{},
		MaxBatchSize:          10,
	}
}

func getDefaultArgsSerialDB() *pruning.StorerArgs {
	cacheConf, dbConf, blConf := getDummyConfig()
	cacheConf.Capacity = 40
	persisterFactory := &mock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			return leveldb.NewSerialDB(path, 1, 20, 10)
		},
	}
	pathManager := &mock.PathManagerStub{PathForEpochCalled: func(shardId string, epoch uint32, identifier string) string {
		return fmt.Sprintf("TestOnly-Epoch_%d/Shard_%s/%s", epoch, shardId, identifier)
	}}
	return &pruning.StorerArgs{
		PruningEnabled:        true,
		Identifier:            "id",
		CleanOldEpochsData:    false,
		ShardCoordinator:      mock.NewShardCoordinatorMock(0, 2),
		PathManager:           pathManager,
		CacheConf:             cacheConf,
		DbPath:                dbConf.FilePath,
		PersisterFactory:      persisterFactory,
		BloomFilterConf:       blConf,
		NumOfEpochsToKeep:     3,
		NumOfActivePersisters: 2,
		Notifier:              &mock.EpochStartNotifierStub{},
		MaxBatchSize:          20,
	}
}

func TestNewPruningStorer_InvalidNumberOfActivePersistersShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.NumOfActivePersisters = 0

	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrInvalidNumberOfPersisters, err)
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
	persistersByPath["Epoch_0"] = memorydb.New()
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := memorydb.New()
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
	persistersByPath["Epoch_0"] = memorydb.New()
	expectedEpoch := uint32(37)
	key := fmt.Sprintf("Epoch_%d", expectedEpoch)
	persistersByPath[key] = memorydb.New()
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := memorydb.New()
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
	args.NumOfEpochsToKeep = 3
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
	persistersByPath["Epoch_0"] = memorydb.New()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := memorydb.New()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.NumOfActivePersisters = 1
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
	persistersByPath["Epoch_0"] = memorydb.New()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := memorydb.New()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.NumOfActivePersisters = 1
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
	assert.Equal(t, testVal1, res[string(testKey1)])
	assert.Equal(t, testVal2, res[string(testKey2)])
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
			newPers := memorydb.New()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.NumOfActivePersisters = 1
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
		dr := factory.NewDirectoryReader()
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
				fmt.Println("destroy called")
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
	persistersByPath["Epoch_0"] = memorydb.New()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := memorydb.New()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.NumOfActivePersisters = 3
	args.NumOfEpochsToKeep = 4

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
	persistersByPath["Epoch_0"] = memorydb.New()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := memorydb.New()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.NumOfActivePersisters = 2
	args.NumOfEpochsToKeep = 4

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
	persistersByPath["Epoch_0"] = memorydb.New()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := memorydb.New()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.NumOfActivePersisters = 2
	args.NumOfEpochsToKeep = 4

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
	persistersByPath["Epoch_0/Shard_0/id"] = memorydb.New()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving activePersisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers := memorydb.New()
			persistersByPath[path] = newPers

			return newPers, nil
		},
	}
	args.NumOfActivePersisters = 2
	args.NumOfEpochsToKeep = 3

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
	restauredVal0, err := ps.Get(key0)
	require.Nil(t, err)
	require.Equal(t, val0, restauredVal0)

	restauredVal1, err := ps.Get(key1)
	require.Nil(t, err)
	require.Equal(t, val1, restauredVal1)

	err = ps.ChangeEpochSimple(2)
	require.Nil(t, err)
	ps.ClearCache()
	restauredVal2, err := ps.Get(key2)
	require.Nil(t, err)
	require.Equal(t, val2, restauredVal2)
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

func TestDirectories(t *testing.T) {
	pathToCreate := "user-directory/go/src/workspace/db/Epoch_2/Shard_27"
	pathParameter := pathToCreate + "/MiniBlock"
	// should become user-directory/go/src/workspace/db

	err := os.MkdirAll(pathToCreate, os.ModePerm)
	assert.Nil(t, err)

	pruning.RemoveDirectoryIfEmpty(pathParameter)

	if _, err = os.Stat(pathParameter); !os.IsNotExist(err) {
		assert.Fail(t, "directory should have been removed")
	}

	_ = os.RemoveAll("user-directory")
}
