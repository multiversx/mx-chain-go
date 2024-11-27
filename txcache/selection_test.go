package txcache

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestTxCache_SelectTransactions_Dummy(t *testing.T) {
	t.Run("all having same PPU", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 5)
		session.SetNonce([]byte("carol"), 1)

		cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7))
		cache.AddTx(createTx([]byte("hash-bob-6"), "bob", 6))
		cache.AddTx(createTx([]byte("hash-bob-5"), "bob", 5))
		cache.AddTx(createTx([]byte("hash-carol-1"), "carol", 1))

		selected, accumulatedGas := cache.SelectTransactions(session, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		require.Len(t, selected, 8)
		require.Equal(t, 400000, int(accumulatedGas))

		// Check order
		require.Equal(t, "hash-alice-1", string(selected[0].TxHash))
		require.Equal(t, "hash-alice-2", string(selected[1].TxHash))
		require.Equal(t, "hash-alice-3", string(selected[2].TxHash))
		require.Equal(t, "hash-alice-4", string(selected[3].TxHash))
		require.Equal(t, "hash-bob-5", string(selected[4].TxHash))
		require.Equal(t, "hash-bob-6", string(selected[5].TxHash))
		require.Equal(t, "hash-bob-7", string(selected[6].TxHash))
		require.Equal(t, "hash-carol-1", string(selected[7].TxHash))
	})

	t.Run("alice > carol > bob", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 5)
		session.SetNonce([]byte("carol"), 3)

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withGasPrice(100))
		cache.AddTx(createTx([]byte("hash-bob-5"), "bob", 5).withGasPrice(50))
		cache.AddTx(createTx([]byte("hash-carol-3"), "carol", 3).withGasPrice(75))

		selected, accumulatedGas := cache.SelectTransactions(session, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		require.Len(t, selected, 3)
		require.Equal(t, 150000, int(accumulatedGas))

		// Check order
		require.Equal(t, "hash-alice-1", string(selected[0].TxHash))
		require.Equal(t, "hash-carol-3", string(selected[1].TxHash))
		require.Equal(t, "hash-bob-5", string(selected[2].TxHash))
	})
}

func TestTxCache_SelectTransactionsWithBandwidth_Dummy(t *testing.T) {
	t.Run("transactions with no data field", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 5)
		session.SetNonce([]byte("carol"), 1)

		cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4).withGasLimit(100000))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3).withGasLimit(100000))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withGasLimit(500000))
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withGasLimit(200000))
		cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7).withGasLimit(400000))
		cache.AddTx(createTx([]byte("hash-bob-6"), "bob", 6).withGasLimit(50000))
		cache.AddTx(createTx([]byte("hash-bob-5"), "bob", 5).withGasLimit(50000))
		cache.AddTx(createTx([]byte("hash-carol-1"), "carol", 1).withGasLimit(50000))

		selected, accumulatedGas := cache.SelectTransactions(session, 760000, math.MaxInt, selectionLoopMaximumDuration)
		require.Len(t, selected, 5)
		require.Equal(t, 750000, int(accumulatedGas))

		// Check order
		require.Equal(t, "hash-bob-5", string(selected[0].TxHash))
		require.Equal(t, "hash-bob-6", string(selected[1].TxHash))
		require.Equal(t, "hash-carol-1", string(selected[2].TxHash))
		require.Equal(t, "hash-alice-1", string(selected[3].TxHash))
		require.Equal(t, "hash-bob-7", string(selected[4].TxHash))
	})
}

