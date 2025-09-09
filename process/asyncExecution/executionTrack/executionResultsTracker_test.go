package executionTrack

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewExecutionResultsTracker(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()
	require.False(t, tracker.IsInterfaceNil())
}

func TestSetAndGetLastNotarizedResult(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	_, err := tracker.GetLastNotarizedExecutionResult()
	require.Equal(t, ErrNilLastNotarizedExecutionResult, err)

	err = tracker.SetLastNotarizedResult(nil)
	require.Equal(t, ErrNilExecutionResult, err)

	execResult := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderNonce: 10,
		},
	}
	err = tracker.SetLastNotarizedResult(execResult)
	require.NoError(t, err)

	lastNotarizedResult, err := tracker.GetLastNotarizedExecutionResult()
	require.NoError(t, err)
	require.Equal(t, execResult, lastNotarizedResult)
}

func TestAddExecutionResult_AllBranches(t *testing.T) {
	t.Parallel()

	t.Run("nil execution result", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()

		err := tracker.AddExecutionResult(nil)
		require.ErrorIs(t, err, ErrNilExecutionResult)
	})

	t.Run("nil last notarized result", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()

		execResult := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash: []byte("hash1"), HeaderNonce: 1,
			},
		}
		err := tracker.AddExecutionResult(execResult)
		require.ErrorIs(t, err, ErrNilLastNotarizedExecutionResult)
	})

	t.Run("execution result nonce lower than last notarized", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		tracker.lastNotarizedResult = &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderNonce: 10,
			},
		}

		execResult := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash2"),
				HeaderNonce: 9,
			},
		}
		err := tracker.AddExecutionResult(execResult)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrWrongExecutionResultNonce))
		require.Contains(t, err.Error(), "is lower than last notarized nonce")
	})

	t.Run("execution result nonce not equal to the subsequent nonce after last executed", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		_ = tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderNonce: 10,
				HeaderHash:  []byte("hh"),
			},
		})

		execResult := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash3"),
				HeaderNonce: 12,
			},
		}
		err := tracker.AddExecutionResult(execResult)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrWrongExecutionResultNonce))
		require.Contains(t, err.Error(), "should be equal to the subsequent nonce after last executed")
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		_ = tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hh"),
				HeaderNonce: 10,
			},
		})

		execResult := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash3"),
				HeaderNonce: 11,
			},
		}
		err := tracker.AddExecutionResult(execResult)
		require.Nil(t, err)
	})

	t.Run("cannot find execution result", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		tracker.lastNotarizedResult = &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderNonce: 10,
			},
		}

		execResult := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash3"),
				HeaderNonce: 13,
			},
		}
		tracker.lastExecutedResultHash = []byte("h1")
		err := tracker.AddExecutionResult(execResult)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrCannotFindExecutionResult))
	})
}

func TestAddExecutionResultAndCleanShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("header with no execution results", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash1"),
				HeaderNonce: 10,
			},
		})
		require.Nil(t, err)

		header := &testscommon.HeaderHandlerWithExecutionResultsStub{
			GetExecutionResultsCalled: func() []data.ExecutionResultHandler {
				return nil
			},
		}
		res, errC := tracker.CleanConfirmedExecutionResults(header)
		require.Nil(t, errC)
		require.Equal(t, CleanResultOK, res.CleanResult)

		lastNotarizedResult, errL := tracker.GetLastNotarizedExecutionResult()
		require.Nil(t, errL)
		require.Equal(t, lastNotarizedResult.GetHeaderNonce(), res.LastMatchingResultNonce)
	})

	t.Run("header with 2 execution results", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash1"),
				HeaderNonce: 10,
			},
		})
		require.Nil(t, err)

		executionResults := []data.ExecutionResultHandler{
			&block.ExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("hash2"),
					HeaderNonce: 11,
				},
			},
			&block.ExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("hash3"),
					HeaderNonce: 12,
				},
			},
		}

		header := &testscommon.HeaderHandlerWithExecutionResultsStub{
			GetExecutionResultsCalled: func() []data.ExecutionResultHandler {
				return executionResults
			},
		}

		err = tracker.AddExecutionResult(executionResults[0])
		require.Nil(t, err)
		err = tracker.AddExecutionResult(executionResults[1])
		require.Nil(t, err)

		res, errC := tracker.CleanConfirmedExecutionResults(header)
		require.Nil(t, errC)
		require.Equal(t, CleanResultOK, res.CleanResult)
		require.Equal(t, uint64(12), res.LastMatchingResultNonce)

		results, errG := tracker.GetPendingExecutionResults()
		require.Nil(t, errG)
		require.Equal(t, 0, len(results))
	})

	t.Run("clean result not found", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash1"),
				HeaderNonce: 10,
			},
		})
		require.Nil(t, err)

		header := &testscommon.HeaderHandlerWithExecutionResultsStub{
			GetExecutionResultsCalled: func() []data.ExecutionResultHandler {
				return []data.ExecutionResultHandler{
					&block.ExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderHash:  []byte("hash22"),
							HeaderNonce: 22,
						},
					},
				}
			},
		}

		res, errC := tracker.CleanConfirmedExecutionResults(header)
		require.Nil(t, errC)
		require.Equal(t, CleanResultNotFound, res.CleanResult)
		require.Equal(t, uint64(10), res.LastMatchingResultNonce)
	})
}

