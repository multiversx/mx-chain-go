package executionTrack

import (
	"errors"
	"fmt"
	"testing"

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

		added, err := tracker.AddExecutionResult(nil)
		require.False(t, added)
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
		added, err := tracker.AddExecutionResult(execResult)
		require.False(t, added)
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
		added, err := tracker.AddExecutionResult(execResult)
		require.False(t, added)
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
		added, err := tracker.AddExecutionResult(execResult)
		require.False(t, added)
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
		added, err := tracker.AddExecutionResult(execResult)
		require.True(t, added)
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
		added, err := tracker.AddExecutionResult(execResult)
		require.False(t, added)
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

		_, err = tracker.AddExecutionResult(executionResults[0])
		require.Nil(t, err)
		_, err = tracker.AddExecutionResult(executionResults[1])
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
	_, err = tracker.AddExecutionResult(executionResult1)
	require.Nil(t, err)
	_, err = tracker.AddExecutionResult(&block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash2"),
			HeaderNonce: 12,
		},
	})
	require.Nil(t, err)
	_, err = tracker.AddExecutionResult(&block.ExecutionResult{
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
	require.True(t, errors.Is(err, ErrExecutionResultMismatch))

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
	_, err = tracker.AddExecutionResult(executionResult1)
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
		_, err = tracker.AddExecutionResult(executionResult1)
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
		_, err = tracker.AddExecutionResult(executionResult1)
		require.NoError(t, err)

		executionResult2 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash2"),
				HeaderNonce: 12,
			},
		}
		_, err = tracker.AddExecutionResult(executionResult2)
		require.NoError(t, err)

		executionResult3 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash3"),
				HeaderNonce: 13,
			},
		}
		_, err = tracker.AddExecutionResult(executionResult3)
		require.NoError(t, err)

		executionResult4 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash4"),
				HeaderNonce: 14,
			},
		}
		_, err = tracker.AddExecutionResult(executionResult4)
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
		_, err = tracker.AddExecutionResult(executionResult1)
		require.NoError(t, err)

		executionResult2 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash2"),
				HeaderNonce: 12,
			},
		}
		_, err = tracker.AddExecutionResult(executionResult2)
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

func TestExecutionResultsTracker_Clean(t *testing.T) {
	t.Parallel()

	t.Run("nil last notarized result should early exit", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		tracker.Clean(nil)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tracker := NewExecutionResultsTracker()
		_ = tracker.SetLastNotarizedResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash0"),
				HeaderNonce: 10,
			},
		})
		_, _ = tracker.AddExecutionResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash1"),
				HeaderNonce: 11,
			},
		})
		_, _ = tracker.AddExecutionResult(&block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("hash2"),
				HeaderNonce: 12,
			},
		})

		newLast := &block.BaseExecutionResult{
			HeaderHash:  []byte("hash_new"),
			HeaderNonce: 2,
		}
		tracker.Clean(newLast)

		pending, err := tracker.GetPendingExecutionResults()
		require.NoError(t, err)
		require.Equal(t, 0, len(pending))

		lastExec, err := tracker.GetLastNotarizedExecutionResult()
		require.NoError(t, err)
		require.Equal(t, newLast, lastExec)
	})
}

func TestExecutionResultsTracker_PopDismissedResults_EmptyByDefault(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()
	batches := tracker.PopDismissedResults()
	require.Nil(t, batches)
}

func TestExecutionResultsTracker_PopDismissedResults_OnCleanOnConsensusReached(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	lastNotarized := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
			RootHash:    []byte("rootHash0"),
		},
	}
	_ = tracker.SetLastNotarizedResult(lastNotarized)

	exec1 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash1"),
			HeaderNonce: 11,
			RootHash:    []byte("rootHash1"),
		},
	}
	exec2 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash2"),
			HeaderNonce: 12,
			RootHash:    []byte("rootHash2"),
		},
	}
	_, _ = tracker.AddExecutionResult(exec1)
	_, _ = tracker.AddExecutionResult(exec2)

	// Consensus commits nonce 11 with a different hash - dismisses nonce 11 and higher
	header := &block.HeaderV3{
		Nonce: 11,
	}
	tracker.CleanOnConsensusReached([]byte("different_hash"), header)

	batches := tracker.PopDismissedResults()
	require.Len(t, batches, 1)
	require.Len(t, batches[0].Results, 2)
	// Anchor should be lastNotarized (no pending results before nonce 11)
	require.Equal(t, lastNotarized.GetRootHash(), batches[0].AnchorResult.GetRootHash())
	require.Equal(t, exec1.GetRootHash(), batches[0].Results[0].GetRootHash())
	require.Equal(t, exec2.GetRootHash(), batches[0].Results[1].GetRootHash())

	// Second pop should return nil (queue drained)
	require.Nil(t, tracker.PopDismissedResults())
}

