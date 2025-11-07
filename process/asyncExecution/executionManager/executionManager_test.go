package executionManager_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/executionManager"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	"github.com/stretchr/testify/require"
)

var errExpected = errors.New("expected error")

func createMockArgs() executionManager.ArgsExecutionManager {
	return executionManager.ArgsExecutionManager{
		BlocksQueue:             &processMocks.BlocksQueueMock{},
		ExecutionResultsTracker: &processMocks.ExecutionTrackerStub{},
		BlockChain:              &testscommon.ChainHandlerMock{},
		Headers:                 &pool.HeadersPoolStub{},
	}
}

func TestNewExecutionManager(t *testing.T) {
	t.Parallel()

	t.Run("nil blocks queue should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.BlocksQueue = nil

		em, err := executionManager.NewExecutionManager(args)
		require.Nil(t, em)
		require.Equal(t, executionManager.ErrNilBlocksQueue, err)
	})

	t.Run("nil execution results tracker should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.ExecutionResultsTracker = nil

		em, err := executionManager.NewExecutionManager(args)
		require.Nil(t, em)
		require.Equal(t, executionManager.ErrNilExecutionResultsTracker, err)
	})

	t.Run("nil blockchain should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.BlockChain = nil

		em, err := executionManager.NewExecutionManager(args)
		require.Nil(t, em)
		require.Equal(t, executionManager.ErrNilBlockchain, err)
	})

	t.Run("nil headers pool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Headers = nil

		em, err := executionManager.NewExecutionManager(args)
		require.Nil(t, em)
		require.Equal(t, executionManager.ErrNilHeadersPool, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		em, err := executionManager.NewExecutionManager(args)
		require.NoError(t, err)
		require.NotNil(t, em)
	})
}

func TestExecutionManager_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.Headers = nil
	em, _ := executionManager.NewExecutionManager(args)
	require.True(t, em.IsInterfaceNil())

	em, _ = executionManager.NewExecutionManager(createMockArgs())
	require.False(t, em.IsInterfaceNil())
}

func TestExecutionManager_StartExecution(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	em, _ := executionManager.NewExecutionManager(args)

	startCalled := false
	mockExecutor := &processMocks.HeadersExecutorStub{
		StartExecutionCalled: func() {
			startCalled = true
		},
	}
	_ = em.SetHeadersExecutor(mockExecutor)

	em.StartExecution()
	require.True(t, startCalled)
}

func TestExecutionManager_SetHeadersExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil headers executor should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		em, _ := executionManager.NewExecutionManager(args)

		err := em.SetHeadersExecutor(nil)
		require.Equal(t, executionManager.ErrNilHeadersExecutor, err)
	})

	t.Run("valid headers executor should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		em, _ := executionManager.NewExecutionManager(args)

		mockExecutor := &processMocks.HeadersExecutorStub{}
		err := em.SetHeadersExecutor(mockExecutor)
		require.NoError(t, err)
	})
}

func TestExecutionManager_AddPairForExecution(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	addOrReplaceCalled := false
	args.BlocksQueue = &processMocks.BlocksQueueMock{
		AddOrReplaceCalled: func(pair queue.HeaderBodyPair) error {
			addOrReplaceCalled = true
			require.NotNil(t, pair.Header)
			return nil
		},
	}
	em, _ := executionManager.NewExecutionManager(args)

	pair := queue.HeaderBodyPair{
		Header: &block.Header{Nonce: 1},
		Body:   &block.Body{},
	}
	err := em.AddPairForExecution(pair)
	require.NoError(t, err)
	require.True(t, addOrReplaceCalled)
}

func TestExecutionManager_GetPendingExecutionResults(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	expectedResults := []data.BaseExecutionResultHandler{
		&block.ExecutionResult{},
	}
	args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
		GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
			return expectedResults, nil
		},
	}
	em, _ := executionManager.NewExecutionManager(args)

	results, err := em.GetPendingExecutionResults()
	require.NoError(t, err)
	require.Equal(t, expectedResults, results)
}

func TestExecutionManager_SetLastNotarizedResult(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	setLastNotarizedCalled := false
	args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
		SetLastNotarizedResultCalled: func(executionResult data.BaseExecutionResultHandler) error {
			setLastNotarizedCalled = true
			return nil
		},
	}
	em, _ := executionManager.NewExecutionManager(args)

	execResult := &block.ExecutionResult{}
	err := em.SetLastNotarizedResult(execResult)
	require.NoError(t, err)
	require.True(t, setLastNotarizedCalled)
}

func TestExecutionManager_CleanConfirmedExecutionResults(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	cleanCalled := false
	args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
		CleanConfirmedExecutionResultsCalled: func(header data.HeaderHandler) error {
			cleanCalled = true
			return nil
		},
	}
	em, _ := executionManager.NewExecutionManager(args)

	header := &block.Header{Nonce: 1}
	err := em.CleanConfirmedExecutionResults(header)
	require.NoError(t, err)
	require.True(t, cleanCalled)
}