func TestAddExecutionResultAndCleanDifferentResultsFromHeader(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
		},
	})
	require.Nil(t, err)

	executionResult1 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash1"),
			HeaderNonce: 11,
		},
	}
	err = tracker.AddExecutionResult(executionResult1)
	require.Nil(t, err)
	err = tracker.AddExecutionResult(&block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash2"),
			HeaderNonce: 12,
		},
	})
	require.Nil(t, err)
	err = tracker.AddExecutionResult(&block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash3"),
			HeaderNonce: 13,
		},
	})
	require.Nil(t, err)

	header := &testscommon.HeaderHandlerWithExecutionResultsStub{
		GetExecutionResultsCalled: func() []data.ExecutionResultHandler {
			return []data.ExecutionResultHandler{
				&block.ExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash:  []byte("hash1"),
						HeaderNonce: 11,
					},
				},
				&block.ExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash:  []byte("different"),
						HeaderNonce: 12,
					},
				},
			}
		},
	}

	res, err := tracker.CleanConfirmedExecutionResults(header)
	require.Nil(t, err)
	require.Equal(t, CleanResultMismatch, res.CleanResult)
	require.Equal(t, uint64(11), res.LastMatchingResultNonce)
	require.Equal(t, []byte("hash1"), tracker.lastExecutedResultHash)

	// check that everything before the mismatch was kept inside tracker
	results, err := tracker.GetPendingExecutionResults()
	require.Nil(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, executionResult1, results[0])
}

func TestExecutionResultsTracker_GetPendingExecutionResultByHashAndHash(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
		},
	})
	require.Nil(t, err)

	executionResult1 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash1"),
			HeaderNonce: 11,
		},
	}
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

func TestExecutionResultsTracker_RemoveByHash(t *testing.T) {

	t.Run("remove header should update the last executed hash to last notarized", func(t *testing.T) {
		tracker := NewExecutionResultsTracker()

		lastNotarizedHash := []byte("hash0")
		err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  lastNotarizedHash,
				HeaderNonce: 10,
			},
		})
		require.Nil(t, err)

		headerHash := []byte("hash1")
		executionResult1 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash1"),
				HeaderNonce: 11,
			},
		}
		err = tracker.AddExecutionResult(executionResult1)
		require.Nil(t, err)

		require.Equal(t, headerHash, tracker.lastExecutedResultHash)

		pending, err := tracker.GetPendingExecutionResults()
		require.Nil(t, err)
		require.Equal(t, 1, len(pending))
		execResult := pending[0].(*block.ExecutionResult)
		require.Equal(t, executionResult1, execResult)

		err = tracker.RemoveByHash(headerHash)
		require.Nil(t, err)

		pending, err = tracker.GetPendingExecutionResults()
		require.Nil(t, err)
		require.Equal(t, 0, len(pending))
		require.Equal(t, lastNotarizedHash, tracker.lastExecutedResultHash)
	})

	t.Run("remove header not found should skip execution result at add", func(t *testing.T) {
		tracker := NewExecutionResultsTracker()

		lastNotarizedHash := []byte("hash0")
		err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  lastNotarizedHash,
				HeaderNonce: 10,
			},
		})
		require.Nil(t, err)

		headerHash := []byte("hash1")
		err = tracker.RemoveByHash(headerHash)
		require.Nil(t, err)

		executionResult1 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  headerHash,
				HeaderNonce: 11,
			},
		}
		err = tracker.AddExecutionResult(executionResult1)
		require.Nil(t, err)

		pending, err := tracker.GetPendingExecutionResults()
		require.Nil(t, err)
		require.Equal(t, 0, len(pending))
		require.Equal(t, lastNotarizedHash, tracker.lastExecutedResultHash)
	})
}
