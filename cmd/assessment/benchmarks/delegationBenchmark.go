package benchmarks

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/delegation"
)

// ArgDelegationBenchmark is the delegation type benchmark argument used in constructor
type ArgDelegationBenchmark struct {
	Name               string
	ScFilename         string
	NumRuns            int
	NumBatches         uint32
	NumTxPerBatch      uint32
	NumQueriesPerBatch uint32
}

type delegationBenchmark struct {
	name               string
	scFilename         string
	numRuns            int
	numBatches         uint32
	numTxPerBatch      uint32
	numQueriesPerBatch uint32
}

// NewDelegationBenchmark creates a new benchmark based on delegation SC execution through Arwen VM
func NewDelegationBenchmark(arg ArgDelegationBenchmark) *delegationBenchmark {
	return &delegationBenchmark{
		name:               arg.Name,
		scFilename:         arg.ScFilename,
		numRuns:            arg.NumRuns,
		numBatches:         arg.NumBatches,
		numQueriesPerBatch: arg.NumQueriesPerBatch,
		numTxPerBatch:      arg.NumTxPerBatch,
	}
}

// Run returns the time needed for the benchmark to be run
func (db *delegationBenchmark) Run() (time.Duration, error) {
	results, err := delegation.RunDelegationStressTest(
		db.scFilename,
		uint32(db.numRuns),
		db.numBatches,
		db.numTxPerBatch,
		db.numQueriesPerBatch,
		createTestGasMap(),
	)

	return getMinimumTimeDuration(results), err
}

func getMinimumTimeDuration(durations []time.Duration) time.Duration {
	min := time.Hour
	for _, d := range durations {
		if min > d {
			min = d
		}
	}

	return min
}

// Name returns the benchmark's name
func (db *delegationBenchmark) Name() string {
	return fmt.Sprintf("%s, %d stake executions, minimum duration", db.name, uint32(db.numRuns)*db.numBatches*db.numTxPerBatch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (db *delegationBenchmark) IsInterfaceNil() bool {
	return db == nil
}
