package txcache

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestTxCache_selectTransactionsFromBunches(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("empty cache", func(t *testing.T) {
		merged, accumulatedGas := selectTransactionsFromBunches([]BunchOfTransactions{}, 10_000_000_000)

		require.Equal(t, 0, len(merged))
		require.Equal(t, uint64(0), accumulatedGas)
	})

	t.Run("numSenders = 1000, numTransactions = 1000", func(t *testing.T) {
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		merged, accumulatedGas := selectTransactionsFromBunches(bunches, 10_000_000_000)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(merged))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 10000, numTransactions = 100", func(t *testing.T) {
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		merged, accumulatedGas := selectTransactionsFromBunches(bunches, 10_000_000_000)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(merged))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 100000, numTransactions = 3", func(t *testing.T) {
		bunches := createBunchesOfTransactionsWithUniformDistribution(100000, 3)

		sw.Start(t.Name())
		merged, accumulatedGas := selectTransactionsFromBunches(bunches, 10_000_000_000)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(merged))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 300000, numTransactions = 1", func(t *testing.T) {
		bunches := createBunchesOfTransactionsWithUniformDistribution(300000, 1)

		sw.Start(t.Name())
		merged, accumulatedGas := selectTransactionsFromBunches(bunches, 10_000_000_000)
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
	// 0.059954s (TestTxCache_selectTransactionsFromBunches/numSenders_=_1000,_numTransactions_=_1000)
	// 0.087949s (TestTxCache_selectTransactionsFromBunches/numSenders_=_10000,_numTransactions_=_100)
	// 0.204968s (TestTxCache_selectTransactionsFromBunches/numSenders_=_100000,_numTransactions_=_3)
	// 0.506842s (TestTxCache_selectTransactionsFromBunches/numSenders_=_300000,_numTransactions_=_1)
}