func TestExecutionResultsTracker_PopDismissedResults_AnchorIsLastPendingBeforeDismissal(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	lastNotarized := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
			RootHash:    []byte("rootHash0"),
		},
	}
	_ = tracker.SetLastNotarizedResult(lastNotarized)

	exec1 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash1"),
			HeaderNonce: 11,
			RootHash:    []byte("rootHash1"),
		},
	}
	exec2 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash2"),
			HeaderNonce: 12,
			RootHash:    []byte("rootHash2"),
		},
	}
	exec3 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash3"),
			HeaderNonce: 13,
			RootHash:    []byte("rootHash3"),
		},
	}
	_, _ = tracker.AddExecutionResult(exec1)
	_, _ = tracker.AddExecutionResult(exec2)
	_, _ = tracker.AddExecutionResult(exec3)

	// RemoveFromNonce(12) should dismiss exec2 and exec3, keeping exec1
	_ = tracker.RemoveFromNonce(12)

	batches := tracker.PopDismissedResults()
	require.Len(t, batches, 1)
	require.Len(t, batches[0].Results, 2)
	// Anchor should be exec1 (last pending result before nonce 12)
	require.Equal(t, exec1.GetRootHash(), batches[0].AnchorResult.GetRootHash())
	require.Equal(t, exec2.GetRootHash(), batches[0].Results[0].GetRootHash())
	require.Equal(t, exec3.GetRootHash(), batches[0].Results[1].GetRootHash())
}

func TestExecutionResultsTracker_PopDismissedResults_OnCleanConfirmedMismatch(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	lastNotarized := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
			RootHash:    []byte("rootHash0"),
		},
	}
	_ = tracker.SetLastNotarizedResult(lastNotarized)

	exec1 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash1"),
			HeaderNonce: 11,
			RootHash:    []byte("rootHash1"),
		},
	}
	exec2 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash2"),
			HeaderNonce: 12,
			RootHash:    []byte("rootHash2"),
		},
	}
	exec3 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash3"),
			HeaderNonce: 13,
			RootHash:    []byte("rootHash3"),
		},
	}
	_, _ = tracker.AddExecutionResult(exec1)
	_, _ = tracker.AddExecutionResult(exec2)
	_, _ = tracker.AddExecutionResult(exec3)

	// Header confirms exec1 matches, but exec2 mismatches at idx=1
	header := &block.HeaderV3{
		ExecutionResults: []*block.ExecutionResult{
			exec1,
			{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("different"),
					HeaderNonce: 12,
					RootHash:    []byte("differentRoot"),
				},
			},
		},
	}
	err := tracker.CleanConfirmedExecutionResults(header)
	require.ErrorIs(t, err, ErrExecutionResultMismatch)

	batches := tracker.PopDismissedResults()
	require.Len(t, batches, 1)
	require.Len(t, batches[0].Results, 2) // exec2 and exec3 dismissed
	// Anchor should be exec1 (last matching result, at idx=0 the preceding pending result)
	require.Equal(t, exec1.GetRootHash(), batches[0].AnchorResult.GetRootHash())
	require.Equal(t, exec2.GetRootHash(), batches[0].Results[0].GetRootHash())
	require.Equal(t, exec3.GetRootHash(), batches[0].Results[1].GetRootHash())
}

