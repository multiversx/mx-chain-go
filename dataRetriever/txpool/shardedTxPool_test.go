package txpool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_NewShardedTxPool(t *testing.T) {
	pool, err := newTxPoolToTest()

	require.Nil(t, err)
	require.NotNil(t, pool)
	require.Implements(t, (*dataRetriever.ShardedDataCacherNotifier)(nil), pool)
}

func Test_NewShardedTxPool_WhenBadConfig(t *testing.T) {
	goodArgs := ArgShardedTxPool{
		Config: storageunit.CacheConfig{
			Capacity:             100,
			SizePerSender:        10,
			SizeInBytes:          409600,
			SizeInBytesPerSender: 40960,
			Shards:               16,
		},
		TxGasHandler: &txcachemocks.TxGasHandlerMock{
			MinimumGasMove:       50000,
			MinimumGasPrice:      1000000000,
			GasProcessingDivisor: 100,
		},
		NumberOfShards: 1,
	}

	args := goodArgs
	args.Config.SizeInBytes = 0
	pool, err := NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrCacheConfigInvalidSizeInBytes.Error())

	args = goodArgs
	args.Config.SizeInBytesPerSender = 0
	pool, err = NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrCacheConfigInvalidSizeInBytes.Error())

	args = goodArgs
	args.Config.Capacity = 0
	pool, err = NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrCacheConfigInvalidSize.Error())

	args = goodArgs
	args.Config.SizePerSender = 0
	pool, err = NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrCacheConfigInvalidSize.Error())

	args = goodArgs
	args.Config.Shards = 0
	pool, err = NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrCacheConfigInvalidShards.Error())

	args = goodArgs
	args.TxGasHandler = nil
	pool, err = NewShardedTxPool(args)
	require.Nil(t, pool)
	require.NotNil(t, err)
	require.Errorf(t, err, dataRetriever.ErrNilTxGasHandler.Error())

	args = goodArgs
	args.TxGasHandler = &txcachemocks.TxGasHandlerMock{
		MinimumGasMove:       50000,
		MinimumGasPrice:      0,
		GasProcessingDivisor: 1,
	}
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
	config := storageunit.CacheConfig{SizeInBytes: 419430400, SizeInBytesPerSender: 614400, Capacity: 600000, SizePerSender: 1000, Shards: 1}
	args := ArgShardedTxPool{
		Config: config,
		TxGasHandler: &txcachemocks.TxGasHandlerMock{
			MinimumGasMove:       50000,
			MinimumGasPrice:      1000000000,
			GasProcessingDivisor: 1,
		},
		NumberOfShards: 2,
	}

	pool, err := NewShardedTxPool(args)
	require.Nil(t, err)

	require.Equal(t, true, pool.configPrototypeSourceMe.EvictionEnabled)
	require.Equal(t, 209715200, int(pool.configPrototypeSourceMe.NumBytesThreshold))
	require.Equal(t, 614400, int(pool.configPrototypeSourceMe.NumBytesPerSenderThreshold))
	require.Equal(t, 1000, int(pool.configPrototypeSourceMe.CountPerSenderThreshold))
	require.Equal(t, 100, int(pool.configPrototypeSourceMe.NumSendersToPreemptivelyEvict))
	require.Equal(t, 300000, int(pool.configPrototypeSourceMe.CountThreshold))

	require.Equal(t, 300000, int(pool.configPrototypeDestinationMe.MaxNumItems))
	require.Equal(t, 209715200, int(pool.configPrototypeDestinationMe.MaxNumBytes))
}

func Test_ShardDataStore_Or_GetTxCache(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	secondPoolAsInterface, _ := newTxPoolToTest()
	secondPool := secondPoolAsInterface.(*shardedTxPool)
	pool.SetEvictionHandler(secondPool)

	fooGenericCache := pool.ShardDataStore("foo")
	fooTxCache := pool.getTxCache("foo")
	require.Equal(t, fooGenericCache, fooTxCache)
}

