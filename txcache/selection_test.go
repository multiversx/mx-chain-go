package txcache

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/multiversx/mx-chain-storage-go/types"
	"github.com/stretchr/testify/require"
)

func TestTxCache_SelectTransactions_Dummy(t *testing.T) {
	t.Run("all having same PPU", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()

		cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7))
		cache.AddTx(createTx([]byte("hash-bob-6"), "bob", 6))
		cache.AddTx(createTx([]byte("hash-bob-5"), "bob", 5))
		cache.AddTx(createTx([]byte("hash-carol-1"), "carol", 1))

		selected, accumulatedGas := cache.SelectTransactions(accountStateProvider, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
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
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withGasPrice(100))
		cache.AddTx(createTx([]byte("hash-bob-5"), "bob", 5).withGasPrice(50))
		cache.AddTx(createTx([]byte("hash-carol-3"), "carol", 3).withGasPrice(75))

		selected, accumulatedGas := cache.SelectTransactions(accountStateProvider, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
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
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()

		cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4).withGasLimit(100000))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3).withGasLimit(100000))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withGasLimit(500000))
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withGasLimit(200000))
		cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7).withGasLimit(400000))
		cache.AddTx(createTx([]byte("hash-bob-6"), "bob", 6).withGasLimit(50000))
		cache.AddTx(createTx([]byte("hash-bob-5"), "bob", 5).withGasLimit(50000))
		cache.AddTx(createTx([]byte("hash-carol-1"), "carol", 1).withGasLimit(50000))

		selected, accumulatedGas := cache.SelectTransactions(accountStateProvider, 760000, math.MaxInt, selectionLoopMaximumDuration)
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

func TestTxCache_SelectTransactions_HandlesGapsAndLowerNonces(t *testing.T) {
	t.Run("with middle gaps", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()

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

		sorted, accumulatedGas := cache.SelectTransactions(accountStateProvider, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		expectedNumSelected := 3 + 1 + 2 // 3 alice + 1 bob + 2 carol
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 300000, int(accumulatedGas))
	})

	t.Run("with initial gaps", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()

		noncesByAddress := accountStateProvider.AccountStateByAddress
		noncesByAddress["alice"] = &types.AccountState{
			Nonce:   1,
			Balance: big.NewInt(1000000000000000000),
		}
		noncesByAddress["bob"] = &types.AccountState{
			Nonce:   42,
			Balance: big.NewInt(1000000000000000000),
		}

		// No gap
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))

		// Initial gap
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 44))
		cache.AddTx(createTx([]byte("hash-bob-43"), "bob", 45))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 46))

		// Unknown
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8))

		sorted, accumulatedGas := cache.SelectTransactions(accountStateProvider, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		expectedNumSelected := 3 + 0 + 2 // 3 alice + 0 bob + 2 carol
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 250000, int(accumulatedGas))
	})

	t.Run("with lower nonces", func(t *testing.T) {
		cache := newUnconstrainedCacheToTest()
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()

		noncesByAddress := accountStateProvider.AccountStateByAddress
		noncesByAddress["alice"] = &types.AccountState{
			Nonce:   1,
			Balance: big.NewInt(1000000000000000000),
		}
		noncesByAddress["bob"] = &types.AccountState{
			Nonce:   42,
			Balance: big.NewInt(1000000000000000000),
		}

		// Good sequence
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3))

		// A few with lower nonce
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 40))
		cache.AddTx(createTx([]byte("hash-bob-43"), "bob", 41))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 42))

		// Unknown
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8))

		sorted, accumulatedGas := cache.SelectTransactions(accountStateProvider, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
		expectedNumSelected := 3 + 1 + 2 // 3 alice + 1 bob + 2 carol
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 300000, int(accumulatedGas))
	})
}

