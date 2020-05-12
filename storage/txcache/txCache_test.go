package txcache

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/require"
)

func Test_NewTxCache(t *testing.T) {
	config := CacheConfig{
		Name:                       "test",
		NumChunksHint:              16,
		NumBytesPerSenderThreshold: math.MaxUint32,
		CountPerSenderThreshold:    math.MaxUint32,
		MinGasPriceMicroErd:        100,
	}

	evictionConfig := CacheConfig{
		Name:                       "test",
		NumChunksHint:              16,
		NumBytesPerSenderThreshold: math.MaxUint32,
		CountPerSenderThreshold:    math.MaxUint32,
		MinGasPriceMicroErd:        100,
		EvictionEnabled:            true,
		NumBytesThreshold:          math.MaxUint32,
		CountThreshold:             math.MaxUint32,
		NumSendersToEvictInOneStep: 100,
	}

	cache, err := NewTxCache(config)
	require.Nil(t, err)
	require.NotNil(t, cache)

	badConfig := config
	badConfig.Name = ""
	requireErrorOnNewTxCache(t, badConfig, "config.Name")

	badConfig = config
	badConfig.NumChunksHint = 0
	requireErrorOnNewTxCache(t, badConfig, "config.NumChunksHint")

	badConfig = config
	badConfig.NumBytesPerSenderThreshold = 0
	requireErrorOnNewTxCache(t, badConfig, "config.NumBytesPerSenderThreshold")

	badConfig = config
	badConfig.CountPerSenderThreshold = 0
	requireErrorOnNewTxCache(t, badConfig, "config.CountPerSenderThreshold")

	badConfig = config
	badConfig.MinGasPriceMicroErd = 0
	requireErrorOnNewTxCache(t, badConfig, "config.MinGasPriceMicroErd")

	badConfig = evictionConfig
	badConfig.NumBytesThreshold = 0
	requireErrorOnNewTxCache(t, badConfig, "config.NumBytesThreshold")

	badConfig = evictionConfig
	badConfig.CountThreshold = 0
	requireErrorOnNewTxCache(t, badConfig, "config.CountThreshold")

	badConfig = evictionConfig
	badConfig.NumSendersToEvictInOneStep = 0
	requireErrorOnNewTxCache(t, badConfig, "config.NumSendersToEvictInOneStep")
}

func requireErrorOnNewTxCache(t *testing.T, config CacheConfig, errMessage string) {
	cache, err := NewTxCache(config)
	require.Contains(t, err.Error(), errMessage)
	require.Nil(t, cache)
}

func Test_AddTx(t *testing.T) {
	cache := newCacheToTest()

	tx := createTx([]byte("hash-1"), "alice", 1)

	ok, added := cache.AddTx(tx)
	require.True(t, ok)
	require.True(t, added)

	// Add it again (no-operation)
	ok, added = cache.AddTx(tx)
	require.True(t, ok)
	require.False(t, added)

	foundTx, ok := cache.GetByTxHash([]byte("hash-1"))
	require.True(t, ok)
	require.Equal(t, tx, foundTx)
}

func Test_AddNilTx_DoesNothing(t *testing.T) {
	cache := newCacheToTest()

	txHash := []byte("hash-1")

	ok, added := cache.AddTx(&WrappedTransaction{Tx: nil, TxHash: txHash})
	require.False(t, ok)
	require.False(t, added)

	foundTx, ok := cache.GetByTxHash(txHash)
	require.False(t, ok)
	require.Nil(t, foundTx)
}

func Test_RemoveByTxHash(t *testing.T) {
	cache := newCacheToTest()

	cache.AddTx(createTx([]byte("hash-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-2"), "alice", 2))

	err := cache.RemoveTxByHash([]byte("hash-1"))
	require.Nil(t, err)
	cache.Remove([]byte("hash-2"))

	foundTx, ok := cache.GetByTxHash([]byte("hash-1"))
	require.False(t, ok)
	require.Nil(t, foundTx)

	foundTx, ok = cache.GetByTxHash([]byte("hash-2"))
	require.False(t, ok)
	require.Nil(t, foundTx)
}

func Test_CountTx_And_Len(t *testing.T) {
	cache := newCacheToTest()

	cache.AddTx(createTx([]byte("hash-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-2"), "alice", 2))
	cache.AddTx(createTx([]byte("hash-3"), "alice", 3))

	require.Equal(t, int64(3), cache.CountTx())
	require.Equal(t, 3, cache.Len())
}

func Test_GetByTxHash_And_Peek_And_Get(t *testing.T) {
	cache := newCacheToTest()

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

func Test_RemoveByTxHash_Error_WhenMissing(t *testing.T) {
	cache := newCacheToTest()
	err := cache.RemoveTxByHash([]byte("missing"))
	require.Equal(t, err, errTxNotFound)
}

func Test_RemoveByTxHash_Error_WhenMapsInconsistency(t *testing.T) {
	cache := newCacheToTest()

	txHash := []byte("hash-1")
	tx := createTx(txHash, "alice", 1)
	cache.AddTx(tx)

	// Cause an inconsistency between the two internal maps (theoretically possible in case of misbehaving eviction)
	cache.txListBySender.removeTx(tx)

	err := cache.RemoveTxByHash(txHash)
	require.Equal(t, err, errMapsSyncInconsistency)
}

func Test_Clear(t *testing.T) {
	cache := newCacheToTest()

	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7))
	cache.AddTx(createTx([]byte("hash-alice-42"), "alice", 42))
	require.Equal(t, int64(3), cache.CountTx())

	cache.Clear()
	require.Equal(t, int64(0), cache.CountTx())
}