func Test_ShardDataStore_CreatesIfMissingWithoutConcurrencyIssues(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	var wg sync.WaitGroup

	// 100 * 10 caches will be created

	for i := 1; i <= 100; i++ {
		wg.Add(1)

		go func(i int) {
			for j := 111; j <= 120; j++ {
				pool.ShardDataStore(fmt.Sprintf("%d_%d", i, j))
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	require.Equal(t, 1000, len(pool.backingMap))

	for i := 1; i <= 100; i++ {
		for j := 111; j <= 120; j++ {
			_, inMap := pool.backingMap[fmt.Sprintf("%d_%d", i, j)]
			require.True(t, inMap)
		}
	}
}

func Test_AddData(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)
	cache := pool.getTxCache("0")

	pool.AddData([]byte("hash-invalid-cache"), createTx("alice", 0), 0, "invalid-cache-id")
	pool.AddData([]byte("hash-x"), createTx("alice", 42), 0, "0")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), 0, "0")
	require.Equal(t, 2, cache.Len())

	// Try to add again, duplication does not occur
	pool.AddData([]byte("hash-x"), createTx("alice", 42), 0, "0")
	require.Equal(t, 2, cache.Len())

	_, ok := cache.GetByTxHash([]byte("hash-x"))
	require.True(t, ok)
	_, ok = cache.GetByTxHash([]byte("hash-y"))
	require.True(t, ok)
}

func Test_AddData_NoPanic_IfNotATransaction(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()

	require.NotPanics(t, func() {
		poolAsInterface.AddData([]byte("hash"), &thisIsNotATransaction{}, 0, "1")
	})
}

func Test_AddData_CallsOnAddedHandlers(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	numAdded := uint32(0)
	pool.RegisterOnAdded(func(key []byte, value interface{}) {
		atomic.AddUint32(&numAdded, 1)
	})

	// Second addition is ignored (txhash-based deduplication)
	pool.AddData([]byte("hash-1"), createTx("alice", 42), 0, "0")
	pool.AddData([]byte("hash-1"), createTx("alice", 42), 0, "0")

	waitABit()
	require.Equal(t, uint32(1), atomic.LoadUint32(&numAdded))
}

func Test_SearchFirstData(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	tx := createTx("alice", 42)
	pool.AddData([]byte("hash-x"), tx, 0, "0")
	pool.AddData([]byte("hash-y"), tx, 0, "0_1")
	pool.AddData([]byte("hash-z"), tx, 0, "2_3")

	foundTx, ok := pool.SearchFirstData([]byte("hash-x"))
	require.True(t, ok)
	require.Equal(t, tx, foundTx)

	foundTx, ok = pool.SearchFirstData([]byte("hash-y"))
	require.True(t, ok)
	require.Equal(t, tx, foundTx)

	foundTx, ok = pool.SearchFirstData([]byte("hash-z"))
	require.True(t, ok)
	require.Equal(t, tx, foundTx)
}

func Test_RemoveData(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), 0, "0")
	pool.AddData([]byte("hash-y"), createTx("bob", 43), 0, "1")

	pool.RemoveData([]byte("hash-x"), "0")
	pool.RemoveData([]byte("hash-y"), "1")
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
	cache := pool.getTxCache("0")

	pool.AddData([]byte("hash-x"), createTx("alice", 42), 0, "0")
	pool.AddData([]byte("hash-y"), createTx("bob", 43), 0, "0")
	require.Equal(t, 2, cache.Len())

	pool.RemoveSetOfDataFromPool([][]byte{[]byte("hash-x"), []byte("hash-y")}, "0")
	require.Zero(t, cache.Len())
}

func Test_RemoveDataFromAllShards(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), 0, "0")
	pool.AddData([]byte("hash-x"), createTx("alice", 42), 0, "1")
	pool.RemoveDataFromAllShards([]byte("hash-x"))

	require.Zero(t, pool.getTxCache("0").Len())
	require.Zero(t, pool.getTxCache("1").Len())
}

func Test_MergeShardStores(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), 0, "1_0")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), 0, "2_0")
	pool.MergeShardStores("1_0", "2_0")

	require.Equal(t, 0, pool.getTxCache("1_0").Len())
	require.Equal(t, 2, pool.getTxCache("2_0").Len())
}

func Test_Clear(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), 0, "0")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), 0, "1")

	pool.Clear()
	require.Zero(t, pool.getTxCache("0").Len())
	require.Zero(t, pool.getTxCache("1").Len())
}