func TestExecutionManager_RemoveAtNonceAndHigher(t *testing.T) {
	t.Parallel()

	t.Run("error getting last notarized result should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return nil, errExpected
			},
		}
		em, _ := executionManager.NewExecutionManager(args)

		err := em.RemoveAtNonceAndHigher(5)
		require.Equal(t, errExpected, err)
	})

	t.Run("nonce lower than last notarized should adjust nonce", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		pauseCalled := false
		resumeCalled := false
		args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return &block.ExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: 10,
					},
				}, nil
			},
		}
		args.BlocksQueue = &processMocks.BlocksQueueMock{
			RemoveAtNonceAndHigherCalled: func(nonce uint64) []uint64 {
				// Should be adjusted to lastNotarizedNonce + 1 = 11
				require.Equal(t, uint64(11), nonce)
				return []uint64{11}
			},
		}
		em, _ := executionManager.NewExecutionManager(args)
		mockExecutor := &processMocks.HeadersExecutorStub{
			PauseExecutionCalled: func() {
				pauseCalled = true
			},
			ResumeExecutionCalled: func() {
				resumeCalled = true
			},
		}
		_ = em.SetHeadersExecutor(mockExecutor)

		err := em.RemoveAtNonceAndHigher(5) // Lower than last notarized (10)
		require.NoError(t, err)
		require.True(t, pauseCalled)
		require.True(t, resumeCalled)
	})

	t.Run("nonce still in queue should resume execution", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		pauseCalled := false
		resumeCalled := false
		args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return &block.ExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: 9,
					},
				}, nil
			},
			RemoveFromNonceCalled: func(nonce uint64) error {
				require.Fail(t, "should not have been called")
				return nil
			},
		}
		args.BlocksQueue = &processMocks.BlocksQueueMock{
			RemoveAtNonceAndHigherCalled: func(nonce uint64) []uint64 {
				require.Equal(t, uint64(10), nonce)
				// First removed nonce matches the requested nonce
				return []uint64{10, 11, 12}
			},
		}
		em, _ := executionManager.NewExecutionManager(args)
		mockExecutor := &processMocks.HeadersExecutorStub{
			PauseExecutionCalled: func() {
				pauseCalled = true
			},
			ResumeExecutionCalled: func() {
				resumeCalled = true
			},
		}
		_ = em.SetHeadersExecutor(mockExecutor)

		err := em.RemoveAtNonceAndHigher(10)
		require.NoError(t, err)
		require.True(t, pauseCalled)
		require.True(t, resumeCalled)
	})

	t.Run("nonce already popped should clean tracker and update blockchain", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		pauseCalled := false
		resumeCalled := false
		removeFromNonceCalled := false
		args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return &block.ExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: 5,
						HeaderHash:  []byte("hash5"),
						RootHash:    []byte("root5"),
					},
				}, nil
			},
			RemoveFromNonceCalled: func(nonce uint64) error {
				removeFromNonceCalled = true
				require.Equal(t, uint64(10), nonce)
				return nil
			},
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return []data.BaseExecutionResultHandler{}, nil
			},
		}
		args.BlocksQueue = &processMocks.BlocksQueueMock{
			RemoveAtNonceAndHigherCalled: func(nonce uint64) []uint64 {
				require.Equal(t, uint64(10), nonce)
				// First removed nonce does NOT match the requested nonce
				return []uint64{11, 12}
			},
		}
		header := &block.Header{Nonce: 5}
		args.Headers = &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return header, nil
			},
		}
		chainMock := &testscommon.ChainHandlerMock{}
		args.BlockChain = chainMock

		em, _ := executionManager.NewExecutionManager(args)
		mockExecutor := &processMocks.HeadersExecutorStub{
			PauseExecutionCalled: func() {
				pauseCalled = true
			},
			ResumeExecutionCalled: func() {
				resumeCalled = true
			},
		}
		_ = em.SetHeadersExecutor(mockExecutor)

		err := em.RemoveAtNonceAndHigher(10)
		require.NoError(t, err)
		require.True(t, pauseCalled)
		require.True(t, resumeCalled)
		require.True(t, removeFromNonceCalled)

		// Verify blockchain was updated
		nonce, hash, rootHash := chainMock.GetFinalBlockInfo()
		require.Equal(t, uint64(5), nonce)
		require.Equal(t, []byte("hash5"), hash)
		require.Equal(t, []byte("root5"), rootHash)
	})

	t.Run("error from tracker remove should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return &block.ExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: 5,
					},
				}, nil
			},
			RemoveFromNonceCalled: func(nonce uint64) error {
				return errExpected
			},
		}
		args.BlocksQueue = &processMocks.BlocksQueueMock{
			RemoveAtNonceAndHigherCalled: func(nonce uint64) []uint64 {
				// First removed nonce does NOT match
				return []uint64{11, 12}
			},
		}
		em, _ := executionManager.NewExecutionManager(args)

		err := em.RemoveAtNonceAndHigher(10)
		require.Equal(t, errExpected, err)
	})

	t.Run("error getting header from pool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return &block.ExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: 5,
						HeaderHash:  []byte("hash5"),
					},
				}, nil
			},
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return []data.BaseExecutionResultHandler{}, nil
			},
		}
		args.BlocksQueue = &processMocks.BlocksQueueMock{
			RemoveAtNonceAndHigherCalled: func(nonce uint64) []uint64 {
				return []uint64{11, 12}
			},
		}
		args.Headers = &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return nil, errExpected
			},
		}
		em, _ := executionManager.NewExecutionManager(args)

		err := em.RemoveAtNonceAndHigher(10)
		require.Equal(t, errExpected, err)
	})

	t.Run("with pending execution results should use last pending", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return &block.ExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: 5,
						HeaderHash:  []byte("hash5"),
						RootHash:    []byte("root5"),
					},
				}, nil
			},
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return []data.BaseExecutionResultHandler{
					&block.ExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderNonce: 6,
							HeaderHash:  []byte("hash6"),
							RootHash:    []byte("root6"),
						},
					},
					&block.ExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderNonce: 7,
							HeaderHash:  []byte("hash7"),
							RootHash:    []byte("root7"),
						},
					},
				}, nil
			},
		}
		args.BlocksQueue = &processMocks.BlocksQueueMock{
			RemoveAtNonceAndHigherCalled: func(nonce uint64) []uint64 {
				return []uint64{11, 12}
			},
		}
		header := &block.Header{Nonce: 7}
		args.Headers = &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				require.Equal(t, []byte("hash7"), hash)
				return header, nil
			},
		}
		chainMock := &testscommon.ChainHandlerMock{}
		args.BlockChain = chainMock

		em, _ := executionManager.NewExecutionManager(args)

		err := em.RemoveAtNonceAndHigher(10)
		require.NoError(t, err)

		// Verify blockchain was updated with last pending
		nonce, hash, rootHash := chainMock.GetFinalBlockInfo()
		require.Equal(t, uint64(7), nonce)
		require.Equal(t, []byte("hash7"), hash)
		require.Equal(t, []byte("root7"), rootHash)

		lastExecHeader := chainMock.GetLastExecutedBlockHeader()
		require.Equal(t, header, lastExecHeader)
	})

	t.Run("error getting pending execution results should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return &block.ExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: 5,
						HeaderHash:  []byte("hash5"),
					},
				}, nil
			},
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, errExpected
			},
		}
		args.BlocksQueue = &processMocks.BlocksQueueMock{
			RemoveAtNonceAndHigherCalled: func(nonce uint64) []uint64 {
				// First removed nonce does NOT match
				return []uint64{11, 12}
			},
		}
		em, _ := executionManager.NewExecutionManager(args)

		err := em.RemoveAtNonceAndHigher(10)
		require.Equal(t, errExpected, err)
	})
}

