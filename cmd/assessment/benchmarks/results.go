package benchmarks

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/display"
)

const totalMarker = "TOTAL"

// SingleResult contains the output data after a benchmark run
type SingleResult struct {
	time.Duration
	Name  string
	Error error
}

// TestResults represents the output structure containing the test results data
type TestResults struct {
	TotalDuration time.Duration
	Error         error
	Results       []SingleResult
}

// ToDisplayTable will output the contained data as an ASCII table
func (tr *TestResults) ToDisplayTable() string {
	hdr := []string{"Benchmark", "Time in seconds", "Error"}
	lines := make([]*display.LineData, 0, len(tr.Results)+1)
	for i, res := range tr.Results {
		errString := ""
		if res.Error != nil {
			errString = res.Error.Error()
		}
		lines = append(lines, display.NewLineData(
			i == len(tr.Results)-1,
			[]string{
				res.Name,
				tr.secondsAsString(res.Seconds()),
				errString,
			},
		))
	}

	lines = append(lines, display.NewLineData(
		false,
		[]string{
			totalMarker,
			tr.secondsAsString(tr.TotalDuration.Seconds()),
			"",
		},
	))

	tbl, err := display.CreateTableString(hdr, lines)
	if err != nil {
		return fmt.Sprintf("[ERR:%s]", err)
	}

	return tbl
}

func (tr *TestResults) secondsAsString(seconds float64) string {
	return fmt.Sprintf("%0.3f", seconds)
}

// ToStrings will return the contained data as strings (to be easily written, e.g. in a file)
func (tr *TestResults) ToStrings() [][]string {
	result := make([][]string, 0)

	for _, sr := range tr.Results {
		result = append(result, []string{
			sr.Name,
			tr.secondsAsString(sr.Seconds()),
			tr.errToString(sr.Error),
		})
	}

	result = append(result, []string{
		totalMarker,
		tr.secondsAsString(tr.TotalDuration.Seconds()),
		"",
	})

	return result
}

func (tr *TestResults) errToString(err error) string {
	if err != nil {
		return err.Error()
	}

	return ""
}
