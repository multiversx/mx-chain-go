package asyncExecution

import (
	"errors"
	"sync"
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

func TestHeadersExecutor_ProcessBlockError(t *testing.T) {
	t.Parallel()

	t.Run("header marked for deletion before processing should be skipped", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		blocksQueue := queue.NewBlocksQueue()
		args.BlocksQueue = blocksQueue
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				require.Fail(t, "should not be called")
				return nil, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) error {
				require.Fail(t, "should not be called")
				return nil
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

		// mark the popped header as evicted
		executor.OnHeaderEvicted(1)

		executor.StartExecution()

		// allow Pop operation
		time.Sleep(time.Millisecond * 100)

		err = executor.Close()
		require.NoError(t, err)
	})

	t.Run("header marked for deletion during processing should be skipped", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		blocksQueue := queue.NewBlocksQueue()
		args.BlocksQueue = blocksQueue
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockProposalCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
				// this should trigger the notification
				blocksQueue.RemoveAtNonceAndHigher(1)

				return &block.BaseExecutionResult{}, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.BaseExecutionResultHandler) error {
				require.Fail(t, "should not be called")
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

		// allow Pop operation
		time.Sleep(time.Millisecond * 100)

		err = executor.Close()
		require.NoError(t, err)
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

	t.Run("should add execution result to blockchain handler", func(t *testing.T) {
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
		args.BlockChain = &testscommon.ChainHandlerStub{
			SetFinalBlockInfoCalled: func(nonce uint64, headerHash, rootHash []byte) {
				setFinalBlockInfoCalled = true
			},
			SetLastExecutedBlockInfoCalled: func(nonce uint64, headerHash, rootHash []byte) {
				setLastExecutedBlockInfoCalled = true
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
	})
}
