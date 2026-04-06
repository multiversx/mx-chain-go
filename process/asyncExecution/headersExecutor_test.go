package asyncExecution

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	atomicCore "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
)

var errExpected = errors.New("expected error")

func createMockArgs() ArgsHeadersExecutor {
	headerCache := cache.NewHeaderBodyCache(config.HeaderBodyCacheConfig{})

	return ArgsHeadersExecutor{
		BlocksCache:      headerCache,
		ExecutionTracker: &processMocks.ExecutionTrackerStub{},
		BlockProcessor:   &processMocks.BlockProcessorStub{},
		BlockChain: &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.BaseExecutionResult{}
			},
		},
	}
}

func TestNewHeadersExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil headers queue", func(t *testing.T) {
		args := createMockArgs()
		args.BlocksCache = nil

		_, err := NewHeadersExecutor(args)
		require.Equal(t, ErrNilHeadersCache, err)
	})

	t.Run("nil execution tracker", func(t *testing.T) {
		args := createMockArgs()
		args.ExecutionTracker = nil

		_, err := NewHeadersExecutor(args)
		require.Equal(t, ErrNilExecutionTracker, err)
	})

	t.Run("nil block processor", func(t *testing.T) {
		args := createMockArgs()
		args.BlockProcessor = nil

		_, err := NewHeadersExecutor(args)
		require.Equal(t, ErrNilBlockProcessor, err)
	})

	t.Run("nil chain handler", func(t *testing.T) {
		args := createMockArgs()
		args.BlockChain = nil

		_, err := NewHeadersExecutor(args)
		require.Equal(t, process.ErrNilBlockChain, err)
	})

	t.Run("should work", func(t *testing.T) {
		args := createMockArgs()

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)
		require.NotNil(t, executor)
		require.False(t, executor.IsInterfaceNil())

		err = executor.Close()
		require.NoError(t, err)
	})
}

func TestHeadersExecutor_StartAndClose(t *testing.T) {
	t.Parallel()

	var prevHash = []byte("prevHash")
	calledProcessBlock := uint32(0)
	calledAddExecutionResult := uint32(0)
	args := createMockArgs()
	blocksCache := cache.NewHeaderBodyCache(config.HeaderBodyCacheConfig{})
	args.BlocksCache = blocksCache
	executedNonce := uint64(1)
	executedHash := prevHash
	args.BlockChain = &testscommon.ChainHandlerStub{
		GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
			return &block.HeaderV3{
				Nonce: executedNonce,
			}
		},
		GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
			return &block.BaseExecutionResult{
				HeaderNonce: executedNonce,
				HeaderHash:  executedHash,
			}
		},
		GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
			return executedNonce, executedHash, nil
		},
		SetLastExecutedBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, blockHash []byte, rootHash []byte) {
			executedNonce = header.GetNonce()
			executedHash = blockHash
		},
	}

	args.BlockProcessor = &processMocks.BlockProcessorStub{
		ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
			atomic.AddUint32(&calledProcessBlock, 1)
			return &block.BaseExecutionResult{
				HeaderNonce: handler.GetNonce(),
				HeaderHash:  headerHash,
			}, nil
		},
	}
	args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
		AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
			atomic.AddUint32(&calledAddExecutionResult, 1)
			return true, nil
		},
	}

	executor, err := NewHeadersExecutor(args)
	require.NoError(t, err)

	executor.StartExecution()

	err = blocksCache.AddOrReplace(cache.HeaderBodyPair{
		Header: &block.Header{
			Nonce:    2,
			PrevHash: prevHash,
		},
		Body:       &block.Body{},
		HeaderHash: []byte("a"),
	})
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	err = executor.Close()
	require.NoError(t, err)
	require.Equal(t, uint32(1), atomic.LoadUint32(&calledProcessBlock))
	require.Equal(t, uint32(1), atomic.LoadUint32(&calledAddExecutionResult))

}

