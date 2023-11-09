package benchmarks

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("assessment/benchmarks")

// ThresholdEnoughComputingPower represents the threshold considered when deciding if a Node Under Test (NUT) is considered
// having (or not) enough compute capability when executing SC calls. if the result.TotalDuration exceeds this const, the
// NUT is considered less powerful (the NUT needed more time to process same amount of work). The value is hardcoded and
// was decided through testing.
const ThresholdEnoughComputingPower = 31 * time.Second

type coordinator struct {
	benchmarks []BenchmarkRunner
}

// NewCoordinator will create a coordinator used to launch all provided benchmarks
func NewCoordinator(benchmarks []BenchmarkRunner) (*coordinator, error) {
	if len(benchmarks) == 0 {
		return nil, ErrEmptyBenchmarksSlice
	}

	for index, b := range benchmarks {
		if check.IfNil(b) {
			return nil, fmt.Errorf("%w at index %d", ErrNilBenchmark, index)
		}
	}

	return &coordinator{
		benchmarks: benchmarks,
	}, nil
}

// RunAllTests will launch all contained tests. Errors if at least one benchmark errored
func (c *coordinator) RunAllTests() *TestResults {
	cumulative := time.Duration(0)
	var lastErr error

	testResult := TestResults{
		Results: make([]SingleResult, 0, len(c.benchmarks)),
	}
	for i, b := range c.benchmarks {
		log.Info(fmt.Sprintf("running benchmark %d out of %d", i+1, len(c.benchmarks)),
			"name", b.Name())
		elapsed, err := b.Run()
		if err != nil {
			log.Error("error running benchmark", "name", b.Name(), "error", err)
			lastErr = err
		}
		cumulative += elapsed

		testResult.Results = append(testResult.Results,
			SingleResult{
				Duration: elapsed,
				Name:     b.Name(),
				Error:    err,
			},
		)
	}

	testResult.Error = lastErr
	testResult.TotalDuration = cumulative
	testResult.EnoughComputingPower = cumulative < ThresholdEnoughComputingPower
	return &testResult
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *coordinator) IsInterfaceNil() bool {
	return c == nil
}
