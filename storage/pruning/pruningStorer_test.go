package pruning_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage"
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
		BatchDelaySeconds: 10,
		MaxBatchSize:      100,
		MaxOpenFiles:      1000,
	}
	blConf := storageUnit.BloomConfig{}
	return cacheConf, dbConf, blConf
}

func TestNewPruningStorer_InvalidNumberOfActivePersistersShouldErr(t *testing.T) {
	t.Parallel()

	cacheConf, dbConf, blConf := getDummyConfig()
	ps, err := pruning.NewPruningStorer(
		"id",
		true,
		cacheConf,
		dbConf,
		blConf,
		0,
		&mock.EpochStartNotifierStub{},
	)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrInvalidNumberOfPersisters, err)
}

func TestNewPruningStorer_NilEpochStartHandlerShouldErr(t *testing.T) {
	t.Parallel()

	cacheConf, dbConf, blConf := getDummyConfig()
	ps, err := pruning.NewPruningStorer(
		"id",
		true,
		cacheConf,
		dbConf,
		blConf,
		2,
		nil,
	)

	assert.Nil(t, ps)
	assert.Equal(t, storage.ErrNilEpochStartNotifier, err)
}

func TestNewPruningStorer_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cacheConf, dbConf, blConf := getDummyConfig()
	ps, err := pruning.NewPruningStorer(
		"id",
		true,
		cacheConf,
		dbConf,
		blConf,
		2,
		&mock.EpochStartNotifierStub{},
	)

	assert.NotNil(t, ps)
	assert.Nil(t, err)
}

func TestNewPruningStorer_PutAndGetShouldWork(t *testing.T) {
	t.Parallel()

	cacheConf, dbConf, blConf := getDummyConfig()
	ps, _ := pruning.NewPruningStorer(
		"id",
		true,
		cacheConf,
		dbConf,
		blConf,
		2,
		&mock.EpochStartNotifierStub{},
	)

	testKey, testVal := []byte("key"), []byte("value")
	err := ps.Put(testKey, testVal)
	assert.Nil(t, err)

	res, err := ps.Get(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)
}

func TestNewPruningStorer_OldDataHasToBeRemoved(t *testing.T) {
	t.Parallel()

	cacheConf, dbConf, blConf := getDummyConfig()
	ps, _ := pruning.NewPruningStorer(
		"id",
		false,
		cacheConf,
		dbConf,
		blConf,
		2,
		&mock.EpochStartNotifierStub{},
	)

	// add a key and then make 2 epoch changes so the data won't be available anymore
	testKey, testVal := []byte("key"), []byte("value")
	err := ps.Put(testKey, testVal)
	assert.Nil(t, err)

	// first check that data is available
	res, err := ps.Get(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)

	// now change the epoch once
	err = ps.ChangeEpoch(1)
	assert.Nil(t, err)

	// check if data is still available
	res, err = ps.Get(testKey)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)

	// now change the epoch again
	err = ps.ChangeEpoch(2)
	assert.Nil(t, err)

	// clear cache because it will break the data availability in persisters purpose of this test
	ps.ClearCache()

	// data shouldn't be available anymore
	res, err = ps.Get(testKey)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
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

	rg := regexp.MustCompile("Epoch_\\d+")

	for _, path := range testPaths {
		assert.Equal(t, expectedRes, rg.ReplaceAllString(path, replacementEpoch))
	}
}
