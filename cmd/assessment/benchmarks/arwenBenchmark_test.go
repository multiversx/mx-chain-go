package benchmarks

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestArwenBenchmark_ShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testName := "fibonacci"
	ab := NewArwenBenchmark(
		ArgArwenBenchmark{
			Name:         testName,
			ScFilename:   "../testdata/fibonacci.wasm",
			TestingValue: 32,
			Function:     "_main",
			Arguments:    nil,
			NumRuns:      10,
		},
	)

	assert.False(t, check.IfNil(ab))

	testDuration, err := ab.Run()
	assert.Nil(t, err)
	assert.True(t, testDuration > 0)
	assert.True(t, strings.Contains(ab.Name(), testName))
	assert.True(t, strings.Contains(ab.Name(), "function"))
	assert.True(t, strings.Contains(ab.Name(), "numRuns"))
}
