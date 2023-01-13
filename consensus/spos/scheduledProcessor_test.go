package spos

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestProcessingStatus_String(t *testing.T) {
	t.Parallel()

	require.Equal(t, processingNotStartedString, processingNotStarted.String())
	require.Equal(t, processingErrorString, processingError.String())
	require.Equal(t, inProgressString, inProgress.String())
	require.Equal(t, processingOKString, processingOK.String())
	require.Equal(t, stoppedString, stopped.String())
}

func TestNewScheduledProcessorWrapper_NilSyncTimerShouldErr(t *testing.T) {
	t.Parallel()

	args := ScheduledProcessorWrapperArgs{
		SyncTimer:                nil,
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}

	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, sp)
	require.Equal(t, ErrNilSyncTimer, err)
}

func TestNewScheduledProcessorWrapper_NilBlockProcessorShouldErr(t *testing.T) {
	t.Parallel()

	args := ScheduledProcessorWrapperArgs{
		SyncTimer:                &mock.SyncTimerMock{},
		Processor:                nil,
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}

	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, sp)
	require.Equal(t, process.ErrNilBlockProcessor, err)
}

func TestNewScheduledProcessorWrapper_NilRoundTimeDurationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := ScheduledProcessorWrapperArgs{
		SyncTimer:                &mock.SyncTimerMock{},
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: nil,
	}

	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, sp)
	require.Equal(t, process.ErrNilRoundTimeDurationHandler, err)
}

func TestNewScheduledProcessorWrapper_NilBlockProcessorOK(t *testing.T) {
	t.Parallel()

	args := ScheduledProcessorWrapperArgs{
		SyncTimer:                &mock.SyncTimerMock{},
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}

	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, err)
	require.NotNil(t, sp)
}

func TestScheduledProcessorWrapper_IsProcessedOKEarlyExit(t *testing.T) {
	t.Parallel()

	called := atomic.Flag{}
	args := ScheduledProcessorWrapperArgs{
		SyncTimer: &mock.SyncTimerMock{
			CurrentTimeCalled: func() time.Time {
				called.SetValue(true)
				return time.Now()
			},
		},
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}

	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	require.False(t, sp.IsProcessedOKWithTimeout())
	require.False(t, called.IsSet())

	sp.setStatus(processingOK)
	require.True(t, sp.IsProcessedOKWithTimeout())
	require.False(t, called.IsSet())

	sp.setStatus(processingError)
	require.False(t, sp.IsProcessedOKWithTimeout())
	require.False(t, called.IsSet())
}

func defaultScheduledProcessorWrapperArgs() ScheduledProcessorWrapperArgs {
	return ScheduledProcessorWrapperArgs{
		SyncTimer: &mock.SyncTimerMock{
			CurrentTimeCalled: func() time.Time {
				return time.Now()
			},
		},
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}
}