func TestExecutionManager_Close(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		executorCloseCalled := false
		queueCloseCalled := false

		mockExecutor := &processMocks.HeadersExecutorStub{
			CloseCalled: func() error {
				executorCloseCalled = true
				return nil
			},
		}
		args.BlocksQueue = &processMocks.BlocksQueueMock{
			CloseCalled: func() {
				queueCloseCalled = true
			},
		}
		em, _ := executionManager.NewExecutionManager(args)
		_ = em.SetHeadersExecutor(mockExecutor)

		err := em.Close()
		require.NoError(t, err)
		require.True(t, executorCloseCalled)
		require.True(t, queueCloseCalled)
	})

	t.Run("error closing headers executor should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		mockExecutor := &processMocks.HeadersExecutorStub{
			CloseCalled: func() error {
				return errExpected
			},
		}
		em, _ := executionManager.NewExecutionManager(args)
		_ = em.SetHeadersExecutor(mockExecutor)

		err := em.Close()
		require.Equal(t, errExpected, err)
	})
}

func TestExecutionManager_Concurrency(t *testing.T) {
	require.NotPanics(t, func() {
		t.Parallel()

		args := createMockArgs()
		args.ExecutionResultsTracker = &processMocks.ExecutionTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return []data.BaseExecutionResultHandler{}, nil
			},
			GetLastNotarizedExecutionResultCalled: func() (data.BaseExecutionResultHandler, error) {
				return &block.ExecutionResult{}, nil
			},
		}
		args.BlockChain = &testscommon.ChainHandlerStub{}
		em, _ := executionManager.NewExecutionManager(args)

		const numCalls = 100
		wg := sync.WaitGroup{}
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(idx int) {
				defer wg.Done()

				switch idx % 2 {
				case 0:
					pair := queue.HeaderBodyPair{
						Header: &block.Header{Nonce: uint64(idx)},
						Body:   &block.Body{},
					}
					_ = em.AddPairForExecution(pair)
				case 1:
					_ = em.RemoveAtNonceAndHigher(uint64(idx))
				}
			}(i)
		}

		wg.Wait()
	})
}