func TestExecutionResultsTracker_PopDismissedResults_OnCleanConfirmedMismatchAtFirstIndex(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	lastNotarized := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
			RootHash:    []byte("rootHash0"),
		},
	}
	_ = tracker.SetLastNotarizedResult(lastNotarized)

	exec1 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash1"),
			HeaderNonce: 11,
			RootHash:    []byte("rootHash1"),
		},
	}
	_, _ = tracker.AddExecutionResult(exec1)

	// Header has mismatch at idx=0
	header := &block.HeaderV3{
		ExecutionResults: []*block.ExecutionResult{
			{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  []byte("different"),
					HeaderNonce: 11,
					RootHash:    []byte("differentRoot"),
				},
			},
		},
	}
	err := tracker.CleanConfirmedExecutionResults(header)
	require.ErrorIs(t, err, ErrExecutionResultMismatch)

	batches := tracker.PopDismissedResults()
	require.Len(t, batches, 1)
	require.Len(t, batches[0].Results, 1)
	// Anchor should be lastNotarized (mismatch at idx=0)
	require.Equal(t, lastNotarized.GetRootHash(), batches[0].AnchorResult.GetRootHash())
	require.Equal(t, exec1.GetRootHash(), batches[0].Results[0].GetRootHash())
}

func TestExecutionResultsTracker_PopDismissedResults_OnClean(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	lastNotarized := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
			RootHash:    []byte("rootHash0"),
		},
	}
	_ = tracker.SetLastNotarizedResult(lastNotarized)

	exec1 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash1"),
			HeaderNonce: 11,
			RootHash:    []byte("rootHash1"),
		},
	}
	exec2 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash2"),
			HeaderNonce: 12,
			RootHash:    []byte("rootHash2"),
		},
	}
	_, _ = tracker.AddExecutionResult(exec1)
	_, _ = tracker.AddExecutionResult(exec2)

	newNotarized := &block.BaseExecutionResult{
		HeaderHash:  []byte("new"),
		HeaderNonce: 50,
		RootHash:    []byte("newRoot"),
	}
	tracker.Clean(newNotarized)

	batches := tracker.PopDismissedResults()
	require.Len(t, batches, 1)
	require.Len(t, batches[0].Results, 2)
	// Anchor should be the OLD lastNotarized (before Clean overwrote it)
	require.Equal(t, lastNotarized.GetRootHash(), batches[0].AnchorResult.GetRootHash())
	require.Equal(t, exec1.GetRootHash(), batches[0].Results[0].GetRootHash())
	require.Equal(t, exec2.GetRootHash(), batches[0].Results[1].GetRootHash())
}

func TestExecutionResultsTracker_PopDismissedResults_ConfirmedNotDismissed(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	lastNotarized := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
			RootHash:    []byte("rootHash0"),
		},
	}
	_ = tracker.SetLastNotarizedResult(lastNotarized)

	exec1 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash1"),
			HeaderNonce: 11,
			RootHash:    []byte("rootHash1"),
		},
	}
	_, _ = tracker.AddExecutionResult(exec1)

	// Confirm exec1 successfully - should NOT appear in dismissed
	header := &block.HeaderV3{
		ExecutionResults: []*block.ExecutionResult{exec1},
	}
	err := tracker.CleanConfirmedExecutionResults(header)
	require.NoError(t, err)

	batches := tracker.PopDismissedResults()
	require.Nil(t, batches)
}

func TestExecutionResultsTracker_PopDismissedResults_MultipleBatches(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	lastNotarized := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
			RootHash:    []byte("rootHash0"),
		},
	}
	_ = tracker.SetLastNotarizedResult(lastNotarized)

	exec1 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash1"),
			HeaderNonce: 11,
			RootHash:    []byte("rootHash1"),
		},
	}
	exec2 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash2"),
			HeaderNonce: 12,
			RootHash:    []byte("rootHash2"),
		},
	}
	exec3 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash3"),
			HeaderNonce: 13,
			RootHash:    []byte("rootHash3"),
		},
	}
	_, _ = tracker.AddExecutionResult(exec1)
	_, _ = tracker.AddExecutionResult(exec2)
	_, _ = tracker.AddExecutionResult(exec3)

	// First dismissal: remove nonce 13 and higher - dismisses exec3
	_ = tracker.RemoveFromNonce(13)

	// Second dismissal: remove nonce 12 and higher - dismisses exec2
	_ = tracker.RemoveFromNonce(12)

	// Should get 2 separate batches with different anchors
	batches := tracker.PopDismissedResults()
	require.Len(t, batches, 2)

	// Batch 1: exec3 dismissed, anchor = exec2 (last pending before nonce 13)
	require.Equal(t, exec2.GetRootHash(), batches[0].AnchorResult.GetRootHash())
	require.Len(t, batches[0].Results, 1)
	require.Equal(t, exec3.GetRootHash(), batches[0].Results[0].GetRootHash())

	// Batch 2: exec2 dismissed, anchor = exec1 (last pending before nonce 12)
	require.Equal(t, exec1.GetRootHash(), batches[1].AnchorResult.GetRootHash())
	require.Len(t, batches[1].Results, 1)
	require.Equal(t, exec2.GetRootHash(), batches[1].Results[0].GetRootHash())
}

