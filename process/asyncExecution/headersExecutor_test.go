package asyncExecution

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	"github.com/stretchr/testify/require"
)

func createMockArgs() ArgsHeadersExecutor {
	headerQueue, _ := queue.NewBlocksQueue()

	return ArgsHeadersExecutor{
		BlocksQueue:      headerQueue,
		ExecutionTracker: &processMocks.ExecutionTrackerStub{},
		BlockProcessor:   &processMocks.BlockProcessorStub{},
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
	blocksQueue, _ := queue.NewBlocksQueue()
	args.BlocksQueue = blocksQueue
	args.BlockProcessor = &processMocks.BlockProcessorStub{
		ProcessBlockCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.ExecutionResultHandler, error) {
			calledProcessBlock++
			return nil, nil
		},
	}
	args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
		AddExecutionResultCalled: func(executionResult data.ExecutionResultHandler) error {
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

	t.Run("block processing error, after retry should work", func(t *testing.T) {
		args := createMockArgs()
		blocksQueue, _ := queue.NewBlocksQueue()
		count := 0
		countAddResult := 0
		args.BlocksQueue = blocksQueue
		wg := &sync.WaitGroup{}
		wg.Add(1)
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.ExecutionResultHandler, error) {
				if count == 1 {
					return nil, nil
				}
				count++
				return nil, errors.New("local error")
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.ExecutionResultHandler) error {
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
		args := createMockArgs()
		blocksQueue, _ := queue.NewBlocksQueue()

		count := 0
		countAddResult := 0
		args.BlocksQueue = blocksQueue
		wg := &sync.WaitGroup{}
		wg.Add(1)
		args.BlockProcessor = &processMocks.BlockProcessorStub{
			ProcessBlockCalled: func(handler data.HeaderHandler, body data.BodyHandler) (data.ExecutionResultHandler, error) {
				time.Sleep(time.Millisecond)
				if handler.GetRound() == 1 {
					return nil, errors.New("local error")
				}

				count++
				return nil, nil
			},
		}
		args.ExecutionTracker = &processMocks.ExecutionTrackerStub{
			AddExecutionResultCalled: func(executionResult data.ExecutionResultHandler) error {
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
		_, ok := blocksQueue.Peak()
		// check if queue is empty
		require.False(t, ok)
	})
}
