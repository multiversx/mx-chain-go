package asyncExecution

import (
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
