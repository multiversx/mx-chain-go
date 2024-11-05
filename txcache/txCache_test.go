package txcache

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/multiversx/mx-chain-storage-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewTxCache(t *testing.T) {
	config := ConfigSourceMe{
		Name:                        "test",
		NumChunks:                   16,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
		CountThreshold:              math.MaxUint32,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             true,
		NumItemsToPreemptivelyEvict: 1,
	}

	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	cache, err := NewTxCache(config, txGasHandler)
	require.Nil(t, err)
	require.NotNil(t, cache)

	badConfig := config
	badConfig.Name = ""
	requireErrorOnNewTxCache(t, badConfig, common.ErrInvalidConfig, "config.Name", txGasHandler)

	badConfig = config
	badConfig.NumChunks = 0
	requireErrorOnNewTxCache(t, badConfig, common.ErrInvalidConfig, "config.NumChunks", txGasHandler)

	badConfig = config
	badConfig.NumBytesPerSenderThreshold = 0
	requireErrorOnNewTxCache(t, badConfig, common.ErrInvalidConfig, "config.NumBytesPerSenderThreshold", txGasHandler)

	badConfig = config
	badConfig.CountPerSenderThreshold = 0
	requireErrorOnNewTxCache(t, badConfig, common.ErrInvalidConfig, "config.CountPerSenderThreshold", txGasHandler)

	badConfig = config
	cache, err = NewTxCache(config, nil)
	require.Nil(t, cache)
	require.Equal(t, common.ErrNilTxGasHandler, err)

	badConfig = config
	badConfig.NumBytesThreshold = 0
	requireErrorOnNewTxCache(t, badConfig, common.ErrInvalidConfig, "config.NumBytesThreshold", txGasHandler)

	badConfig = config
	badConfig.CountThreshold = 0
	requireErrorOnNewTxCache(t, badConfig, common.ErrInvalidConfig, "config.CountThreshold", txGasHandler)
}

func requireErrorOnNewTxCache(t *testing.T, config ConfigSourceMe, errExpected error, errPartialMessage string, txGasHandler TxGasHandler) {
	cache, errReceived := NewTxCache(config, txGasHandler)
	require.Nil(t, cache)
	require.True(t, errors.Is(errReceived, errExpected))
	require.Contains(t, errReceived.Error(), errPartialMessage)
}

func Test_AddTx(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	tx := createTx([]byte("hash-1"), "alice", 1)

	ok, added := cache.AddTx(tx)
	require.True(t, ok)
	require.True(t, added)
	require.True(t, cache.Has([]byte("hash-1")))

	// Add it again (no-operation)
	ok, added = cache.AddTx(tx)
	require.True(t, ok)
	require.False(t, added)
	require.True(t, cache.Has([]byte("hash-1")))

	foundTx, ok := cache.GetByTxHash([]byte("hash-1"))
	require.True(t, ok)
	require.Equal(t, tx, foundTx)
}

func Test_AddNilTx_DoesNothing(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	txHash := []byte("hash-1")

	ok, added := cache.AddTx(&WrappedTransaction{Tx: nil, TxHash: txHash})
	require.False(t, ok)
	require.False(t, added)

	foundTx, ok := cache.GetByTxHash(txHash)
	require.False(t, ok)
	require.Nil(t, foundTx)
}

func Test_AddTx_AppliesSizeConstraintsPerSenderForNumTransactions(t *testing.T) {
	cache := newCacheToTest(maxNumBytesPerSenderUpperBound, 3)

	cache.AddTx(createTx([]byte("tx-alice-1"), "alice", 1))
	cache.AddTx(createTx([]byte("tx-alice-2"), "alice", 2))
	cache.AddTx(createTx([]byte("tx-alice-4"), "alice", 4))
	cache.AddTx(createTx([]byte("tx-bob-1"), "bob", 1))
	cache.AddTx(createTx([]byte("tx-bob-2"), "bob", 2))
	require.Equal(t, []string{"tx-alice-1", "tx-alice-2", "tx-alice-4"}, cache.getHashesForSender("alice"))
	require.Equal(t, []string{"tx-bob-1", "tx-bob-2"}, cache.getHashesForSender("bob"))
	require.True(t, cache.areInternalMapsConsistent())

	cache.AddTx(createTx([]byte("tx-alice-3"), "alice", 3))
	require.Equal(t, []string{"tx-alice-1", "tx-alice-2", "tx-alice-3"}, cache.getHashesForSender("alice"))
	require.Equal(t, []string{"tx-bob-1", "tx-bob-2"}, cache.getHashesForSender("bob"))
	require.True(t, cache.areInternalMapsConsistent())
}

