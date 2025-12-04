package asyncExecution

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
)

var errExpected = errors.New("expected error")

func createMockArgs() ArgsHeadersExecutor {
	headerQueue := queue.NewBlocksQueue()

	return ArgsHeadersExecutor{
		BlocksQueue:      headerQueue,
		ExecutionTracker: &processMocks.ExecutionTrackerStub{},
		BlockProcessor:   &processMocks.BlockProcessorStub{},
		BlockChain:       &testscommon.ChainHandlerStub{},
	}
}

func TestNewHeadersExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil headers queue", func(t *testing.T) {
		args := createMockArgs()
		args.BlocksQueue = nil

		_, err := NewHeadersExecutor(args)
		require.Equal(t, ErrNilHeadersQueue, err)
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

	calledProcessBlock := 0
	calledAddExecutionResult := 0
	args := createMockArgs()
	blocksQueue := queue.NewBlocksQueue()
	args.BlocksQueue = blocksQueue
	args.BlockProcessor = &processMocks.BlockProcessorStub{
		ProcessBlockProposalCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
			calledProcessBlock++
			return &block.BaseExecutionResult{}, nil
		},
	}
	args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
		AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) error {
			calledAddExecutionResult++
			return nil
		},
	}

	executor, err := NewHeadersExecutor(args)
	require.NoError(t, err)

	executor.StartExecution()

	err = blocksQueue.AddOrReplace(queue.HeaderBodyPair{
		Header: &block.Header{},
		Body:   &block.Body{},
	})
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	err = executor.Close()
	require.NoError(t, err)
	require.Equal(t, 1, calledProcessBlock)
	require.Equal(t, 1, calledAddExecutionResult)

}

func TestHeadersExecutor_ProcessBlock(t *testing.T) {
	t.Parallel()

	t.Run("pause/resume should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		cntWasPopCalled := uint32(0)
		args.BlocksQueue = &processMocks.BlocksQueueMock{
			PopCalled: func() (queue.HeaderBodyPair, bool, bool) {
				atomic.AddUint32(&cntWasPopCalled, 1)

				return queue.HeaderBodyPair{
					Header: &block.Header{
						Nonce: 1,
					},
					Body: &block.Body{},
				}, true, true
			},
		}

		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{}, nil
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		executor.StartExecution()

		// allow some Pop operations
		time.Sleep(time.Millisecond * 200)

		executor.PauseExecution()
		time.Sleep(time.Millisecond * 20) // allow current processing to finish
		cntWasPopCalledAtPause := atomic.LoadUint32(&cntWasPopCalled)

		executor.PauseExecution() // coverage, already paused

		// wait a bit more
		time.Sleep(time.Millisecond * 200)

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
				ProcessBlockProposalCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
					return &block.BaseExecutionResult{}, nil
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
		blocksQueue := queue.NewBlocksQueue()
		args.BlocksQueue = blocksQueue
		wasAddExecutionResultCalled := false
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) error {
				wasAddExecutionResultCalled = true
				return errExpected
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		err = blocksQueue.AddOrReplace(queue.HeaderBodyPair{
			Header: &block.Header{
				Nonce: 1,
			},
			Body: &block.Body{},
		})
		require.NoError(t, err)

		executor.StartExecution()

		// allow Pop operation
		time.Sleep(time.Millisecond * 100)

		err = executor.Close()
		require.NoError(t, err)
		require.True(t, wasAddExecutionResultCalled)
	})

	t.Run("block processing error, after retry should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		blocksQueue := queue.NewBlocksQueue()
		count := 0
		countAddResult := 0
		args.BlocksQueue = blocksQueue
		wg := &sync.WaitGroup{}
		wg.Add(1)
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				if count == 1 {
					return &block.BaseExecutionResult{}, nil
				}
				count++
				return nil, errExpected
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) error {
				countAddResult++
				wg.Done()
				return nil
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		executor.StartExecution()

		err = blocksQueue.AddOrReplace(queue.HeaderBodyPair{
			Header: &block.Header{
				Nonce: 1,
			},
			Body: &block.Body{},
		})
		require.NoError(t, err)

		wg.Wait()
		err = executor.Close()
		require.NoError(t, err)
		require.Equal(t, 1, countAddResult)
	})

	t.Run("block processing error, pop header for queue with the same nonce", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		blocksQueue := queue.NewBlocksQueue()

		count := 0
		countAddResult := 0
		args.BlocksQueue = blocksQueue
		wg := &sync.WaitGroup{}
		wg.Add(1)
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				time.Sleep(time.Millisecond)
				if handler.GetRound() == 1 {
					return nil, errExpected
				}

				count++
				return &block.BaseExecutionResult{}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) error {
				countAddResult++
				wg.Done()
				return nil
			},
		}

		executor, err := NewHeadersExecutor(args)
		require.NoError(t, err)

		executor.StartExecution()

		err = blocksQueue.AddOrReplace(queue.HeaderBodyPair{
			Header: &block.Header{
				Nonce: 1,
				Round: 1,
			},
			Body: &block.Body{},
		})
		require.NoError(t, err)

		time.Sleep(time.Millisecond)
		err = blocksQueue.AddOrReplace(queue.HeaderBodyPair{
			Header: &block.Header{
				Nonce: 1,
				Round: 2,
			},
			Body: &block.Body{},
		})
		require.NoError(t, err)

		wg.Wait()

		require.Equal(t, 1, count)
		_, ok := blocksQueue.Peek()
		// check if queue is empty
		require.False(t, ok)
	})
}

func TestHeadersExecutor_Process(t *testing.T) {
	t.Parallel()

	t.Run("should return error on failing to process block", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		expectedErr := errors.New("expected error")
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return nil, expectedErr
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := queue.HeaderBodyPair{
			Header: &block.Header{
				Nonce: 1,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return nil on failing to add execution results to execution tracker", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		countAddResult := 0

		expectedErr := errors.New("expected error")
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) error {
				countAddResult++
				return expectedErr
			},
		}

		executor, _ := NewHeadersExecutor(args)

		pair := queue.HeaderBodyPair{
			Header: &block.Header{
				Nonce: 1,
			},
			Body: &block.Body{},
		}

		err := executor.Process(pair)
		require.Nil(t, err)
	})

	t.Run("should add execution result info to blockchain handler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				return &block.BaseExecutionResult{}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) error {
				return nil
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
		}

		executor, _ := NewHeadersExecutor(args)

		pair := queue.HeaderBodyPair{
			Header: &block.Header{
				Nonce: 1,
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