func TestTxCache_SelectTransactions_HandlesNotExecutableTransactions(t *testing.T) {
	t.Run("with middle gaps", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 42)
		session.SetNonce([]byte("carol"), 7)

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))
		cache.AddTx(createTx([]byte("hash-alice-5"), "alice", 5)) // gap
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 42))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 44)) // gap
		cache.AddTx(createTx([]byte("hash-bob-45"), "bob", 45))
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8))
		cache.AddTx(createTx([]byte("hash-carol-10"), "carol", 10)) // gap
		cache.AddTx(createTx([]byte("hash-carol-11"), "carol", 11))

		sorted, accumulatedGas := cache.SelectTransactions(session, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		expectedNumSelected := 3 + 1 + 2 // 3 alice + 1 bob + 2 carol
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 300000, int(accumulatedGas))
	})

	t.Run("with initial gaps", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 42)
		session.SetNonce([]byte("carol"), 7)

		// Good
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))

		// Initial gap
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 44))
		cache.AddTx(createTx([]byte("hash-bob-43"), "bob", 45))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 46))

		// Good
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8))

		sorted, accumulatedGas := cache.SelectTransactions(session, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		expectedNumSelected := 3 + 0 + 2 // 3 alice + 0 bob + 2 carol
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 250000, int(accumulatedGas))
	})

	t.Run("with lower nonces", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 42)
		session.SetNonce([]byte("carol"), 7)

		// Good
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))

		// A few with lower nonce
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 40))
		cache.AddTx(createTx([]byte("hash-bob-43"), "bob", 41))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 42))

		// Good
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8))

		sorted, accumulatedGas := cache.SelectTransactions(session, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		expectedNumSelected := 3 + 1 + 2 // 3 alice + 1 bob + 2 carol
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 300000, int(accumulatedGas))
	})

	t.Run("with duplicated nonces", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3a"), "alice", 3))
		cache.AddTx(createTx([]byte("hash-alice-3b"), "alice", 3).withGasPrice(oneBillion * 2))
		cache.AddTx(createTx([]byte("hash-alice-3c"), "alice", 3))
		cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4))

		sorted, accumulatedGas := cache.SelectTransactions(session, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		require.Len(t, sorted, 4)
		require.Equal(t, 200000, int(accumulatedGas))

		require.Equal(t, "hash-alice-1", string(sorted[0].TxHash))
		require.Equal(t, "hash-alice-2", string(sorted[1].TxHash))
		require.Equal(t, "hash-alice-3b", string(sorted[2].TxHash))
		require.Equal(t, "hash-alice-4", string(sorted[3].TxHash))
	})

	t.Run("with fee exceeding balance", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetBalance([]byte("alice"), big.NewInt(150000000000000))
		session.SetNonce([]byte("bob"), 42)
		session.SetBalance([]byte("bob"), big.NewInt(70000000000000))

		// Enough balance
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))

		// Not enough balance
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 40))
		cache.AddTx(createTx([]byte("hash-bob-43"), "bob", 41))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 42))

		sorted, accumulatedGas := cache.SelectTransactions(session, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		expectedNumSelected := 3 + 1 // 3 alice + 1 bob
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 200000, int(accumulatedGas))
	})

	t.Run("with badly guarded", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 42)

		session.IsBadlyGuardedCalled = func(tx data.TransactionHandler) bool {
			if bytes.Equal(tx.GetData(), []byte("t")) {
				return true
			}

			return false
		}

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withData([]byte("x")).withGasLimit(100000))
		cache.AddTx(createTx([]byte("hash-bob-42a"), "bob", 42).withData([]byte("y")).withGasLimit(100000))
		cache.AddTx(createTx([]byte("hash-bob-43a"), "bob", 43).withData([]byte("z")).withGasLimit(100000))
		cache.AddTx(createTx([]byte("hash-bob-43b"), "bob", 43).withData([]byte("t")).withGasLimit(100000))

		sorted, accumulatedGas := cache.SelectTransactions(session, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		require.Len(t, sorted, 3)
		require.Equal(t, 300000, int(accumulatedGas))

		require.Equal(t, "hash-alice-1", string(sorted[0].TxHash))
		require.Equal(t, "hash-bob-42a", string(sorted[1].TxHash))
		require.Equal(t, "hash-bob-43a", string(sorted[2].TxHash))
	})
}