func Test_AddTx_AppliesSizeConstraintsPerSenderForNumBytes(t *testing.T) {
	cache := newCacheToTest(1024, math.MaxUint32)

	cache.AddTx(createTx([]byte("tx-alice-1"), "alice", 1).withSize(128).withGasLimit(50000))
	cache.AddTx(createTx([]byte("tx-alice-2"), "alice", 2).withSize(512).withGasLimit(1500000))
	cache.AddTx(createTx([]byte("tx-alice-4"), "alice", 3).withSize(256).withGasLimit(1500000))
	cache.AddTx(createTx([]byte("tx-bob-1"), "bob", 1).withSize(512).withGasLimit(1500000))
	cache.AddTx(createTx([]byte("tx-bob-2"), "bob", 2).withSize(513).withGasLimit(1500000))

	require.Equal(t, []string{"tx-alice-1", "tx-alice-2", "tx-alice-4"}, cache.getHashesForSender("alice"))
	require.Equal(t, []string{"tx-bob-1"}, cache.getHashesForSender("bob"))
	require.True(t, cache.areInternalMapsConsistent())

	cache.AddTx(createTx([]byte("tx-alice-3"), "alice", 3).withSize(256).withGasLimit(1500000))
	cache.AddTx(createTx([]byte("tx-bob-2"), "bob", 3).withSize(512).withGasLimit(1500000))
	require.Equal(t, []string{"tx-alice-1", "tx-alice-2", "tx-alice-3"}, cache.getHashesForSender("alice"))
	require.Equal(t, []string{"tx-bob-1", "tx-bob-2"}, cache.getHashesForSender("bob"))
	require.True(t, cache.areInternalMapsConsistent())
}

func Test_RemoveByTxHash(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	cache.AddTx(createTx([]byte("hash-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-2"), "alice", 2))

	removed := cache.RemoveTxByHash([]byte("hash-1"))
	require.True(t, removed)
	cache.Remove([]byte("hash-2"))

	foundTx, ok := cache.GetByTxHash([]byte("hash-1"))
	require.False(t, ok)
	require.Nil(t, foundTx)

	foundTx, ok = cache.GetByTxHash([]byte("hash-2"))
	require.False(t, ok)
	require.Nil(t, foundTx)
}

func Test_CountTx_And_Len(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	cache.AddTx(createTx([]byte("hash-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-2"), "alice", 2))
	cache.AddTx(createTx([]byte("hash-3"), "alice", 3))

	require.Equal(t, uint64(3), cache.CountTx())
	require.Equal(t, 3, cache.Len())
}

func Test_GetByTxHash_And_Peek_And_Get(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	txHash := []byte("hash-1")
	tx := createTx(txHash, "alice", 1)
	cache.AddTx(tx)

	foundTx, ok := cache.GetByTxHash(txHash)
	require.True(t, ok)
	require.Equal(t, tx, foundTx)

	foundTxPeek, okPeek := cache.Peek(txHash)
	require.True(t, okPeek)
	require.Equal(t, tx.Tx, foundTxPeek)

	foundTxPeek, okPeek = cache.Peek([]byte("missing"))
	require.False(t, okPeek)
	require.Nil(t, foundTxPeek)

	foundTxGet, okGet := cache.Get(txHash)
	require.True(t, okGet)
	require.Equal(t, tx.Tx, foundTxGet)

	foundTxGet, okGet = cache.Get([]byte("missing"))
	require.False(t, okGet)
	require.Nil(t, foundTxGet)
}

func Test_RemoveByTxHash_WhenMissing(t *testing.T) {
	cache := newUnconstrainedCacheToTest()
	removed := cache.RemoveTxByHash([]byte("missing"))
	require.False(t, removed)
}

func Test_RemoveByTxHash_RemovesFromByHash_WhenMapsInconsistency(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	txHash := []byte("hash-1")
	tx := createTx(txHash, "alice", 1)
	cache.AddTx(tx)

	// Cause an inconsistency between the two internal maps (theoretically possible in case of misbehaving eviction)
	cache.txListBySender.removeTx(tx)

	_ = cache.RemoveTxByHash(txHash)
	require.Equal(t, 0, cache.txByHash.backingMap.Count())
}

func Test_Clear(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7))
	cache.AddTx(createTx([]byte("hash-alice-42"), "alice", 42))
	require.Equal(t, uint64(3), cache.CountTx())

	cache.Clear()
	require.Equal(t, uint64(0), cache.CountTx())
}

