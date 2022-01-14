package pruning_test

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTriePruningStorer_GetFromOldEpochsWithoutCacheSearchesOnlyOldEpochs(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, _ := pruning.NewTriePruningStorer(args)
	cacher := testscommon.NewCacherMock()
	ps.SetCacher(cacher)

	testKey1 := []byte("key1")
	testVal1 := []byte("value1")
	testKey2 := []byte("key2")
	testVal2 := []byte("value2")

	err := ps.PutInEpochWithoutCache(testKey1, testVal1, 0)
	assert.Nil(t, err)

	err = ps.ChangeEpochSimple(1)
	assert.Nil(t, err)
	ps.SetEpochForPutOperation(1)

	err = ps.PutInEpochWithoutCache(testKey2, testVal2, 1)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(cacher.Keys()))

	res, err := ps.GetFromOldEpochsWithoutAddingToCache(testKey1)
	assert.Equal(t, testVal1, res)
	assert.Nil(t, err)

	res, err = ps.GetFromOldEpochsWithoutAddingToCache(testKey2)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestTriePruningStorer_GetFromOldEpochsWithoutCacheLessActivePersisters(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.NumOfActivePersisters = 2
	ps, _ := pruning.NewTriePruningStorer(args)

	testKey1 := []byte("key1")
	testVal1 := []byte("value1")

	err := ps.PutInEpochWithoutCache(testKey1, testVal1, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, ps.GetNumActivePersisters())
	_ = ps.ChangeEpochSimple(1)

	val, err := ps.GetFromOldEpochsWithoutAddingToCache(testKey1)
	assert.Nil(t, err)
	assert.Equal(t, testVal1, val)
}

func TestTriePruningStorer_GetFromOldEpochsWithoutCacheMoreActivePersisters(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.NumOfActivePersisters = 2
	args.NumOfEpochsToKeep = 4
	ps, _ := pruning.NewTriePruningStorer(args)

	testKey1 := []byte("key1")
	testVal1 := []byte("value1")

	err := ps.PutInEpochWithoutCache(testKey1, testVal1, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, ps.GetNumActivePersisters())
	_ = ps.ChangeEpochSimple(1)
	_ = ps.ChangeEpochSimple(2)
	_ = ps.ChangeEpochSimple(3)

	val, err := ps.GetFromOldEpochsWithoutAddingToCache(testKey1)
	assert.Nil(t, err)
	assert.Equal(t, testVal1, val)
}

func TestTriePruningStorer_GetFromOldEpochsWithoutCacheAllPersistersClosed(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.NumOfActivePersisters = 2
	args.NumOfEpochsToKeep = 4

	persistersMap := make(map[string]storage.Persister)
	persisterFactory := &mock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			persister, exists := persistersMap[path]
			if !exists {
				persister = &mock.PersisterStub{
					GetCalled: func(key []byte) ([]byte, error) {
						return nil, storage.ErrSerialDBIsClosed
					},
				}
				persistersMap[path] = persister
			}

			return persister, nil
		},
	}
	args.PersisterFactory = persisterFactory
	ps, _ := pruning.NewTriePruningStorer(args)

	_ = ps.ChangeEpochSimple(1)
	_ = ps.ChangeEpochSimple(2)
	_ = ps.ChangeEpochSimple(3)
	_ = ps.Close()

	val, err := ps.GetFromOldEpochsWithoutAddingToCache([]byte("key"))
	assert.Nil(t, val)
	assert.Equal(t, storage.ErrSerialDBIsClosed, err)
}

