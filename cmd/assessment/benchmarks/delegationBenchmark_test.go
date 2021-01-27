package benchmarks

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDelegationBenchmark_ShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testName := "delegation"
	db := NewDelegationBenchmark(
		ArgDelegationBenchmark{
			Name:               testName,
			ScFilename:         "../testdata/delegation_v0_5_2_full.wasm",
			NumRuns:            1,
			NumTxPerBatch:      100,
			NumBatches:         1,
			NumQueriesPerBatch: 0,
		},
	)

	assert.False(t, check.IfNil(db))

	testDuration, err := db.Run()
	assert.Nil(t, err)
	assert.True(t, testDuration > 0)
	assert.True(t, strings.Contains(db.Name(), testName))
	assert.True(t, strings.Contains(db.Name(), "stake executions"))
}
