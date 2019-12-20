package txpool

import (
	"fmt"
	"sync"
	"testing"
	"time"

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

func Test_AddData_Or_AddTx(t *testing.T) {
	poolAsInterface := NewShardedTxPool(storageUnit.CacheConfig{Size: 75000, Shards: 16})
	pool := poolAsInterface.(*shardedTxPool)
	cache := pool.getTxCache("1")

	pool.AddData([]byte("hash-x"), createTx("alice", 42), "1")
	pool.addTx([]byte("hash-y"), createTx("alice", 43), "1")
	require.Equal(t, int64(2), cache.CountTx())

	// Try to add again, duplication does not occur
	pool.AddData([]byte("hash-x"), createTx("alice", 42), "1")
	pool.addTx([]byte("hash-y"), createTx("alice", 43), "1")
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
