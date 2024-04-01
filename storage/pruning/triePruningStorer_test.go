package pruning_test

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTriePruningStorer(t *testing.T) {
	t.Parallel()

	t.Run("empty args struct, should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			require.Nil(t, r)
		}()
		emptyAndInvalidConfig := pruning.StorerArgs{}
		tps, err := pruning.NewTriePruningStorer(emptyAndInvalidConfig)
		require.Error(t, err)
		require.Nil(t, tps)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgs()
		ps, err := pruning.NewTriePruningStorer(args)
		require.NoError(t, err)
		require.NotNil(t, ps)
	})
}

func TestTriePruningStorer_GetFromOldEpochsWithoutCacheSearchesOnlyOldEpochsAndReturnsEpoch(t *testing.T) {
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

	res, epoch, err := ps.GetFromOldEpochsWithoutAddingToCache(testKey1)
	assert.Equal(t, testVal1, res)
	assert.Nil(t, err)
	assert.True(t, epoch.HasValue)
	assert.Equal(t, uint32(0), epoch.Value)

	res, epoch, err = ps.GetFromOldEpochsWithoutAddingToCache(testKey2)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.False(t, epoch.HasValue)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestTriePruningStorer_GetFromOldEpochsWithCache(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	ps, _ := pruning.NewTriePruningStorer(args)
	cacher := testscommon.NewCacherMock()
	ps.SetCacher(cacher)

	testKey1 := []byte("key1")
	testVal1 := []byte("value1")

	err := ps.PutInEpoch(testKey1, testVal1, 0)
	assert.Nil(t, err)

	err = ps.ChangeEpochSimple(1)
	assert.Nil(t, err)
	ps.SetEpochForPutOperation(1)

	res, epoch, err := ps.GetFromOldEpochsWithoutAddingToCache(testKey1)
	assert.Equal(t, testVal1, res)
	assert.Nil(t, err)
	assert.False(t, epoch.HasValue)
	assert.Equal(t, uint32(0), epoch.Value)
}

func TestTriePruningStorer_GetFromOldEpochsWithoutCacheLessActivePersisters(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.EpochsData.NumOfActivePersisters = 2
	ps, _ := pruning.NewTriePruningStorer(args)

	testKey1 := []byte("key1")
	testVal1 := []byte("value1")

	err := ps.PutInEpochWithoutCache(testKey1, testVal1, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, ps.GetNumActivePersisters())
	_ = ps.ChangeEpochSimple(1)

	val, _, err := ps.GetFromOldEpochsWithoutAddingToCache(testKey1)
	assert.Nil(t, err)
	assert.Equal(t, testVal1, val)
}

func TestTriePruningStorer_GetFromOldEpochsWithoutCacheMoreActivePersisters(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.EpochsData.NumOfActivePersisters = 2
	args.EpochsData.NumOfEpochsToKeep = 4
	ps, _ := pruning.NewTriePruningStorer(args)

	testKey1 := []byte("key1")
	testVal1 := []byte("value1")

	err := ps.PutInEpochWithoutCache(testKey1, testVal1, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, ps.GetNumActivePersisters())
	_ = ps.ChangeEpochSimple(1)
	_ = ps.ChangeEpochSimple(2)
	_ = ps.ChangeEpochSimple(3)

	val, _, err := ps.GetFromOldEpochsWithoutAddingToCache(testKey1)
	assert.Nil(t, err)
	assert.Equal(t, testVal1, val)
}

func TestTriePruningStorer_GetFromOldEpochsWithoutCacheAllPersistersClosed(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.EpochsData.NumOfActivePersisters = 2
	args.EpochsData.NumOfEpochsToKeep = 4

	persistersMap := make(map[string]storage.Persister)
	persisterFactory := &mock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			persister, exists := persistersMap[path]
			if !exists {
				persister = &mock.PersisterStub{
					GetCalled: func(key []byte) ([]byte, error) {
						return nil, storage.ErrDBIsClosed
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

	val, _, err := ps.GetFromOldEpochsWithoutAddingToCache([]byte("key"))
	assert.Nil(t, val)
	assert.Equal(t, storage.ErrDBIsClosed, err)
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

	res, _, err := ps.GetFromOldEpochsWithoutAddingToCache(testKey1)
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

	tps.SetEpochForPutOperation(2)
	err = tps.Put([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal))
	assert.Nil(t, err)

	_ = tps.ChangeEpochSimple(3)
	_ = tps.ChangeEpochSimple(4)

	err = tps.Close()
	assert.Nil(t, err)

	args.EpochsData.StartingEpoch = 4
	args.EpochsData.NumOfEpochsToKeep = 5
	args.PersistersTracker = pruning.NewPersistersTracker(args.EpochsData)
	ps, _ := pruning.NewPruningStorer(args)
	assert.Equal(t, 2, ps.GetNumActivePersisters())
	args.PersistersTracker = pruning.NewTriePersisterTracker(args.EpochsData)
	tps, _ = pruning.NewTriePruningStorer(args)
	assert.Equal(t, 4, tps.GetNumActivePersisters())
}

func TestTriePruningStorer_KeepMoreDbsOpenIfNecessary(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	args.EpochsData.NumOfActivePersisters = 3
	args.EpochsData.NumOfEpochsToKeep = 3
	tps, _ := pruning.NewTriePruningStorer(args)

	assert.Equal(t, 1, tps.GetNumActivePersisters())
	_ = tps.ChangeEpochSimple(1)

	tps.SetEpochForPutOperation(1)
	err := tps.Put([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal))
	assert.Nil(t, err)

	assert.Equal(t, 2, tps.GetNumActivePersisters())
	_ = tps.ChangeEpochSimple(2)
	assert.Equal(t, 3, tps.GetNumActivePersisters())
	_ = tps.ChangeEpochSimple(3)
	assert.Equal(t, 4, tps.GetNumActivePersisters())
	_ = tps.ChangeEpochSimple(4)
	assert.Equal(t, 5, tps.GetNumActivePersisters())

	tps.SetEpochForPutOperation(4)
	err = tps.Put([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal))
	assert.Nil(t, err)

	_ = tps.ChangeEpochSimple(5)
	tps.SetEpochForPutOperation(5)
	err = tps.Put([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal))
	assert.Nil(t, err)

	_ = tps.ChangeEpochSimple(6)
	assert.Equal(t, 3, tps.GetNumActivePersisters())

	err = tps.Close()
	assert.Nil(t, err)
}

func TestTriePruningStorer_GetLatestStorageEpoch(t *testing.T) {
	t.Parallel()

	epochsData := pruning.EpochArgs{
		NumOfEpochsToKeep:     2,
		NumOfActivePersisters: 2,
		StartingEpoch:         5,
	}
	args := getDefaultArgs()
	args.PersistersTracker = pruning.NewPersistersTracker(epochsData)

	t.Run("no active db should error", func(t *testing.T) {
		t.Parallel()

		tps, _ := pruning.NewTriePruningStorer(args)
		latestEpoch, err := tps.GetLatestStorageEpoch()
		expectedErrString := "there are no active persisters"
		assert.Equal(t, expectedErrString, err.Error())
		assert.Zero(t, latestEpoch)
	})
	t.Run("with at least one active DB should work", func(t *testing.T) {
		t.Parallel()

		tps, _ := pruning.NewTriePruningStorer(args)
		_ = tps.ChangeEpochSimple(5)
		latestEpoch, err := tps.GetLatestStorageEpoch()
		assert.Nil(t, err)
		assert.Equal(t, uint32(5), latestEpoch)
	})
}

func TestTriePruningStorer_RemoveFromAllActiveEpochs(t *testing.T) {
	t.Parallel()

	t.Run("no active persisters returns nil err", func(t *testing.T) {
		t.Parallel()

		epochsData := pruning.EpochArgs{
			NumOfEpochsToKeep:     2,
			NumOfActivePersisters: 2,
			StartingEpoch:         5,
		}
		args := getDefaultArgs()
		args.PersistersTracker = pruning.NewPersistersTracker(epochsData)

		tps, _ := pruning.NewTriePruningStorer(args)
		err := tps.RemoveFromAllActiveEpochs([]byte("test"))
		assert.Nil(t, err)
	})

	t.Run("key not present returns nil err", func(t *testing.T) {
		args := getDefaultArgs()

		tps, _ := pruning.NewTriePruningStorer(args)
		_ = tps.ChangeEpochSimple(3)

		err := tps.RemoveFromAllActiveEpochs([]byte("test"))
		assert.Nil(t, err)
	})

	t.Run("key is removed from all persisters", func(t *testing.T) {
		args := getDefaultArgs()

		tps, _ := pruning.NewTriePruningStorer(args)
		_ = tps.ChangeEpochSimple(1)

		key := []byte("test")
		value := []byte("test data")
		_ = tps.PutInEpoch(key, value, 0)
		_ = tps.PutInEpoch(key, value, 1)

		val, err := tps.GetFromEpoch(key, 0)
		assert.Nil(t, err)
		assert.Equal(t, value, val)

		val, err = tps.GetFromEpoch(key, 1)
		assert.Nil(t, err)
		assert.Equal(t, value, val)

		err = tps.RemoveFromAllActiveEpochs([]byte("test"))
		assert.Nil(t, err)

		val, err = tps.GetFromEpoch(key, 0)
		assert.Nil(t, val)
		assert.NotNil(t, err)

		val, err = tps.GetFromEpoch(key, 1)
		assert.Nil(t, val)
		assert.NotNil(t, err)
	})
}

func TestTriePruningStorer_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	tps, _ := pruning.NewTriePruningStorer(pruning.StorerArgs{})
	require.True(t, tps.IsInterfaceNil())

	args := getDefaultArgs()
	tps, _ = pruning.NewTriePruningStorer(args)
	require.False(t, tps.IsInterfaceNil())
}