func TestTxCache_SelectTransactions_WhenTransactionsAddedInReversedNonceOrder(t *testing.T) {
	cache := newUnconstrainedCacheToTest()
	accountStateProvider := txcachemocks.NewAccountStateProviderMock()

	// Add "nSenders" * "nTransactionsPerSender" transactions in the cache (in reversed nonce order)
	nSenders := 1000
	nTransactionsPerSender := 100
	nTotalTransactions := nSenders * nTransactionsPerSender

	for senderTag := 0; senderTag < nSenders; senderTag++ {
		sender := fmt.Sprintf("sender:%d", senderTag)

		for txNonce := nTransactionsPerSender; txNonce > 0; txNonce-- {
			txHash := fmt.Sprintf("hash:%d:%d", senderTag, txNonce)
			tx := createTx([]byte(txHash), sender, uint64(txNonce))
			cache.AddTx(tx)
		}
	}

	require.Equal(t, uint64(nTotalTransactions), cache.CountTx())

	sorted, accumulatedGas := cache.SelectTransactions(accountStateProvider, math.MaxUint64, math.MaxInt, selectionLoopMaximumDuration)
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
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()
		selected, accumulatedGas := selectTransactionsFromBunches(accountStateProvider, []bunchOfTransactions{}, 10_000_000_000, math.MaxInt, selectionLoopMaximumDuration)

		require.Equal(t, 0, len(selected))
		require.Equal(t, uint64(0), accumulatedGas)
	})
}

func TestBenchmarkTxCache_selectTransactionsFromBunches(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numSenders = 1000, numTransactions = 1000", func(t *testing.T) {
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(accountStateProvider, bunches, 10_000_000_000, math.MaxInt, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 10000, numTransactions = 100", func(t *testing.T) {
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(accountStateProvider, bunches, 10_000_000_000, math.MaxInt, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 100000, numTransactions = 3", func(t *testing.T) {
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()
		bunches := createBunchesOfTransactionsWithUniformDistribution(100000, 3)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(accountStateProvider, bunches, 10_000_000_000, math.MaxInt, selectionLoopMaximumDuration)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 300000, numTransactions = 1", func(t *testing.T) {
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()
		bunches := createBunchesOfTransactionsWithUniformDistribution(300000, 1)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(accountStateProvider, bunches, 10_000_000_000, math.MaxInt, selectionLoopMaximumDuration)
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
	// 0.029651s (TestTxCache_selectTransactionsFromBunches/numSenders_=_1000,_numTransactions_=_1000)
	// 0.026440s (TestTxCache_selectTransactionsFromBunches/numSenders_=_10000,_numTransactions_=_100)
	// 0.122592s (TestTxCache_selectTransactionsFromBunches/numSenders_=_100000,_numTransactions_=_3)
	// 0.219072s (TestTxCache_selectTransactionsFromBunches/numSenders_=_300000,_numTransactions_=_1)
}

func TestTxCache_selectTransactionsFromBunches_lookBreaks_whenTakesTooLong(t *testing.T) {
	t.Run("numSenders = 300000, numTransactions = 1", func(t *testing.T) {
		accountStateProvider := txcachemocks.NewAccountStateProviderMock()
		bunches := createBunchesOfTransactionsWithUniformDistribution(300000, 1)
		selected, accumulatedGas := selectTransactionsFromBunches(accountStateProvider, bunches, 10_000_000_000, 50_000, 1*time.Millisecond)

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
	accountStateProvider := txcachemocks.NewAccountStateProviderMock()

	sw := core.NewStopWatch()

	t.Run("numSenders = 50000, numTransactions = 2, maxNum = 50_000", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 50000, 2)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		selected, accumulatedGas := cache.SelectTransactions(accountStateProvider, 10_000_000_000, 50_000, selectionLoopMaximumDuration)
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
		selected, accumulatedGas := cache.SelectTransactions(accountStateProvider, 10_000_000_000, 50_000, selectionLoopMaximumDuration)
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
		selected, accumulatedGas := cache.SelectTransactions(accountStateProvider, 10_000_000_000, 50_000, selectionLoopMaximumDuration)
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
	// 0.060508s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_50000,_numTransactions_=_2,_maxNum_=_50_000)
	// 0.103369s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_100000,_numTransactions_=_1,_maxNum_=_50_000)
	// 0.245621s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_300000,_numTransactions_=_1,_maxNum_=_50_000)
}
