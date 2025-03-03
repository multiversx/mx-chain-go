package spos_test

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"

	"github.com/stretchr/testify/require"
)

func TestProcessingStatus_String(t *testing.T) {
	t.Parallel()

	require.Equal(t, spos.ProcessingNotStartedString, spos.ProcessingNotStarted.String())
	require.Equal(t, spos.ProcessingErrorString, spos.ProcessingError.String())
	require.Equal(t, spos.InProgressString, spos.InProgress.String())
	require.Equal(t, spos.ProcessingOKString, spos.ProcessingOK.String())
	require.Equal(t, spos.StoppedString, spos.Stopped.String())
}

func TestNewScheduledProcessorWrapper_NilSyncTimerShouldErr(t *testing.T) {
	t.Parallel()

	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                nil,
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}

	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, sp)
	require.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestNewScheduledProcessorWrapper_NilBlockProcessorShouldErr(t *testing.T) {
	t.Parallel()

	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                &consensus.SyncTimerMock{},
		Processor:                nil,
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}

	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, sp)
	require.Equal(t, process.ErrNilBlockProcessor, err)
}

func TestNewScheduledProcessorWrapper_NilRoundTimeDurationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                &consensus.SyncTimerMock{},
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: nil,
	}

	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, sp)
	require.Equal(t, process.ErrNilRoundTimeDurationHandler, err)
}

func TestNewScheduledProcessorWrapper_NilBlockProcessorOK(t *testing.T) {
	t.Parallel()

	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                &consensus.SyncTimerMock{},
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}

	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, err)
	require.NotNil(t, sp)
}

func TestScheduledProcessorWrapper_IsProcessedOKEarlyExit(t *testing.T) {
	t.Parallel()

	called := atomic.Flag{}
	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer: &consensus.SyncTimerMock{
			CurrentTimeCalled: func() time.Time {
				called.SetValue(true)
				return time.Now()
			},
		},
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}

	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	require.False(t, sp.IsProcessedOKWithTimeout())
	require.False(t, called.IsSet())

	sp.SetStatus(spos.ProcessingOK)
	require.True(t, sp.IsProcessedOKWithTimeout())
	require.False(t, called.IsSet())

	sp.SetStatus(spos.ProcessingError)
	require.False(t, sp.IsProcessedOKWithTimeout())
	require.False(t, called.IsSet())
}

func defaultScheduledProcessorWrapperArgs() spos.ScheduledProcessorWrapperArgs {
	return spos.ScheduledProcessorWrapperArgs{
		SyncTimer: &consensus.SyncTimerMock{
			CurrentTimeCalled: func() time.Time {
				return time.Now()
			},
		},
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}
}

func TestScheduledProcessorWrapper_IsProcessedInProgressNegativeRemainingTime(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.SetStatus(spos.InProgress)
	require.False(t, sp.IsProcessedOKWithTimeout())

	startTime := time.Now()
	sp.SetStartTime(startTime.Add(-200 * time.Millisecond))
	require.False(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	timeSpent := endTime.Sub(startTime)
	require.Less(t, timeSpent, sp.GetRoundTimeHandler().TimeDuration())
}

func TestScheduledProcessorWrapper_IsProcessedInProgressStartingInFuture(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.SetStatus(spos.InProgress)
	startTime := time.Now()
	sp.SetStartTime(startTime.Add(500 * time.Millisecond))
	require.False(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	require.Less(t, endTime.Sub(startTime), time.Millisecond*100)
}

func TestScheduledProcessorWrapper_IsProcessedInProgressEarlyCompletion(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.SetStatus(spos.InProgress)
	sp.SetStartTime(time.Now())
	go func() {
		time.Sleep(10 * time.Millisecond)
		sp.SetStatus(spos.ProcessingOK)
	}()
	require.True(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	timeSpent := endTime.Sub(sp.GetStartTime())
	require.Less(t, timeSpent, sp.GetRoundTimeHandler().TimeDuration())
}

func TestScheduledProcessorWrapper_IsProcessedInProgressEarlyCompletionWithError(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.SetStatus(spos.InProgress)
	sp.SetStartTime(time.Now())
	go func() {
		time.Sleep(10 * time.Millisecond)
		sp.SetStatus(spos.ProcessingError)
	}()
	require.False(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	timeSpent := endTime.Sub(sp.GetStartTime())
	require.Less(t, timeSpent, sp.GetRoundTimeHandler().TimeDuration())
}

func TestScheduledProcessorWrapper_IsProcessedInProgressAlreadyStartedNoCompletion(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.SetStatus(spos.InProgress)
	startTime := time.Now()
	sp.SetStartTime(startTime.Add(-10 * time.Millisecond))
	require.False(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	require.Less(t, endTime.Sub(startTime), sp.GetRoundTimeHandler().TimeDuration())
	require.Greater(t, endTime.Sub(startTime), sp.GetRoundTimeHandler().TimeDuration()-10*time.Millisecond)
}

func TestScheduledProcessorWrapper_IsProcessedInProgressTimeout(t *testing.T) {
	t.Parallel()

	args := defaultScheduledProcessorWrapperArgs()
	sp, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	sp.SetStatus(spos.InProgress)
	sp.SetStartTime(time.Now())
	require.False(t, sp.IsProcessedOKWithTimeout())
	endTime := time.Now()
	require.Greater(t, endTime.Sub(sp.GetStartTime()), sp.GetRoundTimeHandler().TimeDuration())
}

func TestScheduledProcessorWrapper_StatusGetterAndSetter(t *testing.T) {
	t.Parallel()

	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer:                &consensus.SyncTimerMock{},
		Processor:                &testscommon.BlockProcessorStub{},
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}

	sp, _ := spos.NewScheduledProcessorWrapper(args)
	require.Equal(t, spos.ProcessingNotStarted, sp.GetStatus())

	sp.SetStatus(spos.ProcessingOK)
	require.Equal(t, spos.ProcessingOK, sp.GetStatus())

	sp.SetStatus(spos.InProgress)
	require.Equal(t, spos.InProgress, sp.GetStatus())

	sp.SetStatus(spos.ProcessingError)
	require.Equal(t, spos.ProcessingError, sp.GetStatus())
}

func TestScheduledProcessorWrapper_StartScheduledProcessingHeaderV1ProcessingOK(t *testing.T) {
	t.Parallel()

	processScheduledCalled := atomic.Flag{}
	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer: &consensus.SyncTimerMock{},
		Processor: &testscommon.BlockProcessorStub{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled.SetValue(true)
				return nil
			},
		},
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}

	sp, _ := spos.NewScheduledProcessorWrapper(args)
	require.Equal(t, spos.ProcessingNotStarted, sp.GetStatus())

	header := &block.Header{}
	body := &block.Body{}
	sp.StartScheduledProcessing(header, body, time.Now())
	time.Sleep(10 * time.Millisecond)
	require.False(t, processScheduledCalled.IsSet())
	require.Equal(t, spos.ProcessingOK, sp.GetStatus())
}

func TestScheduledProcessorWrapper_StartScheduledProcessingHeaderV2ProcessingWithError(t *testing.T) {
	t.Parallel()

	processScheduledCalled := atomic.Flag{}
	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer: &consensus.SyncTimerMock{},
		Processor: &testscommon.BlockProcessorStub{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled.SetValue(true)
				return errors.New("processing error")
			},
		},
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}

	sp, _ := spos.NewScheduledProcessorWrapper(args)
	require.Equal(t, spos.ProcessingNotStarted, sp.GetStatus())

	header := &block.HeaderV2{}
	body := &block.Body{}
	sp.StartScheduledProcessing(header, body, time.Now())
	require.Equal(t, spos.InProgress, sp.GetStatus())

	time.Sleep(100 * time.Millisecond)
	require.True(t, processScheduledCalled.IsSet())
	require.Equal(t, spos.ProcessingError, sp.GetStatus())
}