func TestHeadersExecutor_ProcessBlock(t *testing.T) {
	t.Parallel()

	var prevHash = []byte("prevHash")
	var currentHash = []byte("currentHash")

	t.Run("pause/resume should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		cntWasPopCalled := uint32(0)
		lastExecutedBlockNonce := 0

		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				if lastExecutedBlockNonce > 0 {
					return &block.HeaderV3{
						Nonce:    uint64(lastExecutedBlockNonce),
						PrevHash: prevHash,
					}
				}
				return &block.HeaderV3{
					Nonce: 0,
				}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				if lastExecutedBlockNonce > 0 {
					return &block.BaseExecutionResult{
						HeaderNonce: uint64(lastExecutedBlockNonce),
						HeaderHash:  currentHash,
					}
				}
				return &block.BaseExecutionResult{
					HeaderNonce: 0,
					HeaderHash:  prevHash,
				}
			},
			GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
				if lastExecutedBlockNonce > 0 {
					return uint64(lastExecutedBlockNonce), currentHash, nil
				}
				return 0, prevHash, nil
			},
			SetLastExecutedBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, blockHash []byte, rootHash []byte) {
				lastExecutedBlockNonce = int(header.GetNonce())
			},
		}
		args.BlocksCache = &processMocks.BlocksCacheMock{
			GetByNonceCalled: func(nonce uint64) (cache.HeaderBodyPair, bool) {
				atomic.AddUint32(&cntWasPopCalled, 1)

				return cache.HeaderBodyPair{
					Header: &block.Header{
						Nonce:    1,
						PrevHash: prevHash,
					},
					Body: &block.Body{},
				}, true
			},
		}

		wasProcessBlockProposalCalled := atomicCore.Flag{}
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				wasProcessBlockProposalCalled.SetValue(true)
				time.Sleep(time.Millisecond * 500) // force pause to be called first

				return &block.BaseExecutionResult{
					HeaderNonce: handler.GetNonce(),
					HeaderHash:  headerHash,
				}, nil
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		executor.PauseExecution() // coverage, should early exit

		executor.StartExecution()

		// allow some Pop operations
		time.Sleep(time.Millisecond * 200)

		require.True(t, wasProcessBlockProposalCalled.IsSet()) // require this here so we know the processing is in progress

		executor.PauseExecution() // blocks until any ongoing processing completes
		cntWasPopCalledAtPause := atomic.LoadUint32(&cntWasPopCalled)

		executor.PauseExecution() // coverage, already paused

		cntWasPopCalledBeforeResume := atomic.LoadUint32(&cntWasPopCalled)
		require.Equal(t, cntWasPopCalledAtPause, cntWasPopCalledBeforeResume)

		executor.ResumeExecution()

		time.Sleep(time.Millisecond * 200)

		cntWasPopCalledAfterResume := atomic.LoadUint32(&cntWasPopCalled)
		require.Greater(t, cntWasPopCalledAfterResume, cntWasPopCalledBeforeResume)

		err = executor.Close()
		require.NoError(t, err)
	})

	t.Run("concurrent pause/resume should work", func(t *testing.T) {
		t.Parallel()

		require.NotPanics(t, func() {
			args := createMockArgs()
			args.BlockProcessor = &processMocks.BlockProcessorStub{
				ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
					return &block.BaseExecutionResult{
						HeaderHash: headerHash,
					}, nil
				},
			}
			args.BlockChain = &testscommon.ChainHandlerStub{
				GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
					return &block.BaseExecutionResult{}
				},
			}

			executor, err := NewHeadersExecutor(args)
			require.NoError(t, err)

			executor.StartExecution()

			wg := sync.WaitGroup{}
			numCalls := 100
			wg.Add(numCalls)
			for i := 0; i < numCalls; i++ {
				go func(idx int) {
					defer wg.Done()

					switch idx % 2 {
					case 0:
						executor.PauseExecution()
					case 1:
						executor.ResumeExecution()
					default:
						require.Fail(t, "should not happen")
					}

					time.Sleep(time.Millisecond * 3)
				}(i)
			}

			wg.Wait()

			err = executor.Close()
			require.NoError(t, err)
		})
	})

	t.Run("add execution result error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		blocksCache := cache.NewHeaderBodyCache(config.HeaderBodyCacheConfig{})
		args.BlocksCache = blocksCache
		wasAddExecutionResultCalled := atomicCore.Flag{}
		executedNonce := uint64(0)
		executedHash := prevHash
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.HeaderV3{
					Nonce: executedNonce,
				}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.BaseExecutionResult{
					HeaderNonce: executedNonce,
					HeaderHash:  executedHash,
				}
			},
			GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
				return executedNonce, executedHash, nil
			},
			SetLastExecutedBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, blockHash []byte, rootHash []byte) {
				executedNonce = header.GetNonce()
				executedHash = blockHash
			},
		}
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderNonce: handler.GetNonce(),
					HeaderHash:  headerHash,
				}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				if wasAddExecutionResultCalled.IsSet() {
					// return nil after first call to allow processing to complete
					return true, nil
				}
				wasAddExecutionResultCalled.SetValue(true)
				return false, errExpected
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		err = blocksCache.AddOrReplace(cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce:    1,
				PrevHash: prevHash,
			},
			Body:       &block.Body{},
			HeaderHash: []byte("a"),
		})
		require.NoError(t, err)

		executor.StartExecution()

		// allow Pop operation
		time.Sleep(time.Millisecond * 100)

		err = executor.Close()
		require.NoError(t, err)
		require.True(t, wasAddExecutionResultCalled.IsSet())
	})

	t.Run("block processing error, after retry should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		blocksCache := cache.NewHeaderBodyCache(config.HeaderBodyCacheConfig{})
		count := 0
		countAddResult := 0
		args.BlocksCache = blocksCache
		wg := &sync.WaitGroup{}
		wg.Add(1)
		nonce := uint64(0)
		executedHash := prevHash
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.HeaderV3{
					Nonce: nonce,
				}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.BaseExecutionResult{
					HeaderNonce: nonce,
					HeaderHash:  executedHash,
				}
			},
			GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
				return nonce, executedHash, nil
			},
			SetLastExecutedBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, blockHash []byte, rootHash []byte) {
				nonce = header.GetNonce()
				executedHash = blockHash
			},
		}
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				if count == 1 {
					return &block.BaseExecutionResult{
						HeaderNonce: handler.GetNonce(),
						HeaderHash:  headerHash,
					}, nil
				}
				count++
				return nil, errExpected
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				countAddResult++
				wg.Done()
				return true, nil
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		executor.StartExecution()

		err = blocksCache.AddOrReplace(cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce:    1,
				PrevHash: prevHash,
			},
			Body:       &block.Body{},
			HeaderHash: []byte("a"),
		})
		require.NoError(t, err)

		wg.Wait()
		err = executor.Close()
		require.NoError(t, err)
		require.Equal(t, 1, countAddResult)
	})

	t.Run("block processing error, different hash should early exit", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		blocksCache := cache.NewHeaderBodyCache(config.HeaderBodyCacheConfig{})

		count := 0
		countAddResult := 0
		nonce := uint64(0)
		executedHash := prevHash
		args.BlocksCache = blocksCache
		wg := &sync.WaitGroup{}
		wg.Add(1)
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.HeaderV3{
					Nonce: nonce,
				}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.BaseExecutionResult{
					HeaderNonce: nonce,
					HeaderHash:  executedHash,
				}
			},
			GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
				return nonce, executedHash, nil
			},
			SetLastExecutedBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, blockHash []byte, rootHash []byte) {
				nonce = header.GetNonce()
				executedHash = blockHash
			},
		}
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				time.Sleep(time.Millisecond)
				if handler.GetRound() == 1 {
					return nil, errExpected
				}

				count++
				return &block.BaseExecutionResult{
					HeaderNonce: handler.GetNonce(),
					HeaderHash:  headerHash,
				}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				countAddResult++
				wg.Done()
				return true, nil
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		executor.StartExecution()

		err = blocksCache.AddOrReplace(cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce:    1,
				Round:    1,
				PrevHash: prevHash,
			},
			Body:       &block.Body{},
			HeaderHash: []byte("a"),
		})
		require.NoError(t, err)

		time.Sleep(time.Millisecond)
		err = blocksCache.AddOrReplace(cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce:    1,
				Round:    2,
				PrevHash: prevHash,
			},
			Body:       &block.Body{},
			HeaderHash: currentHash,
		})
		require.NoError(t, err)

		wg.Wait()

		require.Equal(t, 1, count)
	})
}

