//go:build !race

package delegation

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/stretchr/testify/require"
)

func TestSimulateExecutionOfStakeTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runDelegationExecutionSimulate(t, 2, 10, 100, 0)
}

func TestSimulateExecutionOfStakeTransactionAndQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runDelegationExecutionSimulate(t, 2, 10, 100, 100)
}

func runDelegationExecutionSimulate(t *testing.T, numRuns uint32, numBatches uint32, numTxPerBatch uint32, numQueriesPerBatch uint32) {
	gasMapFilename := integrationTests.GasSchedulePath
	gasSchedule, err := common.LoadGasScheduleConfig(gasMapFilename)
	gasSchedule[common.MaxPerTransaction]["MaxBuiltInCallsPerTx"] = 100000
	gasSchedule[common.MaxPerTransaction]["MaxNumberOfTransfersPerTx"] = 100000
	gasSchedule[common.MaxPerTransaction]["MaxNumberOfTrieReadsPerTx"] = 100000
	require.Nil(t, err)

	delegationScFilename := "../testdata/delegation/delegation_v0_5_2_full.wasm"
	benchmarks, err := RunDelegationStressTest(delegationScFilename, numRuns, numBatches, numTxPerBatch, numQueriesPerBatch, gasSchedule)
	require.Nil(t, err)
	require.True(t, len(benchmarks) > 0)

	minTime := time.Hour
	maxTime := time.Duration(0)
	sumTime := time.Duration(0)
	for _, d := range benchmarks {
		sumTime += d
		if minTime > d {
			minTime = d
		}
		if maxTime < d {
			maxTime = d
		}
	}

	log.Info("execution complete",
		"total time", sumTime,
		"average time", sumTime/time.Duration(len(benchmarks)),
		"total stake execution", len(benchmarks),
		"min time", minTime,
		"max time", maxTime,
	)
}