func TestTriePruningStorer_GetFromOldEpochsWithoutCacheDoesNotSearchInCurrentStorer(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, _ := pruning.NewTriePruningStorer(args)
	cacher := testscommon.NewCacherStub()
	cacher.PutCalled = func(_ []byte, _ interface{}, _ int) bool {
		require.Fail(t, "this should not be called")
		return false
	}
	ps.SetCacher(cacher)
	testKey1 := []byte("key1")
	testVal1 := []byte("value1")

	err := ps.PutInEpochWithoutCache(testKey1, testVal1, 0)
	assert.Nil(t, err)
	ps.ClearCache()

	res, err := ps.GetFromOldEpochsWithoutAddingToCache(testKey1)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestTriePruningStorer_GetFromLastEpochSearchesOnlyLastEpoch(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, _ := pruning.NewTriePruningStorer(args)
	cacher := testscommon.NewCacherMock()
	ps.SetCacher(cacher)

	testKey1 := []byte("key1")
	testVal1 := []byte("value1")
	testKey2 := []byte("key2")
	testVal2 := []byte("value2")
	testKey3 := []byte("key3")
	testVal3 := []byte("value3")

	err := ps.PutInEpochWithoutCache(testKey1, testVal1, 0)
	assert.Nil(t, err)

	err = ps.ChangeEpochSimple(1)
	assert.Nil(t, err)
	ps.SetEpochForPutOperation(1)

	err = ps.PutInEpochWithoutCache(testKey2, testVal2, 1)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(cacher.Keys()))

	err = ps.ChangeEpochSimple(2)
	assert.Nil(t, err)
	ps.SetEpochForPutOperation(2)

	err = ps.PutInEpochWithoutCache(testKey3, testVal3, 2)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(cacher.Keys()))

	res, err := ps.GetFromLastEpoch(testKey2)
	assert.Equal(t, testVal2, res)
	assert.Nil(t, err)

	res, err = ps.GetFromLastEpoch(testKey1)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))

	res, err = ps.GetFromLastEpoch(testKey3)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestTriePruningStorer_GetFromCurrentEpochSearchesOnlyCurrentEpoch(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, _ := pruning.NewTriePruningStorer(args)
	cacher := testscommon.NewCacherMock()
	ps.SetCacher(cacher)

	testKey1 := []byte("key1")
	testVal1 := []byte("value1")
	testKey2 := []byte("key2")
	testVal2 := []byte("value2")
	testKey3 := []byte("key3")
	testVal3 := []byte("value3")

	err := ps.PutInEpochWithoutCache(testKey1, testVal1, 0)
	assert.Nil(t, err)

	err = ps.ChangeEpochSimple(1)
	assert.Nil(t, err)
	ps.SetEpochForPutOperation(1)

	err = ps.PutInEpochWithoutCache(testKey2, testVal2, 1)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(cacher.Keys()))

	err = ps.ChangeEpochSimple(2)
	assert.Nil(t, err)
	ps.SetEpochForPutOperation(2)

	err = ps.PutInEpochWithoutCache(testKey3, testVal3, 2)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(cacher.Keys()))

	res, err := ps.GetFromCurrentEpoch(testKey3)
	assert.Equal(t, testVal3, res)
	assert.Nil(t, err)

	res, err = ps.GetFromCurrentEpoch(testKey1)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))

	res, err = ps.GetFromCurrentEpoch(testKey2)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestTriePruningStorer_OpenMoreDbsIfNecessary(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	tps, _ := pruning.NewTriePruningStorer(args)

	_ = tps.ChangeEpochSimple(1)

	tps.SetEpochForPutOperation(1)
	err := tps.Put([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal))
	assert.Nil(t, err)

	_ = tps.ChangeEpochSimple(2)
	_ = tps.ChangeEpochSimple(3)
	_ = tps.ChangeEpochSimple(4)

	err = tps.Close()
	assert.Nil(t, err)

	args.StartingEpoch = 4
	args.NumOfEpochsToKeep = 5
	ps, _ := pruning.NewPruningStorer(args)
	assert.Equal(t, 2, ps.GetNumActivePersisters())
	tps, _ = pruning.NewTriePruningStorer(args)
	assert.Equal(t, 4, tps.GetNumActivePersisters())
}

func TestTriePruningStorer_KeepMoreDbsOpenIfNecessary(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	tps, _ := pruning.NewTriePruningStorer(args)

	assert.Equal(t, 1, tps.GetNumActivePersisters())
	_ = tps.ChangeEpochSimple(1)

	tps.SetEpochForPutOperation(1)
	err := tps.Put([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal))
	assert.Nil(t, err)

	assert.Equal(t, 2, tps.GetNumActivePersisters())
	_ = tps.ChangeEpochSimple(2)
	assert.Equal(t, 2, tps.GetNumActivePersisters())
	_ = tps.ChangeEpochSimple(3)
	assert.Equal(t, 3, tps.GetNumActivePersisters())
	_ = tps.ChangeEpochSimple(4)
	assert.Equal(t, 4, tps.GetNumActivePersisters())

	tps.SetEpochForPutOperation(4)
	err = tps.Put([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal))
	assert.Nil(t, err)

	_ = tps.ChangeEpochSimple(5)
	assert.Equal(t, 2, tps.GetNumActivePersisters())

	err = tps.Close()
	assert.Nil(t, err)
}