func TestHeadersExecutor_Process(t *testing.T) {
	t.Parallel()

	var testPrevHash = []byte("prevHash")
	var testDifferentHash = []byte("differentHash")
	var testCommittedHash = []byte("committedHash")
	var testChangedHash = []byte("changedHash")
	var testNonce = uint64(1)

	t.Run("should return error on failing to process block", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		expectedErr := errors.New("expected error")
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return nil, expectedErr
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: testNonce,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return error on failing to add execution results to execution tracker", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		countAddResult := 0

		expectedErr := errors.New("expected error")
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderHash: headerHash,
				}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				countAddResult++
				return false, expectedErr
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: testNonce,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should call CommitBlockProposalState when all checks pass", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		commitCalled := false
		revertCalled := false
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderHash: headerHash,
				}, nil
			},
			CommitBlockProposalStateCalled: func(headerHandler data.HeaderHandler) error {
				commitCalled = true
				return nil
			},
			RevertBlockProposalStateCalled: func() {
				revertCalled = true
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				return true, nil
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: testNonce,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Nil(t, err)
		require.True(t, commitCalled)
		require.False(t, revertCalled)
	})

	t.Run("should call RevertBlockProposalState when nil execution result is returned", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		revertCalled := false
		commitCalled := false
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return nil, nil
			},
			CommitBlockProposalStateCalled: func(headerHandler data.HeaderHandler) error {
				commitCalled = true
				return nil
			},
			RevertBlockProposalStateCalled: func() {
				revertCalled = true
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: testNonce,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Equal(t, ErrNilExecutionResult, err)
		require.True(t, revertCalled)
		require.False(t, commitCalled)
	})

	t.Run("should call RevertBlockProposalState when context mismatch after processing", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		revertCalled := false
		commitCalled := false
		callCount := 0
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				callCount++
				if callCount == 1 {
					// First call (before processing) - return valid context
					return &block.BaseExecutionResult{
						HeaderNonce: 0,
						HeaderHash:  testPrevHash,
					}
				}
				// Second call (after processing) - return mismatched context to trigger revert
				return &block.BaseExecutionResult{
					HeaderNonce: 5,
					HeaderHash:  testDifferentHash,
				}
			},
		}
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderHash: headerHash,
				}, nil
			},
			CommitBlockProposalStateCalled: func(headerHandler data.HeaderHandler) error {
				commitCalled = true
				return nil
			},
			RevertBlockProposalStateCalled: func() {
				revertCalled = true
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce:    testNonce,
				PrevHash: testPrevHash,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Nil(t, err)
		require.True(t, revertCalled)
		require.False(t, commitCalled)
	})

	t.Run("should call RevertBlockProposalState when committed block hash differs", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		revertCalled := false
		commitCalled := false
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.BaseExecutionResult{}
			},
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Nonce: testNonce,
				}
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return testCommittedHash
			},
		}
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderNonce: testNonce,
					HeaderHash:  testDifferentHash,
				}, nil
			},
			CommitBlockProposalStateCalled: func(headerHandler data.HeaderHandler) error {
				commitCalled = true
				return nil
			},
			RevertBlockProposalStateCalled: func() {
				revertCalled = true
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: testNonce,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Nil(t, err)
		require.True(t, revertCalled)
		require.False(t, commitCalled)
	})

	t.Run("should call RevertBlockProposalState when last execution result header hash mismatches prevHash", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		revertCalled := false
		commitCalled := false
		getLastExecResultCallCount := 0
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				getLastExecResultCallCount++
				if getLastExecResultCallCount <= 2 {
					// First two calls (checkLastExecutionResultContext before and after processing):
					// return matching context so those checks pass
					return &block.BaseExecutionResult{
						HeaderNonce: 0,
						HeaderHash:  testPrevHash,
					}
				}
				// Third call (explicit prevHash comparison):
				// return mismatched header hash to trigger the revert
				return &block.BaseExecutionResult{
					HeaderNonce: 0,
					HeaderHash:  testChangedHash,
				}
			},
		}
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderHash: headerHash,
				}, nil
			},
			CommitBlockProposalStateCalled: func(headerHandler data.HeaderHandler) error {
				commitCalled = true
				return nil
			},
			RevertBlockProposalStateCalled: func() {
				revertCalled = true
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce:    testNonce,
				PrevHash: testPrevHash,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Nil(t, err)
		require.True(t, revertCalled)
		require.False(t, commitCalled)
	})

	t.Run("should call RevertBlockProposalState when CommitBlockProposalState fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		revertCalled := false
		expectedErr := errors.New("commit failed")
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderHash: headerHash,
				}, nil
			},
			CommitBlockProposalStateCalled: func(headerHandler data.HeaderHandler) error {
				return expectedErr
			},
			RevertBlockProposalStateCalled: func() {
				revertCalled = true
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: testNonce,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Equal(t, expectedErr, err)
		require.True(t, revertCalled)
	})

	t.Run("should not call RevertBlockProposalState when AddExecutionResult fails after commit", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		commitCalled := false
		revertCalled := false
		expectedErr := errors.New("add result failed")
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderHash: headerHash,
				}, nil
			},
			CommitBlockProposalStateCalled: func(headerHandler data.HeaderHandler) error {
				commitCalled = true
				return nil
			},
			RevertBlockProposalStateCalled: func() {
				revertCalled = true
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				return false, expectedErr
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: testNonce,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Equal(t, expectedErr, err)
		require.True(t, commitCalled)
		require.False(t, revertCalled)
	})

	t.Run("should not call RevertBlockProposalState when AddExecutionResult rejects after commit", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		commitCalled := false
		revertCalled := false
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderHash: headerHash,
				}, nil
			},
			CommitBlockProposalStateCalled: func(headerHandler data.HeaderHandler) error {
				commitCalled = true
				return nil
			},
			RevertBlockProposalStateCalled: func() {
				revertCalled = true
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				return false, nil
			},
		}

		setFinalBlockInfoCalled := false
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.BaseExecutionResult{}
			},
			SetFinalBlockInfoCalled: func(nonce uint64, headerHash, rootHash []byte) {
				setFinalBlockInfoCalled = true
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: testNonce,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Nil(t, err)
		require.True(t, commitCalled)
		require.False(t, revertCalled)
		require.False(t, setFinalBlockInfoCalled)
	})

	t.Run("should add execution result info to blockchain handler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderHash: headerHash,
				}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				return true, nil
			},
		}

		setFinalBlockInfoCalled := false
		setLastExecutedBlockInfoCalled := false
		setLastExecutionResultCalled := false
		args.BlockChain = &testscommon.ChainHandlerStub{
			SetFinalBlockInfoCalled: func(nonce uint64, headerHash, rootHash []byte) {
				setFinalBlockInfoCalled = true
			},
			SetLastExecutedBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, headerHash, rootHash []byte) {
				setLastExecutedBlockInfoCalled = true
			},
			SetLastExecutionResultCalled: func(result data.BaseExecutionResultHandler) {
				setLastExecutionResultCalled = true
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.BaseExecutionResult{}
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: testNonce,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Nil(t, err)

		require.True(t, setFinalBlockInfoCalled)
		require.True(t, setLastExecutedBlockInfoCalled)
		require.True(t, setLastExecutionResultCalled)
	})
}

