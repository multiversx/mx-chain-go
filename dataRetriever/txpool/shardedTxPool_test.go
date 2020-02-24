package txpool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/require"
)

func Test_NewShardedTxPool(t *testing.T) {
	pool, err := newTxPoolToTest()

	require.Nil(t, err)
	require.NotNil(t, pool)
	require.Implements(t, (*dataRetriever.ShardedDataCacherNotifier)(nil), pool)
}

func Test_NewShardedTxPool_WhenBadConfig(t *testing.T) {
	goodArgs := ArgShardedTxPool{Config: storageUnit.CacheConfig{Size: 100, SizeInBytes: 40960, Shards: 16}, MinGasPrice: 100000000000000, NumberOfShards: 1}

	args := goodArgs
	args.Config = storageUnit.CacheConfig{SizeInBytes: 1}
	pool, err := NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrCacheConfigInvalidSizeInBytes.Error())

	args = goodArgs
	args.Config = storageUnit.CacheConfig{SizeInBytes: 40960, Size: 1}
	pool, err = NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrCacheConfigInvalidShards.Error())

	args = goodArgs
	args.Config = storageUnit.CacheConfig{SizeInBytes: 40960, Shards: 1}
	pool, err = NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrCacheConfigInvalidSize.Error())

	args = goodArgs
	args.MinGasPrice = 0
	pool, err = NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrCacheConfigInvalidEconomics.Error())

	args = goodArgs
	args.NumberOfShards = 0
	pool, err = NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrCacheConfigInvalidSharding.Error())
}

func Test_NewShardedTxPool_ComputesCacheConfig(t *testing.T) {
	config := storageUnit.CacheConfig{SizeInBytes: 524288000, Size: 900000, Shards: 1}
	args := ArgShardedTxPool{Config: config, MinGasPrice: 100000000000000, NumberOfShards: 5}

	poolAsInterface, err := NewShardedTxPool(args)
	require.Nil(t, err)

	pool := poolAsInterface.(*shardedTxPool)

	require.Equal(t, true, pool.cacheConfigPrototype.EvictionEnabled)
	require.Equal(t, uint32(58254222), pool.cacheConfigPrototype.NumBytesThreshold)
	require.Equal(t, uint32(100000), pool.cacheConfigPrototype.CountThreshold)
	require.Equal(t, uint32(100), pool.cacheConfigPrototype.NumSendersToEvictInOneStep)
	require.Equal(t, uint32(500), pool.cacheConfigPrototype.LargeNumOfTxsForASender)
	require.Equal(t, uint32(100), pool.cacheConfigPrototype.NumTxsToEvictFromASender)
	require.Equal(t, uint32(100), pool.cacheConfigPrototype.MinGasPriceMicroErd)
}

func Test_ShardDataStore_Or_GetTxCache(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	fooGenericCache := pool.ShardDataStore("foo")
	fooTxCache := pool.getTxCache("foo")
	require.Equal(t, fooGenericCache, fooTxCache)
}

func Test_ShardDataStore_CreatesIfMissingWithoutConcurrencyIssues(t *testing.T) {
	t.Skip("Skip this because it requires non-merged tx pool caches")

	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	var wg sync.WaitGroup

	// 100 * 10 caches will be created

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			for j := 0; j < 10; j++ {
				pool.ShardDataStore(fmt.Sprintf("%d_%d", i, j))
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	require.Equal(t, 1000, len(pool.backingMap))

	for i := 0; i < 100; i++ {
		for j := 0; j < 10; j++ {
			_, inMap := pool.backingMap[fmt.Sprintf("%d_%d", i, j)]
			require.True(t, inMap)
		}
	}
}

func Test_AddData(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)
	cache := pool.getTxCache("1")

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "1")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), "1")
	require.Equal(t, int64(2), cache.CountTx())

	// Try to add again, duplication does not occur
	pool.AddData([]byte("hash-x"), createTx("alice", 42), "1")
	require.Equal(t, int64(2), cache.CountTx())

	_, ok := cache.GetByTxHash([]byte("hash-x"))
	require.True(t, ok)
	_, ok = cache.GetByTxHash([]byte("hash-y"))
	require.True(t, ok)
}

func Test_AddData_NoPanic_IfNotATransaction(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()

	require.NotPanics(t, func() {
		poolAsInterface.AddData([]byte("hash"), &thisIsNotATransaction{}, "1")
	})
}

func Test_AddData_CallsOnAddedHandlers(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	numAdded := uint32(0)
	pool.RegisterHandler(func(key []byte) {
		atomic.AddUint32(&numAdded, 1)
	})

	// Second addition is ignored (txhash-based deduplication)
	pool.AddData([]byte("hash-1"), createTx("alice", 42), "1")
	pool.AddData([]byte("hash-1"), createTx("whatever", 43), "1")

	waitABit()
	require.Equal(t, uint32(1), atomic.LoadUint32(&numAdded))
}

