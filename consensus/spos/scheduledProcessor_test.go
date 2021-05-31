package spos

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func TestProcessingStatus_String(t *testing.T) {
	t.Parallel()

	require.Equal(t, processingNotStartedString, processingNotStarted.String())
	require.Equal(t, processingErrorString, processingError.String())
	require.Equal(t, inProgressString, inProgress.String())
	require.Equal(t, processingOKString, processingOK.String())
}

func TestNewScheduledProcessor_NilSyncTimerShouldErr(t *testing.T) {
	t.Parallel()

	args := ScheduledProcessorArgs{
		SyncTimer:                  nil,
		Processor:                  &mock.BlockProcessorMock{},
		ProcessingTimeMilliSeconds: 10,
	}

	sp, err := NewScheduledProcessor(args)
	require.Nil(t, sp)
	require.Equal(t, ErrNilSyncTimer, err)
}

func TestNewScheduledProcessor_NilBlockProcessorShouldErr(t *testing.T) {
	t.Parallel()

	args := ScheduledProcessorArgs{
		SyncTimer:                  &mock.SyncTimerMock{},
		Processor:                  nil,
		ProcessingTimeMilliSeconds: 10,
	}

	sp, err := NewScheduledProcessor(args)
	require.Nil(t, sp)
	require.Equal(t, process.ErrNilBlockProcessor, err)
}

func TestNewScheduledProcessor_NilBlockProcessorOK(t *testing.T) {
	t.Parallel()

	args := ScheduledProcessorArgs{
		SyncTimer:                  &mock.SyncTimerMock{},
		Processor:                  &mock.BlockProcessorMock{},
		ProcessingTimeMilliSeconds: 10,
	}

	sp, err := NewScheduledProcessor(args)
	require.Nil(t, err)
	require.NotNil(t, sp)
}

func TestScheduledProcessor_IsStatusOK(t *testing.T) {
	t.Parallel()

	args := ScheduledProcessorArgs{
		SyncTimer:                  &mock.SyncTimerMock{},
		Processor:                  &mock.BlockProcessorMock{},
		ProcessingTimeMilliSeconds: 10,
	}

	sp, err := NewScheduledProcessor(args)

	require.Nil(t, err)
	require.False(t, sp.IsProcessedOK())

	sp.setStatus(processingOK)
	require.True(t, sp.IsProcessedOK())

	sp.setStatus(processingError)
	require.False(t, sp.IsProcessedOK())

	sp.setStatus(inProgress)
	require.False(t, sp.IsProcessedOK())
}

func TestScheduledProcessor_StatusGetterAndSetter(t *testing.T) {
	t.Parallel()

	args := ScheduledProcessorArgs{
		SyncTimer:                  &mock.SyncTimerMock{},
		Processor:                  &mock.BlockProcessorMock{},
		ProcessingTimeMilliSeconds: 10,
	}

	sp, _ := NewScheduledProcessor(args)
	require.Equal(t, processingNotStarted, sp.getStatus())

	sp.setStatus(processingOK)
	require.Equal(t, processingOK, sp.getStatus())

	sp.setStatus(inProgress)
	require.Equal(t, inProgress, sp.getStatus())

	sp.setStatus(processingError)
	require.Equal(t, processingError, sp.getStatus())
}

func TestScheduledProcessor_StartScheduledProcessingHeaderV1ProcessingOK(t *testing.T) {
	t.Parallel()

	processScheduledCalled := false
	args := ScheduledProcessorArgs{
		SyncTimer: &mock.SyncTimerMock{},
		Processor: &mock.BlockProcessorMock{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled = true
				return nil
			},
		},
		ProcessingTimeMilliSeconds: 10,
	}

	sp, _ := NewScheduledProcessor(args)
	require.Equal(t, processingNotStarted, sp.getStatus())

	header := &block.Header{}
	body := &block.Body{}
	sp.StartScheduledProcessing(header, body)
	time.Sleep(10 * time.Millisecond)
	require.False(t, processScheduledCalled)
	require.Equal(t, processingOK, sp.getStatus())
}

func TestScheduledProcessor_StartScheduledProcessingHeaderV2ProcessingWithError(t *testing.T) {
	t.Parallel()

	processScheduledCalled := false
	args := ScheduledProcessorArgs{
		SyncTimer: &mock.SyncTimerMock{},
		Processor: &mock.BlockProcessorMock{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled = true
				return errors.New("processing error")
			},
		},
		ProcessingTimeMilliSeconds: 10,
	}

	sp, _ := NewScheduledProcessor(args)
	require.Equal(t, processingNotStarted, sp.getStatus())

	header := &block.HeaderV2{}
	body := &block.Body{}
	sp.StartScheduledProcessing(header, body)
	require.Equal(t, inProgress, sp.getStatus())

	time.Sleep(100 * time.Millisecond)
	require.True(t, processScheduledCalled)
	require.Equal(t, processingError, sp.getStatus())
}

func TestScheduledProcessor_StartScheduledProcessingHeaderV2ProcessingOK(t *testing.T) {
	t.Parallel()

	processScheduledCalled := false
	args := ScheduledProcessorArgs{
		SyncTimer: &mock.SyncTimerMock{},
		Processor: &mock.BlockProcessorMock{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled = true
				return nil
			},
		},
		ProcessingTimeMilliSeconds: 10,
	}

	sp, _ := NewScheduledProcessor(args)
	require.Equal(t, processingNotStarted, sp.getStatus())

	header := &block.HeaderV2{}
	body := &block.Body{}
	sp.StartScheduledProcessing(header, body)
	require.Equal(t, inProgress, sp.getStatus())

	time.Sleep(100 * time.Millisecond)
	require.True(t, processScheduledCalled)
	require.Equal(t, processingOK, sp.getStatus())
}
