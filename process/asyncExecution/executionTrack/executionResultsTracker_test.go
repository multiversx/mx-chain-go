package executionTrack

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
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

		header := &block.HeaderV3{}
		errC := tracker.CleanConfirmedExecutionResults(header)
		require.Nil(t, errC)
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

		executionResults := []*block.ExecutionResult{
			{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("hash2"),
					HeaderNonce: 11,
				},
			},
			{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("hash3"),
					HeaderNonce: 12,
				},
			},
		}

		header := &block.HeaderV3{
			ExecutionResults: executionResults,
		}

		err = tracker.AddExecutionResult(executionResults[0])
		require.Nil(t, err)
		err = tracker.AddExecutionResult(executionResults[1])
		require.Nil(t, err)

		errC := tracker.CleanConfirmedExecutionResults(header)
		require.Nil(t, errC)

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

		header := &block.HeaderV3{
			ExecutionResults: []*block.ExecutionResult{
				{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash:  []byte("hash22"),
						HeaderNonce: 22,
					},
				},
			},
		}

		errC := tracker.CleanConfirmedExecutionResults(header)
		require.True(t, errors.Is(errC, ErrCannotFindExecutionResult))
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

	header := &block.HeaderV3{
		ExecutionResults: []*block.ExecutionResult{
			{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("hash1"),
					HeaderNonce: 11,
				},
			},
			{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("different"),
					HeaderNonce: 12,
				},
			},
		},
	}

	err = tracker.CleanConfirmedExecutionResults(header)
	require.True(t, errors.Is(err, ErrExecutionResultMissmatch))

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

func TestExecutionResultsTracker_RemoveFromNonce(t *testing.T) {
	t.Parallel()

	t.Run("getPendingExecutionResults error should error", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash0"),
				HeaderNonce: 10,
			},
		})
		require.Nil(t, err)

		// Add execution result with nonce 12 (skipping 11) to create an inconsistent state
		// This will cause getPendingExecutionResults to return an error
		tracker.executionResultsByHash["hash2"] = &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash2"),
				HeaderNonce: 12,
			},
		}
		tracker.nonceHash.addNonceHash(12, "hash2")

		err = tracker.RemoveFromNonce(12)
		require.True(t, errors.Is(err, ErrDifferentNoncesConfirmedExecutionResults))
	})

	t.Run("remove single execution result should update lastExecutedResultHash to last notarized", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		lastNotarizedHash := []byte("hash0")
		err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  lastNotarizedHash,
				HeaderNonce: 10,
			},
		})
		require.NoError(t, err)

		executionResult1 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash1"),
				HeaderNonce: 11,
			},
		}
		err = tracker.AddExecutionResult(executionResult1)
		require.NoError(t, err)

		err = tracker.RemoveFromNonce(11)
		require.NoError(t, err)

		pending, err := tracker.GetPendingExecutionResults()
		require.NoError(t, err)
		require.Equal(t, 0, len(pending))
		lastNotarizedExecRes, err := tracker.GetLastNotarizedExecutionResult()
		require.NoError(t, err)
		require.Equal(t, lastNotarizedHash, lastNotarizedExecRes.GetHeaderHash())
	})

	t.Run("remove from middle hash should remove that hash and all with higher nonces", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash0"),
				HeaderNonce: 10,
			},
		})
		require.NoError(t, err)

		executionResult1 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash1"),
				HeaderNonce: 11,
			},
		}
		err = tracker.AddExecutionResult(executionResult1)
		require.NoError(t, err)

		executionResult2 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash2"),
				HeaderNonce: 12,
			},
		}
		err = tracker.AddExecutionResult(executionResult2)
		require.NoError(t, err)

		executionResult3 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash3"),
				HeaderNonce: 13,
			},
		}
		err = tracker.AddExecutionResult(executionResult3)
		require.NoError(t, err)

		executionResult4 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash4"),
				HeaderNonce: 14,
			},
		}
		err = tracker.AddExecutionResult(executionResult4)
		require.NoError(t, err)

		// Remove from hash2 (nonce 12), should keep only hash1 (nonce 11)
		err = tracker.RemoveFromNonce(12)
		require.NoError(t, err)

		pending, err := tracker.GetPendingExecutionResults()
		require.NoError(t, err)
		require.Equal(t, 1, len(pending))
		require.Equal(t, executionResult1, pending[0])
		require.Equal(t, []byte("hash1"), tracker.lastExecutedResultHash)

		// Verify removed results
		_, err = tracker.GetPendingExecutionResultByHash([]byte("hash2"))
		require.True(t, errors.Is(err, ErrCannotFindExecutionResult))
		_, err = tracker.GetPendingExecutionResultByHash([]byte("hash3"))
		require.True(t, errors.Is(err, ErrCannotFindExecutionResult))
		_, err = tracker.GetPendingExecutionResultByHash([]byte("hash4"))
		require.True(t, errors.Is(err, ErrCannotFindExecutionResult))

		// Verify kept result
		result, err := tracker.GetPendingExecutionResultByHash([]byte("hash1"))
		require.NoError(t, err)
		require.Equal(t, executionResult1, result)
	})

	t.Run("remove a missing nonce should remove the higher ones", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		err := tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash0"),
				HeaderNonce: 10,
			},
		})
		require.NoError(t, err)

		executionResult1 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash1"),
				HeaderNonce: 11,
			},
		}
		err = tracker.AddExecutionResult(executionResult1)
		require.NoError(t, err)

		executionResult2 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash2"),
				HeaderNonce: 12,
			},
		}
		err = tracker.AddExecutionResult(executionResult2)
		require.NoError(t, err)

		// Remove from nonce 8(missing) should remove all
		err = tracker.RemoveFromNonce(8)
		require.NoError(t, err)

		pending, err := tracker.GetPendingExecutionResults()
		require.NoError(t, err)
		require.Equal(t, 0, len(pending))

		// Verify removed results
		_, err = tracker.GetPendingExecutionResultByHash([]byte("hash1"))
		require.True(t, errors.Is(err, ErrCannotFindExecutionResult))
		_, err = tracker.GetPendingExecutionResultByHash([]byte("hash2"))
		require.True(t, errors.Is(err, ErrCannotFindExecutionResult))
	})
}

func TestExecutionResultsTracker_OnHeaderEvicted(t *testing.T) {
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

	err = tracker.AddExecutionResult(executionResults[0])
	require.Nil(t, err)

	err = tracker.AddExecutionResult(executionResults[1])
	require.Nil(t, err)

	// evicting already processed nonce should remove it from pending
	tracker.OnHeaderEvicted(executionResults[1].GetHeaderNonce())

	results, errG := tracker.GetPendingExecutionResults()
	require.Nil(t, errG)
	require.Equal(t, 1, len(results))

	// evicting already processed nonce should remove it from pending
	tracker.OnHeaderEvicted(executionResults[0].GetHeaderNonce())

	results, errG = tracker.GetPendingExecutionResults()
	require.Nil(t, errG)
	require.Equal(t, 0, len(results))
}
