package factory

import (
	"github.com/ElrondNetwork/elrond-go/cmd/assessment/benchmarks"
)

type runner struct {
	coordinator benchmarkCoordinator
}

// NewRunner is a wrapper over the coordinator implementation that will assemble all the defined benchmarks
func NewRunner(testDataDirectory string) (*runner, error) {
	r := &runner{}

	list := CreateBenchmarksList(testDataDirectory)

	var err error
	r.coordinator, err = benchmarks.NewCoordinator(list)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// RunAllTests will call the inner coordinator's RunAllTests function
func (r *runner) RunAllTests() *benchmarks.TestResults {
	return r.coordinator.RunAllTests()
}
