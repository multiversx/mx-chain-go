package benchmarks

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestResults_ToDisplayTable(t *testing.T) {
	t.Parallel()

	errFound := errors.New("error found")
	tr := &TestResults{
		TotalDuration: time.Second,
		Error:         errFound,
		Results: []SingleResult{
			{
				Duration: time.Second * 2,
				Name:     "test 1",
				Error:    nil,
			},
			{
				Duration: time.Second * 3,
				Name:     "test 2",
				Error:    errFound,
			},
			{
				Duration: time.Second * 4,
				Name:     "test 3",
				Error:    nil,
			},
		},
	}

	tbl := tr.ToDisplayTable()
	fmt.Println(tbl)

	stringsToContain := []string{totalMarker, errFound.Error(), "test 1", "test 2", "test 3", "1.000", "2.000", "3.000", "4.000"}
	for _, str := range stringsToContain {
		assert.True(t, strings.Contains(tbl, str), "string %s not contained", str)
	}
}

func TestTestResults_ToStrings(t *testing.T) {
	t.Parallel()

	errFound := errors.New("error found")
	tr := &TestResults{
		TotalDuration: time.Second,
		Error:         errFound,
		Results: []SingleResult{
			{
				Duration: time.Second * 2,
				Name:     "test 1",
				Error:    nil,
			},
			{
				Duration: time.Second * 3,
				Name:     "test 2",
				Error:    errFound,
			},
			{
				Duration: time.Second * 4,
				Name:     "test 3",
				Error:    nil,
			},
		},
	}

	data := tr.ToStrings()

	stringsToContain := []string{totalMarker, errFound.Error(), "test 1", "test 2", "test 3", "1.000", "2.000", "3.000", "4.000"}
	for _, str := range stringsToContain {
		found := false
		for _, line := range data {
			for _, cell := range line {
				if strings.Contains(cell, str) {
					found = true
					break
				}
			}
		}

		assert.True(t, found, "string %s not contained", str)
	}
}