func TestHeadersExecutor_GetSignalProcessCompletionChan(t *testing.T) {
	t.Parallel()

	t.Run("should return nil when channel is not set", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SignalProcessCompletionChan = nil

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		ch := executor.GetSignalProcessCompletionChan()
		require.Nil(t, ch)

		err = executor.Close()
		require.NoError(t, err)
	})

	t.Run("should return channel when set", func(t *testing.T) {
		t.Parallel()

		signalChan := make(chan uint64, 1)
		args := createMockArgs()
		args.SignalProcessCompletionChan = signalChan

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		ch := executor.GetSignalProcessCompletionChan()
		require.NotNil(t, ch)
		require.Equal(t, signalChan, ch)

		err = executor.Close()
		require.NoError(t, err)
	})
}

func TestHeadersExecutor_SignalProcessCompletion(t *testing.T) {
	t.Parallel()

	t.Run("should signal channel on successful process", func(t *testing.T) {
		t.Parallel()

		signalChan := make(chan uint64, 1)
		args := createMockArgs()
		args.SignalProcessCompletionChan = signalChan

		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderNonce: handler.GetNonce(),
					HeaderHash:  headerHash,
				}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				return true, nil
			},
		}
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return nil // Return nil so checkLastExecutionResultContext passes
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: 5,
			},
			Body:       &block.Body{},
			HeaderHash: []byte("hash"),
		}

		err = executor.Process(pair)
		require.NoError(t, err)

		// Check that signal was sent
		select {
		case nonce := <-signalChan:
			require.Equal(t, uint64(5), nonce)
		default:
			require.Fail(t, "expected signal on channel")
		}

		err = executor.Close()
		require.NoError(t, err)
	})

	t.Run("should not block when channel is full", func(t *testing.T) {
		t.Parallel()

		signalChan := make(chan uint64, 1)
		// Pre-fill the channel
		signalChan <- 1

		args := createMockArgs()
		args.SignalProcessCompletionChan = signalChan

		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderNonce: handler.GetNonce(),
					HeaderHash:  headerHash,
				}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				return true, nil
			},
		}
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return nil // Return nil so checkLastExecutionResultContext passes
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: 5,
			},
			Body:       &block.Body{},
			HeaderHash: []byte("hash"),
		}

		// Should not block even though channel is full
		done := make(chan struct{})
		go func() {
			err := executor.Process(pair)
			require.NoError(t, err)
			close(done)
		}()

		select {
		case <-done:
			// Success - process completed without blocking
		case <-time.After(time.Second):
			require.Fail(t, "process blocked when channel was full")
		}

		err = executor.Close()
		require.NoError(t, err)
	})

	t.Run("should not panic when channel is nil", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SignalProcessCompletionChan = nil

		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{
					HeaderNonce: handler.GetNonce(),
					HeaderHash:  headerHash,
				}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) (bool, error) {
				return true, nil
			},
		}
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return nil // Return nil so checkLastExecutionResultContext passes
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		pair := cache.HeaderBodyPair{
			Header: &block.Header{
				Nonce: 5,
			},
			Body:       &block.Body{},
			HeaderHash: []byte("hash"),
		}

		require.NotPanics(t, func() {
			err := executor.Process(pair)
			require.NoError(t, err)
		})

		err = executor.Close()
		require.NoError(t, err)
	})
}
