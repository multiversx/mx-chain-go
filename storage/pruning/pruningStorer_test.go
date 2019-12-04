package pruning_test

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func getDummyConfig() (storageUnit.CacheConfig, storageUnit.DBConfig, storageUnit.BloomConfig) {
	cacheConf := storageUnit.CacheConfig{
		Size:   10,
		Type:   "LRU",
		Shards: 3,
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

func getDefaultArgs() *pruning.PruningStorerArgs {
	cacheConf, dbConf, blConf := getDummyConfig()
	persisterFactory := &mock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			return memorydb.New()
		},
	}
	return &pruning.PruningStorerArgs{
		Identifier:            "id",
		FullArchive:           false,
		CacheConf:             cacheConf,
		DbPath:                dbConf.FilePath,
		PersisterFactory:      persisterFactory,
		BloomFilterConf:       blConf,
		NumOfEpochsToKeep:     2,
		NumOfActivePersisters: 2,
		Notifier:              &mock.EpochStartNotifierStub{},
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

func TestNewPruningStorer_NilPersisterFactoryShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.PersisterFactory = nil
	ps, err := pruning.NewPruningStorer(args)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrNilPersisterFactory, err)
}

func TestNewPruningStorer_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, err := pruning.NewPruningStorer(args)

	assert.NotNil(t, ps)
	assert.Nil(t, err)
	assert.False(t, ps.IsInterfaceNil())
}

func TestNewPruningStorer_PutAndGetShouldWork(t *testing.T) {
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

func TestNewPruningStorer_Has_MultiplePersistersShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.NumOfActivePersisters = 1
	args.NumOfEpochsToKeep = 2
	ps, _ := pruning.NewPruningStorer(args)

	testKey, testVal := []byte("key"), []byte("value")
	err := ps.Put(testKey, testVal)
	assert.Nil(t, err)

	ps.ClearCache()
	err = ps.Has(testKey)
	assert.Nil(t, err)

	_ = ps.ChangeEpoch(1)
	ps.ClearCache()

	// data should still be available in the closed persister
	err = ps.Has(testKey)
	assert.Nil(t, err)

	// after one more epoch change, the persister which holds the data should be removed and the key should not be available
	_ = ps.ChangeEpoch(2)
	ps.ClearCache()

	err = ps.Has(testKey)
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
	err = ps.ChangeEpoch(1)
	assert.Nil(t, err)

	ps.ClearCache()

	// check if data is still available
	res, err = ps.Get(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)

	// now change the epoch again
	err = ps.ChangeEpoch(2)
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
	persistersByPath["Epoch_0"], _ = memorydb.New()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.PersisterFactory = &mock.PersisterFactoryStub{
		// simulate an opening of an existing database from the file path by saving persisters in a map based on their path
		CreateCalled: func(path string) (storage.Persister, error) {
			if _, ok := persistersByPath[path]; ok {
				return persistersByPath[path], nil
			}
			newPers, _ := memorydb.New()
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
	err = ps.ChangeEpoch(1)
	assert.Nil(t, err)

	ps.ClearCache()

	// check if data is still available after searching in closed persisters
	res, err = ps.Get(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)
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

	if _, err := os.Stat(pathParameter); !os.IsNotExist(err) {
		assert.Fail(t, "directory should have been removed")
	}

	_ = os.RemoveAll("user-directory")
}
