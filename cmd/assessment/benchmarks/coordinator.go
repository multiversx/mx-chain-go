package benchmarks

import (
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

var log = logger.GetOrCreate("assessment/benchmarks")

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
	return &testResult
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *coordinator) IsInterfaceNil() bool {
	return c == nil
}