func Test_ForEachTransaction(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7))

	counter := 0
	cache.ForEachTransaction(func(txHash []byte, value *WrappedTransaction) {
		counter++
	})
	require.Equal(t, 2, counter)
}

func Test_GetTransactionsPoolForSender(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	txHashes1 := [][]byte{[]byte("hash-1"), []byte("hash-2")}
	txSender1 := "alice"
	wrappedTxs1 := []*WrappedTransaction{
		createTx(txHashes1[1], txSender1, 2),
		createTx(txHashes1[0], txSender1, 1),
	}
	txHashes2 := [][]byte{[]byte("hash-3"), []byte("hash-4"), []byte("hash-5")}
	txSender2 := "bob"
	wrappedTxs2 := []*WrappedTransaction{
		createTx(txHashes2[1], txSender2, 4),
		createTx(txHashes2[0], txSender2, 3),
		createTx(txHashes2[2], txSender2, 5),
	}
	cache.AddTx(wrappedTxs1[0])
	cache.AddTx(wrappedTxs1[1])
	cache.AddTx(wrappedTxs2[0])
	cache.AddTx(wrappedTxs2[1])
	cache.AddTx(wrappedTxs2[2])

	sort.Slice(wrappedTxs1, func(i, j int) bool {
		return wrappedTxs1[i].Tx.GetNonce() < wrappedTxs1[j].Tx.GetNonce()
	})
	txs := cache.GetTransactionsPoolForSender(txSender1)
	require.Equal(t, wrappedTxs1, txs)

	sort.Slice(wrappedTxs2, func(i, j int) bool {
		return wrappedTxs2[i].Tx.GetNonce() < wrappedTxs2[j].Tx.GetNonce()
	})
	txs = cache.GetTransactionsPoolForSender(txSender2)
	require.Equal(t, wrappedTxs2, txs)

	cache.RemoveTxByHash(txHashes2[0])
	expectedTxs := wrappedTxs2[1:]
	txs = cache.GetTransactionsPoolForSender(txSender2)
	require.Equal(t, expectedTxs, txs)
}

func Test_Keys(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	cache.AddTx(createTx([]byte("alice-x"), "alice", 42))
	cache.AddTx(createTx([]byte("alice-y"), "alice", 43))
	cache.AddTx(createTx([]byte("bob-x"), "bob", 42))
	cache.AddTx(createTx([]byte("bob-y"), "bob", 43))

	keys := cache.Keys()
	require.Equal(t, 4, len(keys))
	require.Contains(t, keys, []byte("alice-x"))
	require.Contains(t, keys, []byte("alice-y"))
	require.Contains(t, keys, []byte("bob-x"))
	require.Contains(t, keys, []byte("bob-y"))
}

func Test_AddWithEviction_UniformDistributionOfTxsPerSender(t *testing.T) {
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	t.Run("numSenders = 11, numTransactions = 10, countThreshold = 100, numItemsToPreemptivelyEvict = 1", func(t *testing.T) {
		config := ConfigSourceMe{
			Name:                        "untitled",
			NumChunks:                   16,
			EvictionEnabled:             true,
			NumBytesThreshold:           maxNumBytesUpperBound,
			NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
			CountThreshold:              100,
			CountPerSenderThreshold:     math.MaxUint32,
			NumItemsToPreemptivelyEvict: 1,
		}

		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)
		require.NotNil(t, cache)

		addManyTransactionsWithUniformDistribution(cache, 11, 10)

		// Eviction happens if the cache capacity is already exceeded,
		// but not if the capacity will be exceeded after the addition.
		// Thus, for the given value of "NumItemsToPreemptivelyEvict", there will be "countThreshold" + 1 transactions in the cache.
		require.Equal(t, 101, int(cache.CountTx()))
	})

	t.Run("numSenders = 3, numTransactions = 5, countThreshold = 4, numItemsToPreemptivelyEvict = 3", func(t *testing.T) {
		config := ConfigSourceMe{
			Name:                        "untitled",
			NumChunks:                   16,
			EvictionEnabled:             true,
			NumBytesThreshold:           maxNumBytesUpperBound,
			NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
			CountThreshold:              4,
			CountPerSenderThreshold:     math.MaxUint32,
			NumItemsToPreemptivelyEvict: 3,
		}

		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)
		require.NotNil(t, cache)

		addManyTransactionsWithUniformDistribution(cache, 3, 5)
		require.Equal(t, 3, int(cache.CountTx()))
	})

	t.Run("numSenders = 11, numTransactions = 10, countThreshold = 100, numItemsToPreemptivelyEvict = 2", func(t *testing.T) {
		config := ConfigSourceMe{
			Name:                        "untitled",
			NumChunks:                   16,
			EvictionEnabled:             true,
			NumBytesThreshold:           maxNumBytesUpperBound,
			NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
			CountThreshold:              100,
			CountPerSenderThreshold:     math.MaxUint32,
			NumItemsToPreemptivelyEvict: 2,
		}

		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)
		require.NotNil(t, cache)

		addManyTransactionsWithUniformDistribution(cache, 11, 10)
		require.Equal(t, 100, int(cache.CountTx()))
	})

	t.Run("numSenders = 100, numTransactions = 1000, countThreshold = 250000 (no eviction)", func(t *testing.T) {
		config := ConfigSourceMe{
			Name:                        "untitled",
			NumChunks:                   16,
			EvictionEnabled:             true,
			NumBytesThreshold:           maxNumBytesUpperBound,
			NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
			CountThreshold:              250000,
			CountPerSenderThreshold:     math.MaxUint32,
			NumItemsToPreemptivelyEvict: 1,
		}

		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)
		require.NotNil(t, cache)

		addManyTransactionsWithUniformDistribution(cache, 100, 1000)
		require.Equal(t, 100000, int(cache.CountTx()))
	})

	t.Run("numSenders = 1000, numTransactions = 500, countThreshold = 250000, NumItemsToPreemptivelyEvict = 50000", func(t *testing.T) {
		config := ConfigSourceMe{
			Name:                        "untitled",
			NumChunks:                   16,
			EvictionEnabled:             true,
			NumBytesThreshold:           maxNumBytesUpperBound,
			NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
			CountThreshold:              250000,
			CountPerSenderThreshold:     math.MaxUint32,
			NumItemsToPreemptivelyEvict: 10000,
		}

		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)
		require.NotNil(t, cache)

		addManyTransactionsWithUniformDistribution(cache, 1000, 500)
		require.Equal(t, 250000, int(cache.CountTx()))
	})
}

