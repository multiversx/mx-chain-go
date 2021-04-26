package pruning_test

import (
	"math"
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFullHistoryPruningStorer_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := &pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 10,
	}
	fhps, err := pruning.NewFullHistoryPruningStorer(fhArgs)

	assert.NotNil(t, fhps)
	assert.Nil(t, err)
	assert.False(t, fhps.IsInterfaceNil())
}

func TestNewFullHistoryPruningStorer_InvalidNumberOfActivePersistersShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()

	fhArgs := &pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 0,
	}
	fhps, err := pruning.NewFullHistoryPruningStorer(fhArgs)

	assert.Nil(t, fhps)
	assert.Equal(t, storage.ErrInvalidNumberOfOldPersisters, err)

	fhArgs = &pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: math.MaxInt32 + 1,
	}
	fhps, err = pruning.NewFullHistoryPruningStorer(fhArgs)

	assert.Nil(t, fhps)
	assert.Equal(t, storage.ErrInvalidNumberOfOldPersisters, err)
}

func TestNewFullHistoryPruningStorer_PutAndGetInEpochShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := &pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 2,
	}
	fhps, _ := pruning.NewFullHistoryPruningStorer(fhArgs)

	testKey, testVal := []byte("key"), []byte("value")
	// init persister for epoch 7
	_, _ = fhps.GetFromEpoch(testKey, 7)

	err := fhps.PutInEpoch(testKey, testVal, 7)
	assert.Nil(t, err)

	res, err := fhps.GetFromEpoch(testKey, 7)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)
}

func TestNewFullHistoryPruningStorer_GetMultipleDifferentEpochsShouldEvict(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := &pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 3,
	}
	fhps, _ := pruning.NewFullHistoryPruningStorer(fhArgs)

	testKey := []byte("key")
	testEpoch := []byte("7")
	testEpochNext := []byte("8")
	// init persister for epoch 7 and 8
	_, _ = fhps.GetFromEpoch(testKey, 7)
	ok := fhps.GetOldEpochsActivePersisters().Has(testEpoch)
	assert.True(t, ok)

	ok = fhps.GetOldEpochsActivePersisters().Has(testEpochNext)
	assert.True(t, ok)

	// init persister for epoch 9
	_, _ = fhps.GetFromEpoch(testKey, 9)
	ok = fhps.GetOldEpochsActivePersisters().Has(testEpoch)
	assert.False(t, ok)
}

func TestNewFullHistoryPruningStorer_GetAfterEvictShouldWork(t *testing.T) {
	t.Parallel()

	persistersByPath := make(map[string]storage.Persister)
	persistersByPath["Epoch_0"] = memorydb.New()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.NumOfActivePersisters = 1
	args.CleanOldEpochsData = true
	args.NumOfEpochsToKeep = 2

	fhArgs := &pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 3,
	}
	fhps, _ := pruning.NewFullHistoryPruningStorer(fhArgs)
	testVal := []byte("value")
	testKey := []byte("key")
	testEpochKey := []byte("7")
	testEpoch := uint32(7)
	// init persister for epoch 7 and 8
	_, _ = fhps.GetFromEpoch(nil, testEpoch)
	ok := fhps.GetOldEpochsActivePersisters().Has(testEpochKey)
	assert.True(t, ok)

	err := fhps.PutInEpoch(testKey, testVal, testEpoch)
	assert.Nil(t, err)

	res, err := fhps.GetFromEpoch(testKey, testEpoch)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)

	// init persister for epoch 9 and 10
	_, _ = fhps.GetFromEpoch(testKey, 9)

	// get from evicted epoch 7
	res, err = fhps.GetFromEpoch(testKey, testEpoch)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res)
}

func TestNewFullHistoryPruningStorer_GetFromEpochShouldSearchAlsoInNext(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := &pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 5,
	}
	fhps, _ := pruning.NewFullHistoryPruningStorer(fhArgs)
	testVal := []byte("value")
	testKey := []byte("key")
	testEpoch := uint32(7)
	// init persister for epoch 7
	_, _ = fhps.GetFromEpoch(nil, testEpoch)

	err := fhps.PutInEpoch(testKey, testVal, testEpoch)
	assert.Nil(t, err)

	res1, err := fhps.GetFromEpoch(testKey, testEpoch)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res1)

	res2, err := fhps.GetFromEpoch(testKey, testEpoch-1)
	assert.Nil(t, err)
	assert.Equal(t, testVal, res2)
}

func TestFullHistoryPruningStorer_IsEpochActive(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := &pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 2,
	}
	fhps, _ := pruning.NewFullHistoryPruningStorer(fhArgs)
	testEpoch := uint32(7)
	_ = fhps.ChangeEpochSimple(testEpoch - 2)
	_ = fhps.ChangeEpochSimple(testEpoch - 1)
	_ = fhps.ChangeEpochSimple(testEpoch)

	assert.True(t, fhps.IsEpochActive(testEpoch))
	assert.True(t, fhps.IsEpochActive(testEpoch-1))
	assert.False(t, fhps.IsEpochActive(testEpoch-2))
	assert.False(t, fhps.IsEpochActive(testEpoch-3))
}

func TestNewFullHistoryShardedPruningStorer_ShouldWork(t *testing.T) {
	t.Parallel()
	args := getDefaultArgs()
	fhArgs := &pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 5,
	}
	fhps, err := pruning.NewShardedFullHistoryPruningStorer(fhArgs, 2)

	require.Nil(t, err)
	require.NotNil(t, fhps)
}
