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

func TestScheduledProcessor_IsProcessedOKEarlyExit(t *testing.T) {
	t.Parallel()

	called := false
	args := ScheduledProcessorArgs{
		SyncTimer: &mock.SyncTimerMock{
			CurrentTimeCalled: func() time.Time {
				called = true
				return time.Now()
			},
		},
		Processor:                  &mock.BlockProcessorMock{},
		ProcessingTimeMilliSeconds: 100,
	}

	sp, err := NewScheduledProcessor(args)
	require.Nil(t, err)

	require.False(t, sp.IsProcessedOK())
	require.False(t, called)

	sp.setStatus(processingOK)
	require.True(t, sp.IsProcessedOK())
	require.False(t, called)

	sp.setStatus(processingError)
	require.False(t, sp.IsProcessedOK())
	require.False(t, called)
}

func defaultScheduledProcessorArgs() ScheduledProcessorArgs {
	return ScheduledProcessorArgs{
		SyncTimer: &mock.SyncTimerMock{
			CurrentTimeCalled: func() time.Time {
				return time.Now()
			},
		},
		Processor:                  &mock.BlockProcessorMock{},
		ProcessingTimeMilliSeconds: 100,
	}
}

func TestScheduledProcessor_IsProcessedInProgressNegativeRemainingTime(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorArgs()
	sp, err := NewScheduledProcessor(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	require.False(t, sp.IsProcessedOK())
}

func TestScheduledProcessor_IsProcessedInProgressStartingInFuture(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorArgs()
	sp, err := NewScheduledProcessor(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	startTime := time.Now()
	sp.startTime = startTime.Add(10 * time.Millisecond)
	require.False(t, sp.IsProcessedOK())
	endTime := time.Now()
	require.Less(t, endTime.Sub(startTime), time.Millisecond)
}

func TestScheduledProcessor_IsProcessedInProgressEarlyCompletion(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorArgs()
	sp, err := NewScheduledProcessor(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	sp.startTime = time.Now()
	go func() {
		time.Sleep(10*time.Millisecond)
		sp.setStatus(processingOK)
	}()
	require.True(t, sp.IsProcessedOK())
	endTime := time.Now()
	timeSpent := endTime.Sub(sp.startTime)
	require.Less(t, timeSpent, sp.processingTime)
}

func TestScheduledProcessor_IsProcessedInProgressEarlyCompletionWithError(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorArgs()
	sp, err := NewScheduledProcessor(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	sp.startTime = time.Now()
	go func() {
		time.Sleep(10*time.Millisecond)
		sp.setStatus(processingError)
	}()
	require.False(t, sp.IsProcessedOK())
	endTime := time.Now()
	timeSpent := endTime.Sub(sp.startTime)
	require.Less(t, timeSpent, sp.processingTime)
}

func TestScheduledProcessor_IsProcessedInProgressAlreadyStartedNoCompletion(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorArgs()
	sp, err := NewScheduledProcessor(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	startTime := time.Now()
	sp.startTime = startTime.Add(-10 * time.Millisecond)
	require.False(t, sp.IsProcessedOK())
	endTime := time.Now()
	require.Less(t, endTime.Sub(startTime), sp.processingTime)
	require.Greater(t, endTime.Sub(startTime), sp.processingTime-10*time.Millisecond)
}

func TestScheduledProcessor_IsProcessedInProgressTimeout(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorArgs()
	sp, err := NewScheduledProcessor(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	sp.startTime = time.Now()
	require.False(t, sp.IsProcessedOK())
	endTime := time.Now()
	require.Greater(t, endTime.Sub(sp.startTime), sp.processingTime)
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
