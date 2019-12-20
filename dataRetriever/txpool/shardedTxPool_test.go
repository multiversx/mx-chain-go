package txpool

import (
	"fmt"
	"sync"
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
	pool := NewShardedTxPool(storageUnit.CacheConfig{})

	require.NotNil(t, pool)
	require.Implements(t, (*dataRetriever.ShardedDataCacherNotifier)(nil), pool)
}

func Test_NewShardedTxPool_ComputesEvictionConfig(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)

	require.Equal(t, true, pool.evictionConfig.Enabled)
	require.Equal(t, uint32(75000), pool.evictionConfig.CountThreshold)
	require.Equal(t, uint32(1000), pool.evictionConfig.ThresholdEvictSenders)
	require.Equal(t, uint32(500), pool.evictionConfig.NumOldestSendersToEvict)
	require.Equal(t, uint32(750), pool.evictionConfig.ALotOfTransactionsForASender)
	require.Equal(t, uint32(750*3/4), pool.evictionConfig.NumTxsToEvictForASenderWithALot)
}

func Test_ShardDataStore_Or_GetTxCache(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000, Shards: 16})
	pool := poolAsInterface.(*shardedTxPool)

	fooGenericCache := pool.ShardDataStore("foo")
	fooTxCache := pool.getTxCache("foo")
	require.Equal(t, fooGenericCache, fooTxCache)
}

func Test_ShardDataStore_CreatesIfMissingWithoutConcurrencyIssues(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000, Shards: 16})
	pool := poolAsInterface.(*shardedTxPool)

	var wg sync.WaitGroup

	// 100 * 100 caches will be created

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			for j := 0; j < 100; j++ {
				pool.ShardDataStore(fmt.Sprintf("%d_%d", i, j))
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	require.Equal(t, 10000, len(pool.backingMap))

	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			_, inMap := pool.backingMap[fmt.Sprintf("%d_%d", i, j)]
			require.True(t, inMap)
		}
	}
}

func Test_AddData(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000, Shards: 16})
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

func Test_AddData_Panics_IfNotATransaction(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{})

	require.Panics(t, func() {
		poolAsInterface.AddData([]byte("hash"), &thisIsNotATransaction{}, "1")
	})
}

func Test_AddData_CallsOnAddedHandlers(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)

	numAdded := 0
	pool.RegisterHandler(func(key []byte) {
		numAdded++
	})

	// Second addition is ignored (txhash-based deduplication)
	pool.AddData([]byte("hash-1"), createTx("alice", 42), "1")
	pool.AddData([]byte("hash-1"), createTx("whatever", 43), "1")

	waitABit()
	require.Equal(t, 1, numAdded)
}

func Test_SearchFirstData(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)

	tx := createTx("alice", 42)
	pool.AddData([]byte("hash-x"), tx, "1")

	foundTx, ok := pool.SearchFirstData([]byte("hash-x"))
	require.True(t, ok)
	require.Equal(t, tx, foundTx)
}

func Test_RemoveData(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
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
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)
	cache := pool.getTxCache("foo")

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-y"), createTx("bob", 43), "foo")
	require.Equal(t, int64(2), cache.CountTx())

	pool.RemoveSetOfDataFromPool([][]byte{[]byte("hash-x"), []byte("hash-y")}, "foo")
	require.Zero(t, cache.CountTx())
}

func Test_RemoveDataFromAllShards(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-x"), createTx("alice", 42), "bar")
	pool.RemoveDataFromAllShards([]byte("hash-x"))

	require.Zero(t, pool.getTxCache("foo").CountTx())
	require.Zero(t, pool.getTxCache("bar").CountTx())
}

func Test_MergeShardStores(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), "bar")
	pool.MergeShardStores("foo", "bar")

	require.Equal(t, int64(0), pool.getTxCache("foo").CountTx())
	require.Equal(t, int64(2), pool.getTxCache("bar").CountTx())
}

func Test_MoveData(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), "bar")

	pool.MoveData("foo", "bar", [][]byte{[]byte("hash-x")})
	require.Equal(t, int64(0), pool.getTxCache("foo").CountTx())
	require.Equal(t, int64(2), pool.getTxCache("bar").CountTx())

	pool.MoveData("bar", "foo", [][]byte{[]byte("hash-x"), []byte("hash-y")})
	require.Equal(t, int64(2), pool.getTxCache("foo").CountTx())
	require.Equal(t, int64(0), pool.getTxCache("bar").CountTx())
}

func Test_Clear(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), "bar")

	pool.Clear()
	require.Zero(t, pool.getTxCache("foo").CountTx())
	require.Zero(t, pool.getTxCache("bar").CountTx())
}

func Test_CreateShardStore(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "foo")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), "foo")
	pool.AddData([]byte("hash-z"), createTx("alice", 15), "bar")

	pool.ClearShardStore("foo")
	require.Equal(t, int64(0), pool.getTxCache("foo").CountTx())
	require.Equal(t, int64(1), pool.getTxCache("bar").CountTx())
}

func Test_RegisterHandler(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)

	pool.RegisterHandler(func(key []byte) {})
	require.Equal(t, 1, len(pool.onAddCallbacks))

	pool.RegisterHandler(nil)
	require.Equal(t, 1, len(pool.onAddCallbacks))
}

func Test_IsInterfaceNil(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	require.False(t, check.IfNil(poolAsInterface))

	makeNil := func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}

	thisIsNil := makeNil()
	require.True(t, check.IfNil(thisIsNil))
}

func Test_NotImplementedFunctions(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000})
	pool := poolAsInterface.(*shardedTxPool)

	require.Panics(t, func() { pool.CreateShardStore("foo") })
}

func createTx(sender string, nonce uint64) data.TransactionHandler {
	return &transaction.Transaction{
		SndAddr: []byte(sender),
		Nonce:   nonce,
	}
}

func waitABit() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		wg.Done()
	}()
	wg.Wait()
}

type thisIsNotATransaction struct {
}
