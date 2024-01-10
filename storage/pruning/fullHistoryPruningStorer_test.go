package pruning_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/pathmanager"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFullHistoryPruningStorer_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 10,
	}
	fhps, err := pruning.NewFullHistoryPruningStorer(fhArgs)

	assert.NotNil(t, fhps)
	assert.Nil(t, err)
}

func TestNewFullHistoryPruningStorer_InvalidNumberOfActivePersistersShouldErr(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()

	fhArgs := pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 0,
	}
	fhps, err := pruning.NewFullHistoryPruningStorer(fhArgs)

	assert.Nil(t, fhps)
	assert.Equal(t, storage.ErrInvalidNumberOfOldPersisters, err)

	fhArgs = pruning.FullHistoryStorerArgs{
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
	fhArgs := pruning.FullHistoryStorerArgs{
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
	fhArgs := pruning.FullHistoryStorerArgs{
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
	persistersByPath["Epoch_0"] = database.NewMemDB()
	args := getDefaultArgs()
	args.DbPath = "Epoch_0"
	args.EpochsData.NumOfActivePersisters = 1
	args.EpochsData.NumOfEpochsToKeep = 2

	fhArgs := pruning.FullHistoryStorerArgs{
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
	fhArgs := pruning.FullHistoryStorerArgs{
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

func TestNewFullHistoryPruningStorer_GetBulkFromEpoch(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 5,
	}
	fhps, _ := pruning.NewFullHistoryPruningStorer(fhArgs)
	testVal0, testVal1 := []byte("value0"), []byte("value1")
	testKey0, testKey1 := []byte("key0"), []byte("key1")
	testEpoch := uint32(7)

	_ = fhps.PutInEpoch(testKey0, testVal0, testEpoch)
	_ = fhps.PutInEpoch(testKey1, testVal1, testEpoch)

	res, err := fhps.GetBulkFromEpoch([][]byte{testKey0, testKey1}, testEpoch)
	assert.Nil(t, err)

	expected := []data.KeyValuePair{
		{Key: testKey0, Value: testVal0},
		{Key: testKey1, Value: testVal1},
	}
	assert.Equal(t, expected, res)
}

func TestNewFullHistoryPruningStorer_GetBulkFromEpochShouldNotLoadFromCache(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 5,
	}
	fhps, _ := pruning.NewFullHistoryPruningStorer(fhArgs)
	testVal0, testVal1 := []byte("value0"), []byte("value1")
	testKey0, testKey1 := []byte("key0"), []byte("key1")
	testEpoch := uint32(7)

	_ = fhps.PutInEpoch(testKey0, testVal0, testEpoch)
	_ = fhps.PutInEpoch(testKey1, testVal1, testEpoch)

	fhps.ClearCache()

	res, err := fhps.GetBulkFromEpoch([][]byte{testKey0, testKey1}, testEpoch)
	assert.Nil(t, err)

	expected := []data.KeyValuePair{
		{Key: testKey0, Value: testVal0},
		{Key: testKey1, Value: testVal1},
	}
	assert.Equal(t, expected, res)
}

func TestFullHistoryPruningStorer_IsEpochActive(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := pruning.FullHistoryStorerArgs{
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
	fhArgs := pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 5,
	}
	fhps, err := pruning.NewShardedFullHistoryPruningStorer(fhArgs, 2)

	require.Nil(t, err)
	require.NotNil(t, fhps)
}

func TestFullHistoryPruningStorer_Close(t *testing.T) {
	t.Parallel()

	args := getDefaultArgs()
	fhArgs := pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 5,
	}
	fhps, _ := pruning.NewShardedFullHistoryPruningStorer(fhArgs, 2)
	for i := 0; i < 10; i++ {
		_, _ = fhps.GetFromEpoch([]byte("key"), uint32(i))
	}
	assert.Equal(t, int(fhArgs.NumOfOldActivePersisters), fhps.GetOldEpochsActivePersisters().Len())

	err := fhps.Close()
	assert.Nil(t, err)
	assert.Equal(t, 0, fhps.GetOldEpochsActivePersisters().Len())
}

func TestFullHistoryPruningStorer_ConcurrentOperations(t *testing.T) {
	t.Skip("this test should be run only when troubleshooting pruning storer concurrent operations")

	startTime := time.Now()

	_ = logger.SetLogLevel("*:DEBUG")
	dbName := "db-concurrent-test"
	testDir := t.TempDir()

	fmt.Println(testDir)
	args := getDefaultArgs()
	pfh := factory.NewPersisterFactoryHandler(2, 1)
	persisterFactory, err := pfh.CreatePersisterHandler(config.DBConfig{
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
	fhArgs := pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 2,
	}

	fhps, _ := pruning.NewFullHistoryPruningStorer(fhArgs)
	require.NotNil(t, fhps)

	defer func() {
		_ = fhps.Close()
	}()

	rnd := random.ConcurrentSafeIntRandomizer{}
	numOperations := 5000
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

				_ = fhps.ChangeEpochSimple(uint32(newEpoch))
				log.Debug("called ChangeEpochSimple", "epoch", newEpoch)
				wg.Done()
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	for idx := 0; idx < numOperations; idx++ {
		if idx%6 != 0 {
			chanChangeEpoch <- struct{}{}
			continue
		}

		go func(index int) {
			time.Sleep(time.Duration(index) * 1 * time.Millisecond)
			switch index % 6 {
			case 1:
				_, _ = fhps.GetFromEpoch([]byte("key"), uint32(index-1))
			case 2:
				_ = fhps.Put([]byte("key"), []byte("value"))
			case 3:
				_, _ = fhps.Get([]byte("key"))
			case 4:
				_, _ = fhps.GetFromEpoch([]byte("key"), uint32(rnd.Intn(100)))
			case 5:
				_, _ = fhps.GetBulkFromEpoch([][]byte{[]byte("key")}, uint32(rnd.Intn(100)))
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

func TestFullHistoryPruningStorer_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var fhps *pruning.FullHistoryPruningStorer
	require.True(t, fhps.IsInterfaceNil())

	args := getDefaultArgs()
	fhArgs := pruning.FullHistoryStorerArgs{
		StorerArgs:               args,
		NumOfOldActivePersisters: 10,
	}
	fhps, _ = pruning.NewFullHistoryPruningStorer(fhArgs)
	require.False(t, fhps.IsInterfaceNil())
}