func TestScheduledProcessorWrapper_IsProcessedInProgressNegativeRemainingTime(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	require.False(t, sp.IsProcessedOKWithTimeout())

	startTime := time.Now()
	sp.startTime = startTime.Add(-200 * time.Millisecond)
	require.False(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	timeSpent := endTime.Sub(startTime)
	require.Less(t, timeSpent, sp.roundTimeDurationHandler.TimeDuration())
}

func TestScheduledProcessorWrapper_IsProcessedInProgressStartingInFuture(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	startTime := time.Now()
	sp.startTime = startTime.Add(500 * time.Millisecond)
	require.False(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	require.Less(t, endTime.Sub(startTime), time.Millisecond*100)
}

func TestScheduledProcessorWrapper_IsProcessedInProgressEarlyCompletion(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	sp.startTime = time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		sp.setStatus(processingOK)
	}()
	require.True(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	timeSpent := endTime.Sub(sp.startTime)
	require.Less(t, timeSpent, sp.roundTimeDurationHandler.TimeDuration())
}

func TestScheduledProcessorWrapper_IsProcessedInProgressEarlyCompletionWithError(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	sp.startTime = time.Now()
	go func() {
		time.Sleep(10 * time.Millisecond)
		sp.setStatus(processingError)
	}()
	require.False(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	timeSpent := endTime.Sub(sp.startTime)
	require.Less(t, timeSpent, sp.roundTimeDurationHandler.TimeDuration())
}

func TestScheduledProcessorWrapper_IsProcessedInProgressAlreadyStartedNoCompletion(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	startTime := time.Now()
	sp.startTime = startTime.Add(-10 * time.Millisecond)
	require.False(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	require.Less(t, endTime.Sub(startTime), sp.roundTimeDurationHandler.TimeDuration())
	require.Greater(t, endTime.Sub(startTime), sp.roundTimeDurationHandler.TimeDuration()-10*time.Millisecond)
}

func TestScheduledProcessorWrapper_IsProcessedInProgressTimeout(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.setStatus(inProgress)
	sp.startTime = time.Now()
	require.False(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	require.Greater(t, endTime.Sub(sp.startTime), sp.roundTimeDurationHandler.TimeDuration())
}

func TestScheduledProcessorWrapper_StatusGetterAndSetter(t *testing.T) {
	t.Parallel()

	args := ScheduledProcessorWrapperArgs{
		SyncTimer:                &mock.SyncTimerMock{},
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}

	sp, _ := NewScheduledProcessorWrapper(args)
	require.Equal(t, processingNotStarted, sp.getStatus())

	sp.setStatus(processingOK)
	require.Equal(t, processingOK, sp.getStatus())

	sp.setStatus(inProgress)
	require.Equal(t, inProgress, sp.getStatus())

	sp.setStatus(processingError)
	require.Equal(t, processingError, sp.getStatus())
}

func TestScheduledProcessorWrapper_StartScheduledProcessingHeaderV1ProcessingOK(t *testing.T) {
	t.Parallel()

	processScheduledCalled := atomic.Flag{}
	args := ScheduledProcessorWrapperArgs{
		SyncTimer: &mock.SyncTimerMock{},
		Processor: &testscommon.BlockProcessorStub{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled.SetValue(true)
				return nil
			},
		},
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}

	sp, _ := NewScheduledProcessorWrapper(args)
	require.Equal(t, processingNotStarted, sp.getStatus())

	header := &block.Header{}
	body := &block.Body{}
	sp.StartScheduledProcessing(header, body, time.Now())
	time.Sleep(10 * time.Millisecond)
	require.False(t, processScheduledCalled.IsSet())
	require.Equal(t, processingOK, sp.getStatus())
}

func TestScheduledProcessorWrapper_StartScheduledProcessingHeaderV2ProcessingWithError(t *testing.T) {
	t.Parallel()

	processScheduledCalled := atomic.Flag{}
	args := ScheduledProcessorWrapperArgs{
		SyncTimer: &mock.SyncTimerMock{},
		Processor: &testscommon.BlockProcessorStub{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled.SetValue(true)
				return errors.New("processing error")
			},
		},
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}

	sp, _ := NewScheduledProcessorWrapper(args)
	require.Equal(t, processingNotStarted, sp.getStatus())

	header := &block.HeaderV2{}
	body := &block.Body{}
	sp.StartScheduledProcessing(header, body, time.Now())
	require.Equal(t, inProgress, sp.getStatus())

	time.Sleep(100 * time.Millisecond)
	require.True(t, processScheduledCalled.IsSet())
	require.Equal(t, processingError, sp.getStatus())
}

func TestScheduledProcessorWrapper_StartScheduledProcessingHeaderV2ProcessingOK(t *testing.T) {
	t.Parallel()

	processScheduledCalled := atomic.Flag{}
	args := ScheduledProcessorWrapperArgs{
		SyncTimer: &mock.SyncTimerMock{},
		Processor: &testscommon.BlockProcessorStub{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled.SetValue(true)
				return nil
			},
		},
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}

	sp, _ := NewScheduledProcessorWrapper(args)
	require.Equal(t, processingNotStarted, sp.getStatus())

	header := &block.HeaderV2{}
	body := &block.Body{}
	sp.StartScheduledProcessing(header, body, time.Now())
	require.Equal(t, inProgress, sp.getStatus())

	time.Sleep(100 * time.Millisecond)
	require.True(t, processScheduledCalled.IsSet())
	require.Equal(t, processingOK, sp.getStatus())
}

func TestScheduledProcessorWrapper_StartScheduledProcessingHeaderV2ForceStopped(t *testing.T) {
	t.Parallel()

	processScheduledCalled := atomic.Flag{}

	args := ScheduledProcessorWrapperArgs{
		SyncTimer: &mock.SyncTimerMock{
			CurrentTimeCalled: func() time.Time {
				return time.Now()
			},
		},
		Processor: &testscommon.BlockProcessorStub{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled.SetValue(true)
				for {
					<-time.After(time.Millisecond)
					remainingTime := haveTime()
					if remainingTime == 0 {
						return errors.New("timeout")
					}
				}
			},
		},
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}

	spw, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	hdr := &block.HeaderV2{}
	blkBody := &block.Body{}
	spw.StartScheduledProcessing(hdr, blkBody, time.Now())
	time.Sleep(time.Second)
	startTime := time.Now()
	spw.ForceStopScheduledExecutionBlocking()
	endTime := time.Now()
	status := spw.getStatus()
	require.True(t, processScheduledCalled.IsSet())
	require.Equal(t, stopped, status, status.String())
	require.Less(t, 10*time.Millisecond, endTime.Sub(startTime))
}

func TestScheduledProcessorWrapper_StartScheduledProcessingHeaderV2ForceStopAfterProcessingEnded(t *testing.T) {
	t.Parallel()

	processScheduledCalled := atomic.Flag{}
	args := ScheduledProcessorWrapperArgs{
		SyncTimer: &mock.SyncTimerMock{
			CurrentTimeCalled: func() time.Time {
				return time.Now()
			},
		},
		Processor: &testscommon.BlockProcessorStub{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled.SetValue(true)
				<-time.After(time.Millisecond)
				return nil
			},
		},
		RoundTimeDurationHandler: &mock.RoundHandlerMock{},
	}

	spw, err := NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	hdr := &block.HeaderV2{}
	blkBody := &block.Body{}
	spw.StartScheduledProcessing(hdr, blkBody, time.Now())
	time.Sleep(200 * time.Millisecond)
	spw.ForceStopScheduledExecutionBlocking()
	status := spw.getStatus()
	require.True(t, processScheduledCalled.IsSet())
	require.Equal(t, processingOK, status, status.String())
}