func TestTxCache_SelectTransactions_WhenTransactionsAddedInReversedNonceOrder(t *testing.T) {
	cache := newUnconstrainedCacheToTest()
	session := txcachemocks.NewSelectionSessionMock()

	// Add "nSenders" * "nTransactionsPerSender" transactions in the cache (in reversed nonce order)
	nSenders := 1000
	nTransactionsPerSender := 100
	nTotalTransactions := nSenders * nTransactionsPerSender

	for senderTag := 0; senderTag < nSenders; senderTag++ {
		sender := fmt.Sprintf("sender:%d", senderTag)

		for txNonce := nTransactionsPerSender - 1; txNonce >= 0; txNonce-- {
			txHash := fmt.Sprintf("hash:%d:%d", senderTag, txNonce)
			tx := createTx([]byte(txHash), sender, uint64(txNonce))
			cache.AddTx(tx)
		}
	}

	require.Equal(t, uint64(nTotalTransactions), cache.CountTx())

	sorted, accumulatedGas := cache.SelectTransactions(session, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
	require.Len(t, sorted, nTotalTransactions)
	require.Equal(t, 5_000_000_000, int(accumulatedGas))

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

func TestTxCache_selectTransactionsFromBunches(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		selected, accumulatedGas := selectTransactionsFromBunches(session, []bunchOfTransactions{}, 10_000_000_000, math.MaxInt, selectionLoopMaximumDuration)

		require.Equal(t, 0, len(selected))
		require.Equal(t, uint64(0), accumulatedGas)
	})
}

func TestBenchmarkTxCache_acquireBunchesOfTransactions(t *testing.T) {
	config := ConfigSourceMe{
		Name:                        "untitled",
		NumChunks:                   16,
		NumBytesThreshold:           1000000000,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
		CountThreshold:              300001,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
	}

	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	sw := core.NewStopWatch()

	t.Run("numSenders = 10000, numTransactions = 100", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 10000, 100)

		require.Equal(t, 1000000, int(cache.CountTx()))

		sw.Start(t.Name())
		bunches := cache.acquireBunchesOfTransactions()
		sw.Stop(t.Name())

		require.Len(t, bunches, 10000)
		require.Len(t, bunches[0], 100)
		require.Len(t, bunches[len(bunches)-1], 100)
	})

	t.Run("numSenders = 50000, numTransactions = 2", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 50000, 2)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		bunches := cache.acquireBunchesOfTransactions()
		sw.Stop(t.Name())

		require.Len(t, bunches, 50000)
		require.Len(t, bunches[0], 2)
		require.Len(t, bunches[len(bunches)-1], 2)
	})

	t.Run("numSenders = 100000, numTransactions = 1", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 100000, 1)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		bunches := cache.acquireBunchesOfTransactions()
		sw.Stop(t.Name())

		require.Len(t, bunches, 100000)
		require.Len(t, bunches[0], 1)
		require.Len(t, bunches[len(bunches)-1], 1)
	})

	t.Run("numSenders = 300000, numTransactions = 1", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 300000, 1)

		require.Equal(t, 300000, int(cache.CountTx()))

		sw.Start(t.Name())
		bunches := cache.acquireBunchesOfTransactions()
		sw.Stop(t.Name())

		require.Len(t, bunches, 300000)
		require.Len(t, bunches[0], 1)
		require.Len(t, bunches[len(bunches)-1], 1)
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
	// 0.014468s (TestBenchmarkTxCache_acquireBunchesOfTransactions/numSenders_=_10000,_numTransactions_=_100)
	// 0.019183s (TestBenchmarkTxCache_acquireBunchesOfTransactions/numSenders_=_50000,_numTransactions_=_2)
	// 0.013876s (TestBenchmarkTxCache_acquireBunchesOfTransactions/numSenders_=_100000,_numTransactions_=_1)
	// 0.056631s (TestBenchmarkTxCache_acquireBunchesOfTransactions/numSenders_=_300000,_numTransactions_=_1)
}

