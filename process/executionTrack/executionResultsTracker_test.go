package executionTrack

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
)

func TestAddConfirmAndPopExecutionResult(t *testing.T) {
	tracker, err := NewExecutionResultsTracker()
	assert.NoError(t, err)

	headerHash := []byte("header1")
	execResult := &block.ExecutionResult{HeaderHash: headerHash}

	added := tracker.AddExecutionResult(execResult)
	assert.True(t, added)

	err = tracker.ConfirmExecutionResult(headerHash)
	assert.NoError(t, err)

	confirmed, err := tracker.PopConfirmedExecutionResults()
	assert.NoError(t, err)
	assert.Len(t, confirmed, 1)
	assert.Equal(t, execResult, confirmed[0])

	confirmed, err = tracker.PopConfirmedExecutionResults()
	assert.NoError(t, err)
	assert.Len(t, confirmed, 0)
}

func TestPopConfirmedExecutionResults_OnlyConfirmedArePopped(t *testing.T) {
	tracker, err := NewExecutionResultsTracker()
	assert.NoError(t, err)

	headerHash1 := []byte("header1")
	headerHash2 := []byte("header2")
	execResult1 := &block.ExecutionResult{HeaderHash: headerHash1}
	execResult2 := &block.ExecutionResult{HeaderHash: headerHash2}

	tracker.AddExecutionResult(execResult1)
	tracker.AddExecutionResult(execResult2)

	err = tracker.ConfirmExecutionResult(headerHash1)
	assert.NoError(t, err)

	confirmed, err := tracker.PopConfirmedExecutionResults()
	assert.NoError(t, err)
	assert.Len(t, confirmed, 1)
	assert.Equal(t, execResult1, confirmed[0])

	err = tracker.ConfirmExecutionResult(headerHash2)
	assert.NoError(t, err)
	confirmed, err = tracker.PopConfirmedExecutionResults()
	assert.NoError(t, err)
	assert.Len(t, confirmed, 1)
	assert.Equal(t, execResult2, confirmed[0])
}
