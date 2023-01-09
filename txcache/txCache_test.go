package txcache

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go-storage/common"
	"github.com/multiversx/mx-chain-go-storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewTxCache(t *testing.T) {
	config := ConfigSourceMe{
		Name:                       "test",
		NumChunks:                  16,
		NumBytesPerSenderThreshold: maxNumBytesPerSenderUpperBound,
		CountPerSenderThreshold:    math.MaxUint32,
	}

	withEvictionConfig := ConfigSourceMe{
		Name:                          "test",
		NumChunks:                     16,
		NumBytesPerSenderThreshold:    maxNumBytesPerSenderUpperBound,
		CountPerSenderThreshold:       math.MaxUint32,
		EvictionEnabled:               true,
		NumBytesThreshold:             maxNumBytesUpperBound,
		CountThreshold:                math.MaxUint32,
		NumSendersToPreemptivelyEvict: 100,
	}
	txGasHandler, _ := dummyParams()

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

	badConfig = withEvictionConfig
	badConfig.NumBytesThreshold = 0
	requireErrorOnNewTxCache(t, badConfig, common.ErrInvalidConfig, "config.NumBytesThreshold", txGasHandler)

	badConfig = withEvictionConfig
	badConfig.CountThreshold = 0
	requireErrorOnNewTxCache(t, badConfig, common.ErrInvalidConfig, "config.CountThreshold", txGasHandler)

	badConfig = withEvictionConfig
	badConfig.NumSendersToPreemptivelyEvict = 0
	requireErrorOnNewTxCache(t, badConfig, common.ErrInvalidConfig, "config.NumSendersToPreemptivelyEvict", txGasHandler)
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

	cache.AddTx(createTxWithParams([]byte("tx-alice-1"), "alice", 1, 128, 42, 42))
	cache.AddTx(createTxWithParams([]byte("tx-alice-2"), "alice", 2, 512, 42, 42))
	cache.AddTx(createTxWithParams([]byte("tx-alice-4"), "alice", 3, 256, 42, 42))
	cache.AddTx(createTxWithParams([]byte("tx-bob-1"), "bob", 1, 512, 42, 42))
	cache.AddTx(createTxWithParams([]byte("tx-bob-2"), "bob", 2, 513, 42, 42))

	require.Equal(t, []string{"tx-alice-1", "tx-alice-2", "tx-alice-4"}, cache.getHashesForSender("alice"))
	require.Equal(t, []string{"tx-bob-1"}, cache.getHashesForSender("bob"))
	require.True(t, cache.areInternalMapsConsistent())

	cache.AddTx(createTxWithParams([]byte("tx-alice-3"), "alice", 3, 256, 42, 42))
	cache.AddTx(createTxWithParams([]byte("tx-bob-2"), "bob", 3, 512, 42, 42))
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

func Test_SelectTransactions_Dummy(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4))
	cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))
	cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7))
	cache.AddTx(createTx([]byte("hash-bob-6"), "bob", 6))
	cache.AddTx(createTx([]byte("hash-bob-5"), "bob", 5))
	cache.AddTx(createTx([]byte("hash-carol-1"), "carol", 1))

	sorted := cache.SelectTransactionsWithBandwidth(10, 2, math.MaxUint64)
	require.Len(t, sorted, 8)
}

func Test_SelectTransactionsWithBandwidth_Dummy(t *testing.T) {
	cache := newUnconstrainedCacheToTest()
	cache.AddTx(createTxWithGasLimit([]byte("hash-alice-4"), "alice", 4, 100000))
	cache.AddTx(createTxWithGasLimit([]byte("hash-alice-3"), "alice", 3, 100000))
	cache.AddTx(createTxWithGasLimit([]byte("hash-alice-2"), "alice", 2, 500000))
	cache.AddTx(createTxWithGasLimit([]byte("hash-alice-1"), "alice", 1, 200000))
	cache.AddTx(createTxWithGasLimit([]byte("hash-bob-7"), "bob", 7, 100000))
	cache.AddTx(createTxWithGasLimit([]byte("hash-bob-6"), "bob", 6, 50000))
	cache.AddTx(createTxWithGasLimit([]byte("hash-bob-5"), "bob", 5, 50000))
	cache.AddTx(createTxWithGasLimit([]byte("hash-carol-1"), "carol", 1, 50000))

	sorted := cache.SelectTransactionsWithBandwidth(5, 2, 200000)
	numSelected := 1 + 1 + 3 // 1 alice, 1 carol, 3 bob

	require.Len(t, sorted, numSelected)
}

func Test_SelectTransactions_BreaksAtNonceGaps(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
	cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))
	cache.AddTx(createTx([]byte("hash-alice-5"), "alice", 5))
	cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 42))
	cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 44))
	cache.AddTx(createTx([]byte("hash-bob-45"), "bob", 45))
	cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
	cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8))
	cache.AddTx(createTx([]byte("hash-carol-10"), "carol", 10))
	cache.AddTx(createTx([]byte("hash-carol-11"), "carol", 11))

	numSelected := 3 + 1 + 2 // 3 alice + 1 bob + 2 carol

	sorted := cache.SelectTransactionsWithBandwidth(10, 2, math.MaxUint64)
	require.Len(t, sorted, numSelected)
}

