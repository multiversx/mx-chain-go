package benchmarks

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenVM"
)

// ArgErc20Benchmark is the erc20 type benchmark argument used in constructor
type ArgErc20Benchmark struct {
	Name               string
	ScFilename         string
	Function           string
	NumRuns            int
	NumTransfersPerRun int
}

type erc20Benchmark struct {
	name               string
	scFilename         string
	function           string
	numRuns            int
	numTransfersPerRun int
}

// NewErc20Benchmark creates a new benchmark based on erc20 SC execution through Arwen VM
func NewErc20Benchmark(arg ArgErc20Benchmark) *erc20Benchmark {
	return &erc20Benchmark{
		name:               arg.Name,
		scFilename:         arg.ScFilename,
		function:           arg.Function,
		numRuns:            arg.NumRuns,
		numTransfersPerRun: arg.NumTransfersPerRun,
	}
}

// Run returns the time needed for the benchmark to be run
func (eb *erc20Benchmark) Run() (time.Duration, error) {
	if !core.DoesFileExist(eb.scFilename) {
		return 0, fmt.Errorf("%w, file %s", ErrFileDoesNotExists, eb.scFilename)
	}

	result, err := arwenVM.DeployAndExecuteERC20WithBigInt(
		eb.numRuns,
		eb.numTransfersPerRun,
		createTestGasMap(),
		eb.scFilename,
		eb.function,
		false,
	)
	if err != nil {
		return 0, err
	}

	return getMinimumTimeDuration(result), err
}

// Name returns the benchmark's name
func (eb *erc20Benchmark) Name() string {
	return fmt.Sprintf("%s, %d transfer executions, minimum duration", eb.name, eb.numRuns*eb.numTransfersPerRun)
}

// IsInterfaceNil returns true if there is no value under the interface
func (eb *erc20Benchmark) IsInterfaceNil() bool {
	return eb == nil
}
