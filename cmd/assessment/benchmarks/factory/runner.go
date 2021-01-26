package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/cmd/assessment/benchmarks"
	"github.com/ElrondNetwork/elrond-go/display"
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

// GetStringTableAfterRun will generate the string table after benchmark runs
func (r *runner) GetStringTableAfterRun() string {
	result := r.coordinator.RunAllTests()

	hdr := []string{"Benchmark", "Time in seconds", "Error"}
	lines := make([]*display.LineData, 0, len(result.Results)+1)
	for i, res := range result.Results {
		errString := ""
		if res.Error != nil {
			errString = res.Error.Error()
		}
		lines = append(lines, display.NewLineData(
			i == len(result.Results)-1,
			[]string{
				res.Name,
				fmt.Sprintf("%0.3f", res.Seconds()),
				errString,
			},
		))
	}

	lines = append(lines, display.NewLineData(
		false,
		[]string{
			"TOTAL",
			fmt.Sprintf("%0.3f", result.TotalDuration.Seconds()),
			"",
		},
	))

	tbl, err := display.CreateTableString(hdr, lines)
	if err != nil {
		return fmt.Sprintf("[ERR:%s]", err)
	}

	return tbl
}