func Test_SelectTransactions(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	// Add "nSenders" * "nTransactionsPerSender" transactions in the cache (in reversed nonce order)
	nSenders := 1000
	nTransactionsPerSender := 100
	nTotalTransactions := nSenders * nTransactionsPerSender
	nRequestedTransactions := math.MaxInt16

	for senderTag := 0; senderTag < nSenders; senderTag++ {
		sender := fmt.Sprintf("sender:%d", senderTag)

		for txNonce := nTransactionsPerSender; txNonce > 0; txNonce-- {
			txHash := fmt.Sprintf("hash:%d:%d", senderTag, txNonce)
			tx := createTx([]byte(txHash), sender, uint64(txNonce))
			cache.AddTx(tx)
		}
	}

	require.Equal(t, uint64(nTotalTransactions), cache.CountTx())

	sorted := cache.SelectTransactionsWithBandwidth(nRequestedTransactions, 2, math.MaxUint64)

	require.Len(t, sorted, core.MinInt(nRequestedTransactions, nTotalTransactions))

	// Check order
	nonces := make(map[string]uint64, nSenders)
	for _, tx := range sorted {
		nonce := tx.Tx.GetNonce()
		sender := string(tx.Tx.GetSndAddr())
		previousNonce := nonces[sender]

		require.LessOrEqual(t, previousNonce, nonce)
		nonces[sender] = nonce
	}
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
	txGasHandler, _ := dummyParams()
	config := ConfigSourceMe{
		Name:                          "untitled",
		NumChunks:                     16,
		EvictionEnabled:               true,
		NumBytesThreshold:             maxNumBytesUpperBound,
		CountThreshold:                100,
		NumSendersToPreemptivelyEvict: 1,
		NumBytesPerSenderThreshold:    maxNumBytesPerSenderUpperBound,
		CountPerSenderThreshold:       math.MaxUint32,
	}

	// 11 * 10
	cache, err := NewTxCache(config, txGasHandler)
	require.Nil(t, err)
	require.NotNil(t, cache)

	addManyTransactionsWithUniformDistribution(cache, 11, 10)
	require.LessOrEqual(t, cache.CountTx(), uint64(100))

	config = ConfigSourceMe{
		Name:                          "untitled",
		NumChunks:                     16,
		EvictionEnabled:               true,
		NumBytesThreshold:             maxNumBytesUpperBound,
		CountThreshold:                250000,
		NumSendersToPreemptivelyEvict: 1,
		NumBytesPerSenderThreshold:    maxNumBytesPerSenderUpperBound,
		CountPerSenderThreshold:       math.MaxUint32,
	}

	// 100 * 1000
	cache, err = NewTxCache(config, txGasHandler)
	require.Nil(t, err)
	require.NotNil(t, cache)

	addManyTransactionsWithUniformDistribution(cache, 100, 1000)
	require.LessOrEqual(t, cache.CountTx(), uint64(250000))
}

func Test_NotImplementedFunctions(t *testing.T) {
	cache := newUnconstrainedCacheToTest()

	evicted := cache.Put(nil, nil, 0)
	require.False(t, evicted)

	has, added := cache.HasOrAdd(nil, nil, 0)
	require.False(t, has)
	require.False(t, added)

	require.NotPanics(t, func() { cache.RegisterHandler(nil, "") })
	require.Zero(t, cache.MaxSize())

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
	cheapTransaction := createTxWithParams([]byte("alice-x-o"), "alice", 0, 128, 50000, 100*oneBillion)
	expensiveTransaction := createTxWithParams([]byte("alice-x-1"), "alice", 1, 128, 50000, 300*oneBillion)
	cache.AddTx(cheapTransaction)
	cache.AddTx(expensiveTransaction)

	wg := sync.WaitGroup{}

	// Simulate selection
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			fmt.Println("Selection", i)
			cache.SelectTransactionsWithBandwidth(100, 100, math.MaxUint64)
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
	cache.txListBySender.addTx(createTx([]byte("alice-x"), "alice", 42))

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
	txGasHandler, _ := dummyParams()
	cache, err := NewTxCache(ConfigSourceMe{
		Name:                       "test",
		NumChunks:                  16,
		NumBytesPerSenderThreshold: maxNumBytesPerSenderUpperBound,
		CountPerSenderThreshold:    math.MaxUint32,
	}, txGasHandler)
	if err != nil {
		panic(fmt.Sprintf("newUnconstrainedCacheToTest(): %s", err))
	}

	return cache
}

func newCacheToTest(numBytesPerSenderThreshold uint32, countPerSenderThreshold uint32) *TxCache {
	txGasHandler, _ := dummyParams()
	cache, err := NewTxCache(ConfigSourceMe{
		Name:                       "test",
		NumChunks:                  16,
		NumBytesPerSenderThreshold: numBytesPerSenderThreshold,
		CountPerSenderThreshold:    countPerSenderThreshold,
	}, txGasHandler)
	if err != nil {
		panic(fmt.Sprintf("newCacheToTest(): %s", err))
	}

	return cache
}
