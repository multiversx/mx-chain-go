package txcache

import (
	"fmt"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestTxCache_DoEviction_BecauseOfCount(t *testing.T) {
	config := ConfigSourceMe{
		Name:                        "untitled",
		NumChunks:                   16,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
		CountThreshold:              4,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             true,
		NumItemsToPreemptivelyEvict: 1,
	}

	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	cache, err := NewTxCache(config, txGasHandler)
	require.Nil(t, err)
	require.NotNil(t, cache)

	cache.AddTx(createTx([]byte("hash-alice"), "alice", 1).withGasPrice(1 * oneBillion))
	cache.AddTx(createTx([]byte("hash-bob"), "bob", 1).withGasPrice(2 * oneBillion))
	cache.AddTx(createTx([]byte("hash-carol"), "carol", 1).withGasPrice(3 * oneBillion))
	cache.AddTx(createTx([]byte("hash-eve"), "eve", 1).withGasPrice(4 * oneBillion))
	cache.AddTx(createTx([]byte("hash-dan"), "dan", 1).withGasPrice(5 * oneBillion))

	journal := cache.doEviction()
	require.Equal(t, 1, journal.numEvicted)
	require.Equal(t, []int{1}, journal.numEvictedByPass)

	// Alice and Bob evicted. Carol still there (better score).
	_, ok := cache.GetByTxHash([]byte("hash-carol"))
	require.True(t, ok)
	require.Equal(t, uint64(4), cache.CountSenders())
	require.Equal(t, uint64(4), cache.CountTx())
}

func TestTxCache_DoEviction_BecauseOfSize(t *testing.T) {
	config := ConfigSourceMe{
		Name:                        "untitled",
		NumChunks:                   16,
		NumBytesThreshold:           1000,
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

	cache.AddTx(createTx([]byte("hash-alice"), "alice", 1).withSize(256).withGasLimit(500000))
	cache.AddTx(createTx([]byte("hash-bob"), "bob", 1).withSize(256).withGasLimit(500000))
	cache.AddTx(createTx([]byte("hash-carol"), "carol", 1).withSize(256).withGasLimit(500000).withGasPrice(1.5 * oneBillion))
	cache.AddTx(createTx([]byte("hash-eve"), "eve", 1).withSize(256).withGasLimit(500000).withGasPrice(3 * oneBillion))

	journal := cache.doEviction()
	require.Equal(t, 1, journal.numEvicted)
	require.Equal(t, []int{1}, journal.numEvictedByPass)

	// Alice and Bob evicted (lower score). Carol and Eve still there.
	_, ok := cache.GetByTxHash([]byte("hash-carol"))
	require.True(t, ok)
	_, ok = cache.GetByTxHash([]byte("hash-eve"))
	require.True(t, ok)
	require.Equal(t, uint64(3), cache.CountSenders())
	require.Equal(t, uint64(3), cache.CountTx())
}

func TestTxCache_DoEviction_DoesNothingWhenAlreadyInProgress(t *testing.T) {
	config := ConfigSourceMe{
		Name:                        "untitled",
		NumChunks:                   1,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
		CountThreshold:              4,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             true,
		NumItemsToPreemptivelyEvict: 1,
	}

	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	cache, err := NewTxCache(config, txGasHandler)
	require.Nil(t, err)
	require.NotNil(t, cache)

	_ = cache.isEvictionInProgress.SetReturningPrevious()

	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", uint64(1)))
	cache.AddTx(createTx([]byte("hash-alice-2"), "alice", uint64(2)))
	cache.AddTx(createTx([]byte("hash-alice-3"), "alice", uint64(3)))
	cache.AddTx(createTx([]byte("hash-alice-4"), "alice", uint64(4)))
	cache.AddTx(createTx([]byte("hash-alice-5"), "alice", uint64(5)))

	// Nothing is evicted because eviction is already in progress.
	journal := cache.doEviction()
	require.Nil(t, journal)
	require.Equal(t, uint64(5), cache.CountTx())

	cache.isEvictionInProgress.Reset()

	// Now eviction can happen.
	journal = cache.doEviction()
	require.NotNil(t, journal)
	require.Equal(t, 1, journal.numEvicted)
	require.Equal(t, 4, int(cache.CountTx()))
}

func TestBenchmarkTxCache_DoEviction(t *testing.T) {
	config := ConfigSourceMe{
		Name:                        "untitled",
		NumChunks:                   16,
		NumBytesThreshold:           1000000000,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
		CountThreshold:              300000,
		CountPerSenderThreshold:     math.MaxUint32,
		NumItemsToPreemptivelyEvict: 50000,
	}

	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	sw := core.NewStopWatch()

	t.Run("numSenders = 35000, numTransactions = 10", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		cache.config.EvictionEnabled = false
		addManyTransactionsWithUniformDistribution(cache, 35000, 10)
		cache.config.EvictionEnabled = true

		require.Equal(t, uint64(350000), cache.CountTx())

		sw.Start(t.Name())
		journal := cache.doEviction()
		sw.Stop(t.Name())

		require.Equal(t, 50000, journal.numEvicted)
		require.Equal(t, 1, len(journal.numEvictedByPass))
	})

	t.Run("numSenders = 100000, numTransactions = 5", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		cache.config.EvictionEnabled = false
		addManyTransactionsWithUniformDistribution(cache, 100000, 5)
		cache.config.EvictionEnabled = true

		require.Equal(t, uint64(500000), cache.CountTx())

		sw.Start(t.Name())
		journal := cache.doEviction()
		sw.Stop(t.Name())

		require.Equal(t, 200000, journal.numEvicted)
		require.Equal(t, 4, len(journal.numEvictedByPass))
	})

	t.Run("numSenders = 400000, numTransactions = 1", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		cache.config.EvictionEnabled = false
		addManyTransactionsWithUniformDistribution(cache, 400000, 1)
		cache.config.EvictionEnabled = true

		require.Equal(t, uint64(400000), cache.CountTx())

		sw.Start(t.Name())
		journal := cache.doEviction()
		sw.Stop(t.Name())

		require.Equal(t, 100000, journal.numEvicted)
		require.Equal(t, 2, len(journal.numEvictedByPass))
	})

	t.Run("numSenders = 10000, numTransactions = 100", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		cache.config.EvictionEnabled = false
		addManyTransactionsWithUniformDistribution(cache, 10000, 100)
		cache.config.EvictionEnabled = true

		require.Equal(t, uint64(1000000), cache.CountTx())

		sw.Start(t.Name())
		journal := cache.doEviction()
		sw.Stop(t.Name())

		require.Equal(t, 700000, journal.numEvicted)
		require.Equal(t, 14, len(journal.numEvictedByPass))
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}

	// (1)
	// Vendor ID:                GenuineIntel
	//   Model name:             11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
	//     CPU family:           6
	//     Model:                140
	//     Thread(s) per core:   2
	//     Core(s) per socket:   4
	//
	// 0.093771s (TestBenchmarkTxCache_DoEviction_Benchmark/numSenders_=_35000,_numTransactions_=_10)
	// 0.424683s (TestBenchmarkTxCache_DoEviction_Benchmark/numSenders_=_100000,_numTransactions_=_5)
	// 0.448017s (TestBenchmarkTxCache_DoEviction_Benchmark/numSenders_=_10000,_numTransactions_=_100)
	// 0.476738s (TestBenchmarkTxCache_DoEviction_Benchmark/numSenders_=_400000,_numTransactions_=_1)
}
