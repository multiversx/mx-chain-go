package executionTrack

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewExecutionResultsTracker(t *testing.T) {
	t.Parallel()

	tracker, err := NewExecutionResultsTracker()
	require.NoError(t, err)
	require.False(t, tracker.IsInterfaceNil())
}

func TestSetAndGetLastNotarizedResult(t *testing.T) {
	t.Parallel()

	tracker, _ := NewExecutionResultsTracker()

	_, err := tracker.GetLastNotarizedExecutionResult()
	require.Equal(t, ErrNilLastNotarizedExecutionResult, err)

	err = tracker.SetLastNotarizedResult(nil)
	require.Equal(t, ErrNilExecutionResult, err)

	execResult := &block.ExecutionResult{
		Nonce: 10,
	}
	err = tracker.SetLastNotarizedResult(execResult)
	require.NoError(t, err)

	lastNotarizedResult, err := tracker.GetLastNotarizedExecutionResult()
	require.NoError(t, err)
	require.Equal(t, execResult, lastNotarizedResult)
}

func TestAddExecutionResult_AllBranches(t *testing.T) {
	t.Parallel()

	tracker, _ := NewExecutionResultsTracker()

	// 1. Nil execution result
	err := tracker.AddExecutionResult(nil)
	require.ErrorIs(t, err, ErrNilExecutionResult)

	// 2. Nil lastNotarizedResult
	execResult := &block.ExecutionResult{HeaderHash: []byte("hash1"), Nonce: 1}
	err = tracker.AddExecutionResult(execResult)
	require.ErrorIs(t, err, ErrNilLastNotarizedExecutionResult)

	// Set lastNotarizedResult for further tests
	tracker.lastNotarizedResult = &block.ExecutionResult{Nonce: 10}

	// 3. executionResult.Nonce < lastNotarizedResult.Nonce
	execResult = &block.ExecutionResult{HeaderHash: []byte("hash2"), Nonce: 9}
	err = tracker.AddExecutionResult(execResult)
	require.Error(t, err)
	require.Contains(t, err.Error(), "is lower than last notarized nonce")

	// 4. lastExecutedResults.Nonce != executionResult.Nonce-1
	execResult = &block.ExecutionResult{HeaderHash: []byte("hash3"), Nonce: 12}
	err = tracker.AddExecutionResult(execResult)
	require.Error(t, err)
	require.Contains(t, err.Error(), "should be equal to the subsequent nonce after last executed")

	// 6. Success path
	execResult = &block.ExecutionResult{HeaderHash: []byte("hash3"), Nonce: 11}
	err = tracker.AddExecutionResult(execResult)
	require.Nil(t, err)

	execResult = &block.ExecutionResult{HeaderHash: []byte("hash3"), Nonce: 13}
	tracker.lastExecutedResultHash = []byte("h1")
	err = tracker.AddExecutionResult(execResult)
	require.Error(t, err)
	require.Contains(t, err.Error(), "last executed result not found")
}

