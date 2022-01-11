package pruning_test

import (
	"strings"
	"testing"

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