func Test_NotImplementedFunctions(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	evicted := cache.Put(nil, nil, 0)
	require.False(t, evicted)

	has, added := cache.HasOrAdd(nil, nil, 0)
	require.False(t, has)
	require.False(t, added)

	require.NotPanics(t, func() { cache.RegisterHandler(nil, "") })

	err := cache.Close()
	require.Nil(t, err)
}

func Test_IsInterfaceNil(t *testing.T) {
	cache := newUnconstrainedCacheToTest()
	require.False(t, check.IfNil(cache))

	makeNil := func() types.Cacher {
		return nil
	}

	thisIsNil := makeNil()
	require.True(t, check.IfNil(thisIsNil))
}

func TestTxCache_ConcurrentMutationAndSelection(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	// Alice will quickly move between two score buckets (chunks)
	cheapTransaction := createTx([]byte("alice-x-o"), "alice", 0).withDataLength(1).withGasLimit(300000000).withGasPrice(oneBillion)
	expensiveTransaction := createTx([]byte("alice-x-1"), "alice", 1).withDataLength(42).withGasLimit(50000000).withGasPrice(10 * oneBillion)
	cache.AddTx(cheapTransaction)
	cache.AddTx(expensiveTransaction)

	wg := sync.WaitGroup{}

	// Simulate selection
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			fmt.Println("Selection", i)
			_, _ = cache.SelectTransactions(math.MaxUint64, math.MaxInt)
		}

		wg.Done()
	}()

	// Simulate add / remove transactions
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			fmt.Println("Add / remove", i)
			cache.Remove([]byte("alice-x-1"))
			cache.AddTx(expensiveTransaction)
		}

		wg.Done()
	}()

	timedOut := waitTimeout(&wg, 1*time.Second)
	require.False(t, timedOut, "Timed out. Perhaps deadlock?")
}

func TestTxCache_TransactionIsAdded_EvenWhenInternalMapsAreInconsistent(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	// Setup inconsistency: transaction already exists in map by hash, but not in map by sender
	cache.txByHash.addTx(createTx([]byte("alice-x"), "alice", 42))

	require.Equal(t, 1, cache.txByHash.backingMap.Count())
	require.True(t, cache.Has([]byte("alice-x")))
	ok, added := cache.AddTx(createTx([]byte("alice-x"), "alice", 42))
	require.True(t, ok)
	require.True(t, added)
	require.Equal(t, uint64(1), cache.CountSenders())
	require.Equal(t, []string{"alice-x"}, cache.getHashesForSender("alice"))
	cache.Clear()

	// Setup inconsistency: transaction already exists in map by sender, but not in map by hash
	cache.txListBySender.addTxReturnEvicted(createTx([]byte("alice-x"), "alice", 42))

	require.False(t, cache.Has([]byte("alice-x")))
	ok, added = cache.AddTx(createTx([]byte("alice-x"), "alice", 42))
	require.True(t, ok)
	require.True(t, added)
	require.Equal(t, uint64(1), cache.CountSenders())
	require.Equal(t, []string{"alice-x"}, cache.getHashesForSender("alice"))
	cache.Clear()
}