func TestAddExecutionResultAndCleanShouldWork(t *testing.T) {
	t.Parallel()

	tracker, _ := NewExecutionResultsTracker()

	err := tracker.SetLastNotarizedResult(&block.ExecutionResult{HeaderHash: []byte("hash1"), Nonce: 10})
	require.Nil(t, err)

	executionResults := []*block.ExecutionResult{
		{
			HeaderHash: []byte("hash2"), Nonce: 11,
		},
		{
			HeaderHash: []byte("hash3"), Nonce: 12,
		},
	}

	err = tracker.AddExecutionResult(executionResults[0])
	require.Nil(t, err)
	err = tracker.AddExecutionResult(executionResults[1])
	require.Nil(t, err)

	header := &testscommon.HeaderHandlerWithExecutionResultsStub{
		GetExecutionResultsCalled: func() []*block.ExecutionResult {
			return executionResults
		},
	}
	res, err := tracker.CleanConfirmedExecutionResults(header)
	require.Nil(t, err)
	require.Equal(t, CleanResultOK, res.CleanResult)
	require.Equal(t, uint64(12), res.LastMatchingResultNonce)

	results, err := tracker.GetPendingExecutionResults()
	require.Nil(t, err)
	require.Equal(t, 0, len(results))

	header.GetExecutionResultsCalled = func() []*block.ExecutionResult {
		return nil
	}
	res, err = tracker.CleanConfirmedExecutionResults(header)
	require.Nil(t, err)
	require.Equal(t, CleanResultOK, res.CleanResult)

	lastNotarizedResult, err := tracker.GetLastNotarizedExecutionResult()
	require.Nil(t, err)
	require.Equal(t, lastNotarizedResult.Nonce, res.LastMatchingResultNonce)

	// notarized results from header not found
	header.GetExecutionResultsCalled = func() []*block.ExecutionResult {
		return []*block.ExecutionResult{
			{
				HeaderHash: []byte("hash22"), Nonce: 22,
			},
		}
	}
	res, err = tracker.CleanConfirmedExecutionResults(header)
	require.Nil(t, err)
	require.Equal(t, CleanResultNotFound, res.CleanResult)
	require.Equal(t, uint64(12), res.LastMatchingResultNonce)
}

func TestAddExecutionResultAndCleanDifferentResultsFromHeader(t *testing.T) {
	t.Parallel()

	tracker, _ := NewExecutionResultsTracker()

	err := tracker.SetLastNotarizedResult(&block.ExecutionResult{HeaderHash: []byte("hash0"), Nonce: 10})
	require.Nil(t, err)

	executionResult1 := &block.ExecutionResult{HeaderHash: []byte("hash1"), Nonce: 11}
	err = tracker.AddExecutionResult(executionResult1)
	require.Nil(t, err)
	err = tracker.AddExecutionResult(&block.ExecutionResult{HeaderHash: []byte("hash2"), Nonce: 12})
	require.Nil(t, err)
	err = tracker.AddExecutionResult(&block.ExecutionResult{HeaderHash: []byte("hash3"), Nonce: 13})
	require.Nil(t, err)

	header := &testscommon.HeaderHandlerWithExecutionResultsStub{
		GetExecutionResultsCalled: func() []*block.ExecutionResult {
			return []*block.ExecutionResult{
				{HeaderHash: []byte("hash1"), Nonce: 11},
				{HeaderHash: []byte("different"), Nonce: 12},
			}
		},
	}

	res, err := tracker.CleanConfirmedExecutionResults(header)
	require.Nil(t, err)
	require.Equal(t, CleanResultMismatch, res.CleanResult)
	require.Equal(t, uint64(11), res.LastMatchingResultNonce)
	require.Equal(t, []byte("hash1"), tracker.lastExecutedResultHash)

	// check that everything before the missmatch was kept inside tracker
	results, err := tracker.GetPendingExecutionResults()
	require.Nil(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, executionResult1, results[0])
}

func TestExecutionResultsTracker_GetPendingExecutionResultByHashAndHash(t *testing.T) {
	tracker, _ := NewExecutionResultsTracker()

	err := tracker.SetLastNotarizedResult(&block.ExecutionResult{HeaderHash: []byte("hash0"), Nonce: 10})
	require.Nil(t, err)

	executionResult1 := &block.ExecutionResult{HeaderHash: []byte("hash1"), Nonce: 11}
	err = tracker.AddExecutionResult(executionResult1)
	require.Nil(t, err)

	res, err := tracker.GetPendingExecutionResultByHash([]byte("hh"))
	require.Nil(t, res)
	require.True(t, errors.Is(err, ErrCannotFindExecutionResult))

	res, err = tracker.GetPendingExecutionResultByHash([]byte("hash1"))
	require.Nil(t, err)
	require.Equal(t, executionResult1, res)

	res, err = tracker.GetPendingExecutionResultByNonce(10)
	require.Nil(t, res)
	require.True(t, errors.Is(err, ErrCannotFindExecutionResult))

	res, err = tracker.GetPendingExecutionResultByNonce(11)
	require.Nil(t, err)
	require.Equal(t, executionResult1, res)
}
