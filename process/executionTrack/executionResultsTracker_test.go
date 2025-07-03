package executionTrack

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestAddExecutionResult(t *testing.T) {
	t.Parallel()

	tracker, _ := NewExecutionResultsTracker()
	tracker.lastNotarizedResult = &block.ExecutionResult{Nonce: 10}

	execResult := &block.ExecutionResult{HeaderHash: []byte("hash1"), Nonce: 11}
	err := tracker.AddExecutionResult(execResult)
	require.NoError(t, err)

	stored, found := tracker.executionResultsByHash[string(execResult.HeaderHash)]
	require.True(t, found)
	require.Equal(t, execResult, stored)

	hashes := tracker.nonceHashes.getNonceHashes(11)
	require.Contains(t, hashes, string(execResult.HeaderHash))

	badResult := &block.ExecutionResult{HeaderHash: []byte("hash2"), Nonce: 10}
	err = tracker.AddExecutionResult(badResult)
	require.Error(t, err)

	badResult2 := &block.ExecutionResult{HeaderHash: []byte("hash3"), Nonce: 9}
	err = tracker.AddExecutionResult(badResult2)
	require.Error(t, err)
}

func TestGetPendingExecutionResults(t *testing.T) {
	t.Parallel()

	tracker, _ := NewExecutionResultsTracker()
	tracker.lastNotarizedResult = &block.ExecutionResult{Nonce: 9}

	res1 := &block.ExecutionResult{HeaderHash: []byte("h1"), Nonce: 10}
	res2 := &block.ExecutionResult{HeaderHash: []byte("h2"), Nonce: 11}
	_ = tracker.AddExecutionResult(res1)
	_ = tracker.AddExecutionResult(res2)

	results, err := tracker.GetPendingExecutionResults()
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Same(t, res1, results[0])
	require.Same(t, res2, results[1])

	tracker, _ = NewExecutionResultsTracker()
	tracker.lastNotarizedResult = &block.ExecutionResult{Nonce: 9}
	_ = tracker.AddExecutionResult(&block.ExecutionResult{HeaderHash: []byte("h1"), Nonce: 10})
	_ = tracker.AddExecutionResult(&block.ExecutionResult{HeaderHash: []byte("h3"), Nonce: 12}) // skip 11

	results, err = tracker.GetPendingExecutionResults()
	require.Error(t, err)
	require.Nil(t, results)
}

func TestGetPendingExecutionResultsByHash(t *testing.T) {
	t.Parallel()

	tracker, _ := NewExecutionResultsTracker()
	tracker.lastNotarizedResult = &block.ExecutionResult{Nonce: 9}

	res1 := &block.ExecutionResult{HeaderHash: []byte("h1"), Nonce: 10}
	_ = tracker.AddExecutionResult(res1)

	res, err := tracker.GetExecutionResultByHash([]byte("h1"))
	require.NoError(t, err)
	require.Equal(t, res1, res)
}

func TestGetPendingExecutionResultsByNonce(t *testing.T) {
	t.Parallel()

	tracker, _ := NewExecutionResultsTracker()
	tracker.lastNotarizedResult = &block.ExecutionResult{Nonce: 9}

	res1 := &block.ExecutionResult{HeaderHash: []byte("h1"), Nonce: 10}
	_ = tracker.AddExecutionResult(res1)
	res2 := &block.ExecutionResult{HeaderHash: []byte("h2"), Nonce: 10}
	_ = tracker.AddExecutionResult(res2)

	res, err := tracker.GetExecutionResultByNonce(10)
	require.NoError(t, err)
	require.Equal(t, res1, res[0])
	require.Equal(t, res2, res[1])
}