func TestExecutionResultsTracker_PopDismissedResults_IndependentSources(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	lastNotarized := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
			RootHash:    []byte("rootHash0"),
		},
	}
	_ = tracker.SetLastNotarizedResult(lastNotarized)

	exec1 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash1"),
			HeaderNonce: 11,
			RootHash:    []byte("rootHash1"),
		},
	}
	exec2 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash2"),
			HeaderNonce: 12,
			RootHash:    []byte("rootHash2"),
		},
	}
	exec3 := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash3"),
			HeaderNonce: 13,
			RootHash:    []byte("rootHash3"),
		},
	}
	_, _ = tracker.AddExecutionResult(exec1)
	_, _ = tracker.AddExecutionResult(exec2)
	_, _ = tracker.AddExecutionResult(exec3)

	// Source 1: CleanOnConsensusReached dismisses nonce 12+ (consensus committed different hash for 12)
	header := &block.HeaderV3{Nonce: 12}
	tracker.CleanOnConsensusReached([]byte("different_hash"), header)

	// Source 2: RemoveFromNonce dismisses the remaining exec1
	_ = tracker.RemoveFromNonce(11)

	batches := tracker.PopDismissedResults()
	require.Len(t, batches, 2)

	// Batch 0: from CleanOnConsensusReached - dismissed exec2+exec3, anchor = exec1
	require.Equal(t, exec1.GetRootHash(), batches[0].AnchorResult.GetRootHash())
	require.Len(t, batches[0].Results, 2)
	require.Equal(t, exec2.GetRootHash(), batches[0].Results[0].GetRootHash())
	require.Equal(t, exec3.GetRootHash(), batches[0].Results[1].GetRootHash())

	// Batch 1: from RemoveFromNonce - dismissed exec1, anchor = lastNotarized
	require.Equal(t, lastNotarized.GetRootHash(), batches[1].AnchorResult.GetRootHash())
	require.Len(t, batches[1].Results, 1)
	require.Equal(t, exec1.GetRootHash(), batches[1].Results[0].GetRootHash())
}

func TestExecutionResultsTracker_DismissedBatchesOverflow(t *testing.T) {
	t.Parallel()

	tracker := NewExecutionResultsTracker()

	lastNotarized := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("hash0"),
			HeaderNonce: 10,
			RootHash:    []byte("rootHash0"),
		},
	}
	_ = tracker.SetLastNotarizedResult(lastNotarized)

	// Fill the queue to capacity by repeatedly adding+removing execution results
	for i := 0; i < maxDismissedBatches+5; i++ {
		nonce := uint64(11)
		exec := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte(fmt.Sprintf("hash_%d", i)),
				HeaderNonce: nonce,
				RootHash:    []byte(fmt.Sprintf("root_%d", i)),
			},
		}
		_, _ = tracker.AddExecutionResult(exec)
		_ = tracker.RemoveFromNonce(nonce)
	}

	batches := tracker.PopDismissedResults()
	// Should be capped at maxDismissedBatches, oldest batches dropped
	require.Len(t, batches, maxDismissedBatches)

	// The surviving batches should be the last maxDismissedBatches ones (oldest dropped)
	// The first surviving batch should be from iteration index 5 (0-4 were dropped)
	firstSurvivingIdx := 5
	require.Equal(t,
		[]byte(fmt.Sprintf("root_%d", firstSurvivingIdx)),
		batches[0].Results[0].GetRootHash(),
	)
	// All anchors should be lastNotarized (no pending results with nonce < 11 exist)
	for i, batch := range batches {
		require.Equal(t, lastNotarized.GetRootHash(), batch.AnchorResult.GetRootHash(),
			"batch %d should have lastNotarized as anchor", i)
	}

	// The last surviving batch should be from the last iteration
	lastIdx := maxDismissedBatches + 5 - 1
	require.Equal(t,
		[]byte(fmt.Sprintf("root_%d", lastIdx)),
		batches[maxDismissedBatches-1].Results[0].GetRootHash(),
	)

	// After pop, queue should be empty
	require.Nil(t, tracker.PopDismissedResults())
}