func TestTxCache_NoCriticalInconsistency_WhenConcurrentAdditionsAndRemovals(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	// A lot of routines concur to add & remove THE FIRST transaction of a sender
	for try := 0; try < 100; try++ {
		var wg sync.WaitGroup

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				cache.AddTx(createTx([]byte("alice-x"), "alice", 42))
				_ = cache.RemoveTxByHash([]byte("alice-x"))
				wg.Done()
			}()
		}

		wg.Wait()
		// In this case, there is the slight chance that:
		// go A: add to map by hash
		// go B: won't add in map by hash, already there
		// go A: add to map by sender
		// go A: remove from map by hash
		// go A: remove from map by sender and delete empty sender
		// go B: add to map by sender
		// go B: can't remove from map by hash, not found
		// go B: won't remove from map by sender (sender unknown)

		// Therefore, the number of senders could be 0 or 1
		require.Equal(t, 0, cache.txByHash.backingMap.Count())
		expectedCountConsistent := 0
		expectedCountSlightlyInconsistent := 1
		actualCount := int(cache.txListBySender.backingMap.Count())
		require.True(t, actualCount == expectedCountConsistent || actualCount == expectedCountSlightlyInconsistent)

		// A further addition works:
		cache.AddTx(createTx([]byte("alice-x"), "alice", 42))
		require.True(t, cache.Has([]byte("alice-x")))
		require.Equal(t, []string{"alice-x"}, cache.getHashesForSender("alice"))
	}

	cache.Clear()

	// A lot of routines concur to add & remove subsequent transactions of a sender
	cache.AddTx(createTx([]byte("alice-w"), "alice", 41))

	for try := 0; try < 100; try++ {
		var wg sync.WaitGroup

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				cache.AddTx(createTx([]byte("alice-x"), "alice", 42))
				_ = cache.RemoveTxByHash([]byte("alice-x"))
				wg.Done()
			}()
		}

		wg.Wait()

		// In this case, there is the slight chance that:
		// go A: add to map by hash
		// go B: won't add in map by hash, already there
		// go A: add to map by sender (existing sender/list)
		// go A: remove from map by hash
		// go A: remove from map by sender
		// go B: add to map by sender (existing sender/list)
		// go B: can't remove from map by hash, not found
		// go B: won't remove from map by sender (sender unknown)

		// Therefore, Alice may have one or two transactions in her list.
		require.Equal(t, 1, cache.txByHash.backingMap.Count())
		expectedTxsConsistent := []string{"alice-w"}
		expectedTxsSlightlyInconsistent := []string{"alice-w", "alice-x"}
		actualTxs := cache.getHashesForSender("alice")
		require.True(t, assert.ObjectsAreEqual(expectedTxsConsistent, actualTxs) || assert.ObjectsAreEqual(expectedTxsSlightlyInconsistent, actualTxs))

		// A further addition works:
		cache.AddTx(createTx([]byte("alice-x"), "alice", 42))
		require.True(t, cache.Has([]byte("alice-w")))
		require.True(t, cache.Has([]byte("alice-x")))
		require.Equal(t, []string{"alice-w", "alice-x"}, cache.getHashesForSender("alice"))
	}

	cache.Clear()
}

func newUnconstrainedCacheToTest() *TxCache {
	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	cache, err := NewTxCache(ConfigSourceMe{
		Name:                        "test",
		NumChunks:                   16,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
		CountThreshold:              math.MaxUint32,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
	}, txGasHandler)
	if err != nil {
		panic(fmt.Sprintf("newUnconstrainedCacheToTest(): %s", err))
	}

	return cache
}

func newCacheToTest(numBytesPerSenderThreshold uint32, countPerSenderThreshold uint32) *TxCache {
	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	cache, err := NewTxCache(ConfigSourceMe{
		Name:                        "test",
		NumChunks:                   16,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  numBytesPerSenderThreshold,
		CountThreshold:              math.MaxUint32,
		CountPerSenderThreshold:     countPerSenderThreshold,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
	}, txGasHandler)
	if err != nil {
		panic(fmt.Sprintf("newCacheToTest(): %s", err))
	}

	return cache
}