func TestBenchmarkTxCache_selectTransactionsFromBunches(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numSenders = 1000, numTransactions = 1000", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(session, bunches, 10_000_000_000, math.MaxInt, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 10000, numTransactions = 100", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(session, bunches, 10_000_000_000, math.MaxInt, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 100000, numTransactions = 3", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		bunches := createBunchesOfTransactionsWithUniformDistribution(100000, 3)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(session, bunches, 10_000_000_000, math.MaxInt, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 300000, numTransactions = 1", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		bunches := createBunchesOfTransactionsWithUniformDistribution(300000, 1)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(session, bunches, 10_000_000_000, math.MaxInt, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
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
	// 0.057519s (TestBenchmarkTxCache_selectTransactionsFromBunches/numSenders_=_1000,_numTransactions_=_1000)
	// 0.048023s (TestBenchmarkTxCache_selectTransactionsFromBunches/numSenders_=_10000,_numTransactions_=_100)
	// 0.289515s (TestBenchmarkTxCache_selectTransactionsFromBunches/numSenders_=_100000,_numTransactions_=_3)
	// 0.460242s (TestBenchmarkTxCache_selectTransactionsFromBunches/numSenders_=_300000,_numTransactions_=_1)
}

func TestTxCache_selectTransactionsFromBunches_loopBreaks_whenTakesTooLong(t *testing.T) {
	t.Run("numSenders = 300000, numTransactions = 1", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		bunches := createBunchesOfTransactionsWithUniformDistribution(300000, 1)
		selected, accumulatedGas := selectTransactionsFromBunches(session, bunches, 10_000_000_000, 50_000, 1*time.Millisecond)

		require.Less(t, len(selected), 50_000)
		require.Less(t, int(accumulatedGas), 10_000_000_000)
	})
}

func TestBenchmarkTxCache_doSelectTransactions(t *testing.T) {
	config := ConfigSourceMe{
		Name:                        "untitled",
		NumChunks:                   16,
		NumBytesThreshold:           1000000000,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBound,
		CountThreshold:              300001,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
	}

	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	session := txcachemocks.NewSelectionSessionMock()

	sw := core.NewStopWatch()

	t.Run("numSenders = 10000, numTransactions = 100, maxNum = 50_000", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 10000, 100)

		require.Equal(t, 1000000, int(cache.CountTx()))

		sw.Start(t.Name())
		selected, accumulatedGas := cache.SelectTransactions(accountStateProvider, 10_000_000_000, 50_000, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 50000, len(selected))
		require.Equal(t, uint64(2_500_000_000), accumulatedGas)
	})

	t.Run("numSenders = 50000, numTransactions = 2, maxNum = 50_000", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 50000, 2)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		selected, accumulatedGas := cache.SelectTransactions(session, 10_000_000_000, 50_000, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 50000, len(selected))
		require.Equal(t, uint64(2_500_000_000), accumulatedGas)
	})

	t.Run("numSenders = 100000, numTransactions = 1, maxNum = 50_000", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 100000, 1)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		selected, accumulatedGas := cache.SelectTransactions(session, 10_000_000_000, 50_000, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 50000, len(selected))
		require.Equal(t, uint64(2_500_000_000), accumulatedGas)
	})

	t.Run("numSenders = 300000, numTransactions = 1, maxNum = 50_000", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 300000, 1)

		require.Equal(t, 300000, int(cache.CountTx()))

		sw.Start(t.Name())
		selected, accumulatedGas := cache.SelectTransactions(session, 10_000_000_000, 50_000, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 50000, len(selected))
		require.Equal(t, uint64(2_500_000_000), accumulatedGas)
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
	// 0.126612s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_10000,_numTransactions_=_100,_maxNum_=_50_000)
	// 0.107361s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_50000,_numTransactions_=_2,_maxNum_=_50_000)
	// 0.168364s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_100000,_numTransactions_=_1,_maxNum_=_50_000)
	// 0.305363s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_300000,_numTransactions_=_1,_maxNum_=_50_000)
}