func Test_ForEachTransaction(t *testing.T) {
	cache := newCacheToTest()

	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7))

	counter := 0
	cache.ForEachTransaction(func(txHash []byte, value *WrappedTransaction) {
		counter++
	})
	require.Equal(t, 2, counter)
}

func Test_SelectTransactions_Dummy(t *testing.T) {
	cache := newCacheToTest()

	cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4))
	cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))
	cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
	cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7))
	cache.AddTx(createTx([]byte("hash-bob-6"), "bob", 6))
	cache.AddTx(createTx([]byte("hash-bob-5"), "bob", 5))
	cache.AddTx(createTx([]byte("hash-carol-1"), "carol", 1))

	sorted := cache.SelectTransactions(10, 2)
	require.Len(t, sorted, 8)
}

func Test_SelectTransactions_BreaksAtNonceGaps(t *testing.T) {
	cache := newCacheToTest()

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

	sorted := cache.SelectTransactions(10, 2)
	require.Len(t, sorted, numSelected)
}

func Test_SelectTransactions(t *testing.T) {
	cache := newCacheToTest()

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

	require.Equal(t, int64(nTotalTransactions), cache.CountTx())

	sorted := cache.SelectTransactions(nRequestedTransactions, 2)

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
	cache := newCacheToTest()

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
	config := CacheConfig{
		Name:                       "untitled",
		NumChunksHint:              16,
		EvictionEnabled:            true,
		NumBytesThreshold:          math.MaxUint32,
		CountThreshold:             100,
		NumSendersToEvictInOneStep: 1,
		NumBytesPerSenderThreshold: math.MaxUint32,
		CountPerSenderThreshold:    math.MaxUint32,
		MinGasPriceMicroErd:        100,
	}

	// 11 * 10
	cache, err := NewTxCache(config)
	require.Nil(t, err)
	require.NotNil(t, cache)

	addManyTransactionsWithUniformDistribution(cache, 11, 10)
	require.LessOrEqual(t, cache.CountTx(), int64(100))

	config = CacheConfig{
		Name:                       "untitled",
		NumChunksHint:              16,
		EvictionEnabled:            true,
		NumBytesThreshold:          math.MaxUint32,
		CountThreshold:             250000,
		NumSendersToEvictInOneStep: 1,
		NumBytesPerSenderThreshold: math.MaxUint32,
		CountPerSenderThreshold:    math.MaxUint32,
		MinGasPriceMicroErd:        100,
	}

	// 100 * 1000
	cache, err = NewTxCache(config)
	require.Nil(t, err)
	require.NotNil(t, cache)

	addManyTransactionsWithUniformDistribution(cache, 100, 1000)
	require.LessOrEqual(t, cache.CountTx(), int64(250000))
}

func Test_NotImplementedFunctions(t *testing.T) {
	cache := newCacheToTest()

	evicted := cache.Put(nil, nil)
	require.False(t, evicted)

	has := cache.Has(nil)
	require.False(t, has)

	ok, evicted := cache.HasOrAdd(nil, nil)
	require.False(t, ok)
	require.False(t, evicted)

	require.NotPanics(t, func() { cache.RemoveOldest() })
	require.NotPanics(t, func() { cache.RegisterHandler(nil) })
	require.Zero(t, cache.MaxSize())
}

func Test_IsInterfaceNil(t *testing.T) {
	cache := newCacheToTest()
	require.False(t, check.IfNil(cache))

	makeNil := func() storage.Cacher {
		return nil
	}

	thisIsNil := makeNil()
	require.True(t, check.IfNil(thisIsNil))
}

func TestTxCache_ConcurrentMutationAndSelection(t *testing.T) {
	cache := newCacheToTest()

	// Alice will quickly move between two score buckets (chunks)
	cheapTransaction := createTxWithParams([]byte("alice-x-o"), "alice", 0, 128, 50000, 100*oneTrilion)
	expensiveTransaction := createTxWithParams([]byte("alice-x-1"), "alice", 1, 128, 50000, 300*oneTrilion)
	cache.AddTx(cheapTransaction)
	cache.AddTx(expensiveTransaction)

	wg := sync.WaitGroup{}

	// Simulate selection
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			fmt.Println("Selection", i)
			cache.SelectTransactions(100, 100)
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

func newCacheToTest() *TxCache {
	cache, err := NewTxCache(CacheConfig{
		Name:                       "test",
		NumChunksHint:              16,
		NumBytesPerSenderThreshold: math.MaxUint32,
		CountPerSenderThreshold:    math.MaxUint32,
		MinGasPriceMicroErd:        100,
	})
	if err != nil {
		panic(fmt.Sprintf("newCacheToTest(): %s", err))
	}

	return cache
}
