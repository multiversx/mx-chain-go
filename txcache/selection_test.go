package txcache

import (
	"fmt"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestTxCache_selectTransactionsFromBunches(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		merged, accumulatedGas := selectTransactionsFromBunches([]BunchOfTransactions{}, 10_000_000_000, math.MaxInt)

		require.Equal(t, 0, len(merged))
		require.Equal(t, uint64(0), accumulatedGas)
	})
}

func TestBenchmarkTxCache_selectTransactionsFromBunches(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numSenders = 1000, numTransactions = 1000", func(t *testing.T) {
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		merged, accumulatedGas := selectTransactionsFromBunches(bunches, 10_000_000_000, math.MaxInt)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(merged))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 10000, numTransactions = 100", func(t *testing.T) {
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		merged, accumulatedGas := selectTransactionsFromBunches(bunches, 10_000_000_000, math.MaxInt)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(merged))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 100000, numTransactions = 3", func(t *testing.T) {
		bunches := createBunchesOfTransactionsWithUniformDistribution(100000, 3)

		sw.Start(t.Name())
		merged, accumulatedGas := selectTransactionsFromBunches(bunches, 10_000_000_000, math.MaxInt)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(merged))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 300000, numTransactions = 1", func(t *testing.T) {
		bunches := createBunchesOfTransactionsWithUniformDistribution(300000, 1)

		sw.Start(t.Name())
		merged, accumulatedGas := selectTransactionsFromBunches(bunches, 10_000_000_000, math.MaxInt)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(merged))
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

func TestBenchmarktTxCache_doSelectTransactions(t *testing.T) {
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

	t.Run("numSenders = 50000, numTransactions = 2, maxNum = 50_000", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 50000, 2)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		merged, accumulatedGas := cache.SelectTransactions(10_000_000_000, 50_000)
		sw.Stop(t.Name())

		require.Equal(t, 50000, len(merged))
		require.Equal(t, uint64(2_500_000_000), accumulatedGas)
	})

	t.Run("numSenders = 100000, numTransactions = 1, maxNum = 50_000", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 100000, 1)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		merged, accumulatedGas := cache.SelectTransactions(10_000_000_000, 50_000)
		sw.Stop(t.Name())

		require.Equal(t, 50000, len(merged))
		require.Equal(t, uint64(2_500_000_000), accumulatedGas)
	})

	t.Run("numSenders = 300000, numTransactions = 1, maxNum = 50_000", func(t *testing.T) {
		cache, err := NewTxCache(config, txGasHandler)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 300000, 1)

		require.Equal(t, 300000, int(cache.CountTx()))

		sw.Start(t.Name())
		merged, accumulatedGas := cache.SelectTransactions(10_000_000_000, 50_000)
		sw.Stop(t.Name())

		require.Equal(t, 50000, len(merged))
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
	// 0.060508s (TestBenchmarktTxCache_doSelectTransactions/numSenders_=_50000,_numTransactions_=_2,_maxNum_=_50_000)
	// 0.103369s (TestBenchmarktTxCache_doSelectTransactions/numSenders_=_100000,_numTransactions_=_1,_maxNum_=_50_000)
	// 0.245621s (TestBenchmarktTxCache_doSelectTransactions/numSenders_=_300000,_numTransactions_=_1,_maxNum_=_50_000)
}