func Test_SearchFirstData(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	tx := createTx("alice", 42)
	pool.AddData([]byte("hash-x"), tx, "1")

	foundTx, ok := pool.SearchFirstData([]byte("hash-x"))
	require.True(t, ok)
	require.Equal(t, tx, foundTx)
}

func Test_RemoveData(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-y"), createTx("bob", 43), "bar")

	pool.RemoveData([]byte("hash-x"), "foo")
	pool.RemoveData([]byte("hash-y"), "bar")
	xTx, xOk := pool.searchFirstTx([]byte("hash-x"))
	yTx, yOk := pool.searchFirstTx([]byte("hash-y"))
	require.False(t, xOk)
	require.False(t, yOk)
	require.Nil(t, xTx)
	require.Nil(t, yTx)
}

func Test_RemoveSetOfDataFromPool(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)
	cache := pool.getTxCache("foo")

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-y"), createTx("bob", 43), "foo")
	require.Equal(t, int64(2), cache.CountTx())

	pool.RemoveSetOfDataFromPool([][]byte{[]byte("hash-x"), []byte("hash-y")}, "foo")
	require.Zero(t, cache.CountTx())
}

func Test_RemoveDataFromAllShards(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-x"), createTx("alice", 42), "bar")
	pool.RemoveDataFromAllShards([]byte("hash-x"))

	require.Zero(t, pool.getTxCache("foo").CountTx())
	require.Zero(t, pool.getTxCache("bar").CountTx())
}

func Test_MergeShardStores(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), "bar")
	pool.MergeShardStores("foo", "bar")

	require.Equal(t, int64(0), pool.getTxCache("foo").CountTx())
	require.Equal(t, int64(2), pool.getTxCache("bar").CountTx())
}

func Test_Clear(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), "bar")

	pool.Clear()
	require.Zero(t, pool.getTxCache("foo").CountTx())
	require.Zero(t, pool.getTxCache("bar").CountTx())
}

func Test_CreateShardStore(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), "foo")
	pool.AddData([]byte("hash-z"), createTx("alice", 15), "bar")

	pool.ClearShardStore("foo")
	require.Equal(t, int64(0), pool.getTxCache("foo").CountTx())
	require.Equal(t, int64(1), pool.getTxCache("bar").CountTx())
}

func Test_RegisterHandler(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.RegisterHandler(func(key []byte) {})
	require.Equal(t, 1, len(pool.onAddCallbacks))

	pool.RegisterHandler(nil)
	require.Equal(t, 1, len(pool.onAddCallbacks))
}

func Test_IsInterfaceNil(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	require.False(t, check.IfNil(poolAsInterface))

	makeNil := func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}

	thisIsNil := makeNil()
	require.True(t, check.IfNil(thisIsNil))
}

func Test_NotImplementedFunctions(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	require.NotPanics(t, func() { pool.CreateShardStore("foo") })
}

func Test_routeToCache(t *testing.T) {
	config := storageUnit.CacheConfig{Size: 100, SizeInBytes: 40960, Shards: 16}
	args := ArgShardedTxPool{Config: config, MinGasPrice: 100000000000000, NumberOfShards: 4, SelfShardID: 42}
	poolAsInterface, _ := NewShardedTxPool(args)
	pool := poolAsInterface.(*shardedTxPool)

	require.Equal(t, "42", pool.routeToCache("42"))
	require.Equal(t, "42", pool.routeToCache("42_0"))
	require.Equal(t, "42", pool.routeToCache("42_1"))
	require.Equal(t, "42", pool.routeToCache("42_2"))
	require.Equal(t, "42", pool.routeToCache("42_*"))
	require.Equal(t, "42", pool.routeToCache("42_42"))
	require.Equal(t, "2_5", pool.routeToCache("2_5"))
	require.Equal(t, "foobar", pool.routeToCache("foobar"))
}

func createTx(sender string, nonce uint64) data.TransactionHandler {
	return &transaction.Transaction{
		SndAddr: []byte(sender),
		Nonce:   nonce,
	}
}

func waitABit() {
	time.Sleep(10 * time.Millisecond)
}

type thisIsNotATransaction struct {
}

func newTxPoolToTest() (dataRetriever.ShardedDataCacherNotifier, error) {
	config := storageUnit.CacheConfig{Size: 100, SizeInBytes: 40960, Shards: 16}
	args := ArgShardedTxPool{Config: config, MinGasPrice: 100000000000000, NumberOfShards: 4}
	return NewShardedTxPool(args)
}
