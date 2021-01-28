package benchmarks

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestErc20InCBenchmark_ShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testName := "erc20C"
	ercb := NewErc20Benchmark(
		ArgErc20Benchmark{
			Name:               testName,
			ScFilename:         "../testdata/erc20_c.wasm",
			Function:           "transferToken",
			NumRuns:            1,
			NumTransfersPerRun: 1000,
		},
	)

	assert.False(t, check.IfNil(ercb))

	testDuration, err := ercb.Run()
	assert.Nil(t, err)
	assert.True(t, testDuration > 0)
	assert.True(t, strings.Contains(ercb.Name(), testName))
	assert.True(t, strings.Contains(ercb.Name(), "transfer executions"))
}

func TestErc20InRustBenchmark_ShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testName := "erc20Rust"
	ercb := NewErc20Benchmark(
		ArgErc20Benchmark{
			Name:               testName,
			ScFilename:         "../testdata/erc20_rust.wasm",
			Function:           "transfer",
			NumRuns:            1,
			NumTransfersPerRun: 1000,
		},
	)

	assert.False(t, check.IfNil(ercb))

	testDuration, err := ercb.Run()
	assert.Nil(t, err)
	assert.True(t, testDuration > 0)
	assert.True(t, strings.Contains(ercb.Name(), testName))
	assert.True(t, strings.Contains(ercb.Name(), "transfer executions"))
}
