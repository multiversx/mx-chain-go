package mempool

import (
	"fmt"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/txcache"
)

var (
	transferredValue = 1
	configSourceMe   = txcache.ConfigSourceMe{
		Name:                        "test",
		NumChunks:                   16,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBoundTest,
		CountThreshold:              math.MaxUint32,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
		TxCacheBoundsConfig: config.TxCacheBoundsConfig{
			MaxNumBytesPerSenderUpperBound: maxNumBytesPerSenderUpperBoundTest,
		},
	}
)

// benchmark for the creation of breadcrumbs (which are created with each proposed block)
func TestBenchmark_OnProposedWithManyTxsAndSenders(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	sw := core.NewStopWatch()

	t.Run("30_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 30_000
		numAddresses := 10_000
		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("30_000 txs with 1000 addresses", func(t *testing.T) {
		numTxs := 30_000
		numAddresses := 1000
		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("30_000 txs with 100 addresses", func(t *testing.T) {
		numTxs := 30_000
		numAddresses := 100
		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("30_000 txs with 10 addresses", func(t *testing.T) {
		numTxs := 30_000
		numAddresses := 10
		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("60_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 60_000
		numAddresses := 10_000
		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("60_000 txs with 1000 addresses", func(t *testing.T) {
		numTxs := 60_000
		numAddresses := 1000
		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("60_000 txs with 100 addresses", func(t *testing.T) {
		numTxs := 60_000
		numAddresses := 100
		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("60_000 txs with 10 addresses", func(t *testing.T) {
		numTxs := 60_000
		numAddresses := 10
		testOnProposed(t, sw, numTxs, numAddresses)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}

// benchmark for the selection of txs
func TestBenchmark_FirstSelectionWithManyTxsAndSenders(t *testing.T) {
	sw := core.NewStopWatch()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("15_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 30_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("30_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 60_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("60_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 120_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("90_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 180_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("100_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 200_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("1_000_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 2_000_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}

// benchmark for the selection of txs
func TestBenchmark_SecondSelectionWithManyTxsAndSenders(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	sw := core.NewStopWatch()

	t.Run("15_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 30_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("30_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 60_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("60_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 120_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("90_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 180_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("100_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 200_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("1_000_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 2_000_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}

func TestBenchmark_FirstSelectionOf10kTransactionsAndVariableNumberOfAddresses(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	sw := core.NewStopWatch()

	numTxs := 20_000
	numTxsToBeSelected := numTxs / 2

	t.Run("10_000 txs out of 20_000 in pool with 10 addresses", func(t *testing.T) {
		numAddresses := 10
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 20_000 in pool with 100 addresses", func(t *testing.T) {
		numAddresses := 100
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 20_000 in pool with 1000 addresses", func(t *testing.T) {
		numAddresses := 1000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 20_000 in pool with 10_000 addresses", func(t *testing.T) {
		numAddresses := 10_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 20_000 in pool with 100_000 addresses", func(t *testing.T) {
		numAddresses := 100_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}

func TestBenchmark_SecondSelection10kTransactionsAndVariableNumberOfAddresses(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	sw := core.NewStopWatch()
	numTxs := 20_000
	numTxsToBeSelected := numTxs / 2

	t.Run("10_000 txs with 10 addresses", func(t *testing.T) {
		numAddresses := 10
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs with 100 addresses", func(t *testing.T) {
		numAddresses := 100
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs with 1000 addresses", func(t *testing.T) {
		numAddresses := 1000
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs with 10_000 addresses", func(t *testing.T) {
		numAddresses := 10_000
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs with 100_000 addresses", func(t *testing.T) {
		numAddresses := 100_000
		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}

// worst case scenario: 100k addresses
func TestBenchmark_FirstSelection10KTransactionAndVariableNumOfTxsInPool(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	sw := core.NewStopWatch()
	numTxsToBeSelected := 10_000

	t.Run("10_000 txs out of 10_000 in pool with 100_000 addresses", func(t *testing.T) {
		numTxs := 10_000
		numAddresses := 100_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 100_000 in pool with 100_000 addresses", func(t *testing.T) {
		numTxs := 100_000
		numAddresses := 100_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 1_000_000 in pool with 100_000 addresses", func(t *testing.T) {
		numTxs := 1_000_000
		numAddresses := 100_000
		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}

func TestBenchmark_SecondSelection10KTransactionAndVariableNumOfTxsInPool(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	sw := core.NewStopWatch()
	numTxsToBeSelected := 10_000

	t.Run("10_000 txs out of 100_000 in pool with 100_000 addresses", func(t *testing.T) {
		numTxs := 100_000
		numAddresses := 100_000
		testSecondSelectionWithManyTxsInPool(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 1_000_000 in pool with 100_000 addresses", func(t *testing.T) {
		numTxs := 1_000_000
		numAddresses := 100_000
		testSecondSelectionWithManyTxsInPool(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}
