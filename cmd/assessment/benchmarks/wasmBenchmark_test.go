package benchmarks

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestWasmBenchmark_ShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testName := "fibonacci"
	wb := NewWasmBenchmark(
		ArgWasmBenchmark{
			Name:         testName,
			ScFilename:   "../testdata/fibonacci.wasm",
			TestingValue: 32,
			Function:     "_main",
			Arguments:    nil,
			NumRuns:      10,
		},
	)

	assert.False(t, check.IfNil(wb))

	testDuration, err := wb.Run()
	assert.Nil(t, err)
	assert.True(t, testDuration > 0)
	assert.True(t, strings.Contains(wb.Name(), testName))
	assert.True(t, strings.Contains(wb.Name(), "function"))
	assert.True(t, strings.Contains(wb.Name(), "numRuns"))
}