func Test_ClearShardStore(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.AddData([]byte("hash-x"), createTx("alice", 42), 0, "1")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), 0, "1")
	pool.AddData([]byte("hash-z"), createTx("alice", 15), 0, "5")

	pool.ClearShardStore("1")
	require.Equal(t, 0, pool.getTxCache("1").Len())
	require.Equal(t, 1, pool.getTxCache("5").Len())
}

func Test_RegisterOnAdded(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	pool.RegisterOnAdded(func(key []byte, value interface{}) {})
	require.Equal(t, 1, len(pool.onAddCallbacks))

	pool.RegisterOnAdded(nil)
	require.Equal(t, 1, len(pool.onAddCallbacks))
}

func Test_GetCounts(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	require.Equal(t, int64(0), pool.GetCounts().GetTotal())
	pool.AddData([]byte("hash-x"), createTx("alice", 42), 0, "1")
	pool.AddData([]byte("hash-y"), createTx("alice", 43), 0, "1")
	pool.AddData([]byte("hash-z"), createTx("bob", 15), 0, "3")
	require.Equal(t, int64(3), pool.GetCounts().GetTotal())
	pool.RemoveDataFromAllShards([]byte("hash-x"))
	require.Equal(t, int64(2), pool.GetCounts().GetTotal())
	pool.Clear()
	require.Equal(t, int64(0), pool.GetCounts().GetTotal())
}

func Test_Keys(t *testing.T) {
	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)

	txsHashes := [][]byte{[]byte("hash-w"), []byte("hash-x"), []byte("hash-y"), []byte("hash-z")}
	pool.AddData(txsHashes[0], createTx("alice", 42), 0, "1")
	pool.AddData(txsHashes[1], createTx("alice", 43), 0, "1")
	pool.AddData(txsHashes[2], createTx("bob", 15), 0, "2")
	pool.AddData(txsHashes[3], createTx("bob", 15), 0, "3")

	require.ElementsMatch(t, txsHashes, pool.Keys())
}

func TestShardedTxPool_Diagnose(t *testing.T) {
	t.Parallel()

	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)
	pool.AddData([]byte("hash"), createTx("alice", 10), 0, "0")
	pool.Diagnose(true)
}

func TestShardedTxPool_ImmunizeSetOfDataAgainstEviction(t *testing.T) {
	t.Parallel()

	poolAsInterface, _ := newTxPoolToTest()
	pool := poolAsInterface.(*shardedTxPool)
	pool.ImmunizeSetOfDataAgainstEviction([][]byte{[]byte("hash")}, "0")
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

func Test_routeToCacheUnions(t *testing.T) {
	config := storageunit.CacheConfig{
		Capacity:             100,
		SizePerSender:        10,
		SizeInBytes:          409600,
		SizeInBytesPerSender: 40960,
		Shards:               1,
	}
	args := ArgShardedTxPool{
		Config: config,
		TxGasHandler: &txcachemocks.TxGasHandlerMock{
			MinimumGasMove:       50000,
			MinimumGasPrice:      200000000000,
			GasProcessingDivisor: 100,
		},
		NumberOfShards: 4,
		SelfShardID:    42,
	}
	pool, _ := NewShardedTxPool(args)

	require.Equal(t, "42", pool.routeToCacheUnions("42"))
	require.Equal(t, "42", pool.routeToCacheUnions("42_0"))
	require.Equal(t, "42", pool.routeToCacheUnions("42_1"))
	require.Equal(t, "42", pool.routeToCacheUnions("42_2"))
	require.Equal(t, "42", pool.routeToCacheUnions("42_42"))
	require.Equal(t, "2_5", pool.routeToCacheUnions("2_5"))
	require.Equal(t, "foobar", pool.routeToCacheUnions("foobar"))
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
	config := storageunit.CacheConfig{
		Capacity:             100,
		SizePerSender:        10,
		SizeInBytes:          409600,
		SizeInBytesPerSender: 40960,
		Shards:               1,
	}
	args := ArgShardedTxPool{
		Config: config,
		TxGasHandler: &txcachemocks.TxGasHandlerMock{
			MinimumGasMove:       50000,
			MinimumGasPrice:      200000000000,
			GasProcessingDivisor: 100,
		},
		NumberOfShards: 4,
		SelfShardID:    0,
	}
	return NewShardedTxPool(args)
}
