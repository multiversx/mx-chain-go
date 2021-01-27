package benchmarks

import (
	"fmt"
	"os"
	"time"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenVM"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
)

// ArgArwenBenchmark is the Arwen type benchmark argument used in constructor
type ArgArwenBenchmark struct {
	Name         string
	ScFilename   string
	TestingValue uint64
	Function     string
	Arguments    [][]byte
	NumRuns      int
}

type arwenBenchmark struct {
	name         string
	scFilename   string
	testingValue uint64
	function     string
	arguments    [][]byte
	numRuns      int
}

// NewArwenBenchmark creates a new benchmark based on SC execution through Arwen VM
func NewArwenBenchmark(arg ArgArwenBenchmark) *arwenBenchmark {
	return &arwenBenchmark{
		name:         arg.Name,
		scFilename:   arg.ScFilename,
		testingValue: arg.TestingValue,
		function:     arg.Function,
		arguments:    arg.Arguments,
		numRuns:      arg.NumRuns,
	}
}

// Run returns the time needed for the benchmark to be run
func (ab *arwenBenchmark) Run() (time.Duration, error) {
	err := fileExists(ab.scFilename)
	if err != nil {
		return 0, err
	}

	result, err := arwenVM.RunTest(ab.scFilename, ab.testingValue, ab.function, ab.arguments, ab.numRuns, createTestGasMap())
	if err != nil {
		return 0, err
	}

	return result.ExecutionTimeSpan, err
}

func createTestGasMap() map[string]map[string]uint64 {
	gasSchedule := arwenConfig.MakeGasMapForTests()
	gasSchedule = defaults.FillGasMapInternal(gasSchedule, 1)

	return gasSchedule
}

func fileExists(filename string) error {
	_, err := os.Stat(filename)
	return err
}

// Name returns the benchmark's name
func (ab *arwenBenchmark) Name() string {
	return fmt.Sprintf("%s, function %s, numRuns %d", ab.name, ab.function, ab.numRuns)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ab *arwenBenchmark) IsInterfaceNil() bool {
	return ab == nil
}