func TestScheduledProcessorWrapper_StartScheduledProcessingHeaderV2ProcessingOK(t *testing.T) {
	t.Parallel()

	processScheduledCalled := atomic.Flag{}
	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer: &consensus.SyncTimerMock{},
		Processor: &testscommon.BlockProcessorStub{
			ProcessScheduledBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				processScheduledCalled.SetValue(true)
				return nil
			},
		},
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}

	sp, _ := spos.NewScheduledProcessorWrapper(args)
	require.Equal(t, spos.ProcessingNotStarted, sp.GetStatus())

	header := &block.HeaderV2{}
	body := &block.Body{}
	sp.StartScheduledProcessing(header, body, time.Now())
	require.Equal(t, spos.InProgress, sp.GetStatus())

	time.Sleep(100 * time.Millisecond)
	require.True(t, processScheduledCalled.IsSet())
	require.Equal(t, spos.ProcessingOK, sp.GetStatus())
}

func TestScheduledProcessorWrapper_StartScheduledProcessingHeaderV2ForceStopped(t *testing.T) {
	t.Parallel()

	processScheduledCalled := atomic.Flag{}

	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer: &consensus.SyncTimerMock{
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
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}

	spw, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	hdr := &block.HeaderV2{}
	blkBody := &block.Body{}
	spw.StartScheduledProcessing(hdr, blkBody, time.Now())
	time.Sleep(time.Second)
	startTime := time.Now()
	spw.ForceStopScheduledExecutionBlocking()
	endTime := time.Now()
	status := spw.GetStatus()
	require.True(t, processScheduledCalled.IsSet())
	require.Equal(t, spos.Stopped, status, status.String())
	require.Less(t, 10*time.Millisecond, endTime.Sub(startTime))
}

func TestScheduledProcessorWrapper_StartScheduledProcessingHeaderV2ForceStopAfterProcessingEnded(t *testing.T) {
	t.Parallel()

	processScheduledCalled := atomic.Flag{}
	args := spos.ScheduledProcessorWrapperArgs{
		SyncTimer: &consensus.SyncTimerMock{
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
		RoundTimeDurationHandler: &consensus.RoundHandlerMock{},
	}

	spw, err := spos.NewScheduledProcessorWrapper(args)
	require.Nil(t, err)

	hdr := &block.HeaderV2{}
	blkBody := &block.Body{}
	spw.StartScheduledProcessing(hdr, blkBody, time.Now())
	time.Sleep(200 * time.Millisecond)
	spw.ForceStopScheduledExecutionBlocking()
	status := spw.GetStatus()
	require.True(t, processScheduledCalled.IsSet())
	require.Equal(t, spos.ProcessingOK, status, status.String())
}
