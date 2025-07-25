package mempool

import (
	"fmt"
	"io"
	"math"
	"os"
	"runtime/pprof"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/txcache"
	"github.com/stretchr/testify/require"
)

// disabledCloser defines a mock for io.Closer
type disabledCloser struct {
}

// Close -
func (d disabledCloser) Close() error {
	return nil
}

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

func setUpPprofFilesIfNecessary() (func(t *testing.T) io.Closer, func(t *testing.T, closable io.Closer)) {
	if !shouldDoProfiling() {
		return disabledCreatePprofFiles, disabledStopProfiling
	}

	return createPprofFiles, stopProfiling
}

func shouldDoProfiling() bool {
	envPprof := os.Getenv("PPROF")
	shouldCreatePprofFiles := envPprof == "1"

	return shouldCreatePprofFiles
}

func createPprofFiles(t *testing.T) io.Closer {
	pprofDir := "./pprof"
	err := os.MkdirAll(pprofDir, os.ModePerm)
	require.Nil(t, err)

	testName := strings.ReplaceAll(t.Name(), "/", "_")
	fileName := fmt.Sprintf("%s/%s.pprof", pprofDir, testName)

	f, err := os.Create(fileName)
	require.NoError(t, err)

	err = pprof.StartCPUProfile(f)
	require.NoError(t, err)

	return f
}

// disabledCreatePprofFiles defines a disabled method for createPprofFiles
func disabledCreatePprofFiles(t *testing.T) io.Closer {
	return &disabledCloser{}
}

func stopProfiling(t *testing.T, f io.Closer) {
	pprof.StopCPUProfile()
	err := f.Close()
	require.NoError(t, err)
}

// disabledStopProfiling defines a disabled method for stopProfiling
func disabledStopProfiling(t *testing.T, f io.Closer) {
	return
}

// benchmark for the creation of breadcrumbs (which are created with each proposed block)
func TestBenchmark_OnProposedWithManyTxsAndSenders(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	sw := core.NewStopWatch()
	startPprof, stopPprof := setUpPprofFilesIfNecessary()

	t.Run("30_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 30_000
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("30_000 txs with 1000 addresses", func(t *testing.T) {
		numTxs := 30_000
		numAddresses := 1000

		f := startPprof(t)
		defer stopPprof(t, f)

		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("30_000 txs with 100 addresses", func(t *testing.T) {
		numTxs := 30_000
		numAddresses := 100

		f := startPprof(t)
		defer stopPprof(t, f)

		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("30_000 txs with 10 addresses", func(t *testing.T) {
		numTxs := 30_000
		numAddresses := 10

		f := startPprof(t)
		defer stopPprof(t, f)

		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("60_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 60_000
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("60_000 txs with 1000 addresses", func(t *testing.T) {
		numTxs := 60_000
		numAddresses := 1000

		f := startPprof(t)
		defer stopPprof(t, f)

		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("60_000 txs with 100 addresses", func(t *testing.T) {
		numTxs := 60_000
		numAddresses := 100

		f := startPprof(t)
		defer stopPprof(t, f)

		testOnProposed(t, sw, numTxs, numAddresses)
	})

	t.Run("60_000 txs with 10 addresses", func(t *testing.T) {
		numTxs := 60_000
		numAddresses := 10

		f := startPprof(t)
		defer stopPprof(t, f)

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
	startPprof, stopPprof := setUpPprofFilesIfNecessary()

	t.Run("15_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 30_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("30_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 60_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("60_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 120_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("90_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 180_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("100_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 200_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("1_000_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 2_000_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

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
	startPprof, stopPprof := setUpPprofFilesIfNecessary()

	t.Run("15_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 30_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("30_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 60_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("60_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 120_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("90_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 180_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("100_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 200_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("1_000_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 2_000_000
		numTxsToBeSelected := numTxs / 2
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

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
	startPprof, stopPprof := setUpPprofFilesIfNecessary()

	numTxs := 20_000
	numTxsToBeSelected := numTxs / 2

	t.Run("10_000 txs out of 20_000 in pool with 10 addresses", func(t *testing.T) {
		numAddresses := 10

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 20_000 in pool with 100 addresses", func(t *testing.T) {
		numAddresses := 100

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 20_000 in pool with 1000 addresses", func(t *testing.T) {
		numAddresses := 1000

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 20_000 in pool with 10_000 addresses", func(t *testing.T) {
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 20_000 in pool with 100_000 addresses", func(t *testing.T) {
		numAddresses := 100_000

		f := startPprof(t)
		defer stopPprof(t, f)

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
	startPprof, stopPprof := setUpPprofFilesIfNecessary()

	t.Run("10_000 txs with 10 addresses", func(t *testing.T) {
		numAddresses := 10

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs with 100 addresses", func(t *testing.T) {
		numAddresses := 100

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs with 1000 addresses", func(t *testing.T) {
		numAddresses := 1000

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs with 10_000 addresses", func(t *testing.T) {
		numAddresses := 10_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs with 100_000 addresses", func(t *testing.T) {
		numAddresses := 100_000

		f := startPprof(t)
		defer stopPprof(t, f)

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
	startPprof, stopPprof := setUpPprofFilesIfNecessary()

	t.Run("10_000 txs out of 10_000 in pool with 100_000 addresses", func(t *testing.T) {
		numTxs := 10_000
		numAddresses := 100_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 100_000 in pool with 100_000 addresses", func(t *testing.T) {
		numTxs := 100_000
		numAddresses := 100_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testFirstSelection(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 1_000_000 in pool with 100_000 addresses", func(t *testing.T) {
		numTxs := 1_000_000
		numAddresses := 100_000

		f := startPprof(t)
		defer stopPprof(t, f)

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
	startPprof, stopPprof := setUpPprofFilesIfNecessary()

	numTxsToBeSelected := 10_000

	t.Run("10_000 txs out of 100_000 in pool with 100_000 addresses", func(t *testing.T) {
		numTxs := 100_000
		numAddresses := 100_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelectionWithManyTxsInPool(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	t.Run("10_000 txs out of 1_000_000 in pool with 100_000 addresses", func(t *testing.T) {
		numTxs := 1_000_000
		numAddresses := 100_000

		f := startPprof(t)
		defer stopPprof(t, f)

		testSecondSelectionWithManyTxsInPool(t, sw, numTxs, numTxsToBeSelected, numAddresses)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}
