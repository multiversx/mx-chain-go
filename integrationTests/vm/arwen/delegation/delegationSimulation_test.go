package delegation

import (
	"testing"
	"time"

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
	gasMapFilename := "../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml"
	delegationScFilename := "../testdata/delegation/delegation_v0_5_2_full.wasm"
	benchmarks, err := RunDelegationStressTest(numRuns, numBatches, numTxPerBatch, numQueriesPerBatch, gasMapFilename, delegationScFilename)
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
