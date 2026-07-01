package chronology_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/chronology"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/round"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func initSubroundHandlerMock() *mock.SubroundHandlerMock {
	srm := &mock.SubroundHandlerMock{}
	srm.CurrentCalled = func() int {
		return 0
	}
	srm.NextCalled = func() int {
		return 1
	}
	srm.DoWorkCalled = func(roundHandler consensus.RoundHandler) bool {
		return false
	}
	srm.NameCalled = func() string {
		return "(TEST)"
	}
	return srm
}

func TestChronology_NewChronologyNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	arg.RoundHandler = nil
	chr, err := chronology.NewChronology(arg)

	assert.Nil(t, chr)
	assert.Equal(t, err, chronology.ErrNilRoundHandler)
}

func TestChronology_NewChronologyNilSyncerShouldFail(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	arg.SyncTimer = nil
	chr, err := chronology.NewChronology(arg)

	assert.Nil(t, chr)
	assert.Equal(t, err, chronology.ErrNilSyncTimer)
}

func TestChronology_NewChronologyNilWatchdogShouldFail(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	arg.Watchdog = nil
	chr, err := chronology.NewChronology(arg)

	assert.Nil(t, chr)
	assert.Equal(t, err, chronology.ErrNilWatchdog)
}

func TestChronology_NewChronologyNilAppStatusHandlerShouldFail(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	arg.AppStatusHandler = nil
	chr, err := chronology.NewChronology(arg)

	assert.Nil(t, chr)
	assert.Equal(t, err, chronology.ErrNilAppStatusHandler)
}

func TestChronology_NewChronologyNilEnableEpochsHandlerShouldFail(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	arg.EnableEpochsHandler = nil
	chr, err := chronology.NewChronology(arg)

	assert.Nil(t, chr)
	assert.Equal(t, err, errors.ErrNilEnableEpochsHandler)
}

func TestChronology_NewChronologyNilEnableRoundsHandlerShouldFail(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	arg.EnableRoundsHandler = nil
	chr, err := chronology.NewChronology(arg)

	assert.Nil(t, chr)
	assert.Equal(t, err, errors.ErrNilEnableRoundsHandler)
}

func TestChronology_NewChronologyNilConfigsHandlerShouldFail(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	arg.ConfigsHandler = nil
	chr, err := chronology.NewChronology(arg)

	assert.Nil(t, chr)
	assert.Equal(t, err, common.ErrNilCommonConfigsHandler)
}

func TestChronology_NewChronologyShouldWork(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	chr, err := chronology.NewChronology(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(chr))
}

func TestChronology_AddSubroundShouldWork(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	chr, _ := chronology.NewChronology(arg)

	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())

	assert.Equal(t, 3, len(chr.SubroundHandlers()))
}

func TestChronology_RemoveAllSubroundsShouldReturnEmptySubroundHandlersArray(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	chr, _ := chronology.NewChronology(arg)

	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())
	chr.AddSubround(initSubroundHandlerMock())

	assert.Equal(t, 3, len(chr.SubroundHandlers()))
	chr.RemoveAllSubrounds()
	assert.Equal(t, 0, len(chr.SubroundHandlers()))
}

func TestChronology_StartRoundShouldReturnWhenRoundIndexIsNegative(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	roundHandlerMock := &round.RoundHandlerMock{}
	roundHandlerMock.IndexCalled = func() int64 {
		return -1
	}
	roundHandlerMock.BeforeGenesisCalled = func() bool {
		return true
	}
	arg.RoundHandler = roundHandlerMock
	chr, _ := chronology.NewChronology(arg)

	srm := initSubroundHandlerMock()
	chr.AddSubround(srm)
	chr.SetSubroundId(0)
	chr.StartRound()

	assert.Equal(t, srm.Current(), chr.SubroundId())
}

func TestChronology_StartRoundShouldReturnWhenLoadSubroundHandlerReturnsNil(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	chr, _ := chronology.NewChronology(arg)

	initSubroundHandlerMock()
	chr.StartRound()

	assert.Equal(t, -1, chr.SubroundId())
}

func TestChronology_StartRoundShouldReturnWhenDoWorkReturnsFalse(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	roundHandlerMock := &round.RoundHandlerMock{}
	roundHandlerMock.UpdateRound(roundHandlerMock.TimeStamp(), roundHandlerMock.TimeStamp().Add(roundHandlerMock.TimeDuration()))
	arg.RoundHandler = roundHandlerMock
	chr, _ := chronology.NewChronology(arg)

	srm := initSubroundHandlerMock()
	chr.AddSubround(srm)
	chr.SetSubroundId(0)
	chr.StartRound()

	assert.Equal(t, -1, chr.SubroundId())
}

func TestChronology_StartRoundShouldWork(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	roundHandlerMock := &round.RoundHandlerMock{}
	roundHandlerMock.UpdateRound(roundHandlerMock.TimeStamp(), roundHandlerMock.TimeStamp().Add(roundHandlerMock.TimeDuration()))
	arg.RoundHandler = roundHandlerMock
	chr, _ := chronology.NewChronology(arg)

	srm := initSubroundHandlerMock()
	srm.DoWorkCalled = func(roundHandler consensus.RoundHandler) bool {
		return true
	}
	chr.AddSubround(srm)
	chr.SetSubroundId(0)
	chr.StartRound()

	assert.Equal(t, srm.Next(), chr.SubroundId())
}

func TestChronology_UpdateRoundShouldInitRound(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	arg.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
		GetActivationRoundCalled: func(flag common.EnableRoundFlag) uint64 {
			return 2
		},
	}
	arg.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.SupernovaFlag
		},
	}
	chr, _ := chronology.NewChronology(arg)

	srm := initSubroundHandlerMock()
	wasSetBaseDurationCalled := false
	srm.SetBaseDurationCalled = func(baseDuration time.Duration) {
		wasSetBaseDurationCalled = true
	}
	chr.AddSubround(srm)

	// first call, supernova not yet active
	chr.UpdateRound()
	require.False(t, wasSetBaseDurationCalled)

	// second call, supernova activation round
	chr.UpdateRound()
	assert.Equal(t, srm.Current(), chr.SubroundId())
	require.True(t, wasSetBaseDurationCalled)

	// third call, coverage only after supernova
	chr.UpdateRound()
}

func TestChronology_LoadSubroundHandlerShouldReturnNilWhenSubroundHandlerNotExists(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	chr, _ := chronology.NewChronology(arg)

	assert.Nil(t, chr.LoadSubroundHandler(0))
}

func TestChronology_LoadSubroundHandlerShouldReturnNilWhenIndexIsOutOfBound(t *testing.T) {
	t.Parallel()
	arg := getDefaultChronologyArg()
	chr, _ := chronology.NewChronology(arg)

	chr.AddSubround(initSubroundHandlerMock())
	chr.SetSubroundHandlers(make([]consensus.SubroundHandler, 0))

	assert.Nil(t, chr.LoadSubroundHandler(0))
}

func TestChronology_InitRoundShouldNotSetSubroundWhenRoundIndexIsNegative(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	roundHandlerMock := &round.RoundHandlerMock{}
	arg.RoundHandler = roundHandlerMock
	arg.GenesisTime = arg.SyncTimer.CurrentTime()
	chr, _ := chronology.NewChronology(arg)

	chr.AddSubround(initSubroundHandlerMock())
	roundHandlerMock.IndexCalled = func() int64 {
		return -1
	}
	roundHandlerMock.BeforeGenesisCalled = func() bool {
		return true
	}
	chr.InitRound()

	assert.Equal(t, -1, chr.SubroundId())
}

func TestChronology_InitRoundShouldSetSubroundWhenRoundIndexIsPositive(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	roundHandlerMock := &round.RoundHandlerMock{}
	roundHandlerMock.UpdateRound(roundHandlerMock.TimeStamp(), roundHandlerMock.TimeStamp().Add(roundHandlerMock.TimeDuration()))
	arg.RoundHandler = roundHandlerMock
	arg.GenesisTime = arg.SyncTimer.CurrentTime()
	chr, _ := chronology.NewChronology(arg)

	sr := initSubroundHandlerMock()
	chr.AddSubround(sr)
	chr.InitRound()

	assert.Equal(t, sr.Current(), chr.SubroundId())
}

func TestChronology_StartRoundShouldNotUpdateRoundWhenCurrentRoundIsNotFinished(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	roundHandlerMock := &round.RoundHandlerMock{}
	arg.RoundHandler = roundHandlerMock
	arg.GenesisTime = arg.SyncTimer.CurrentTime()
	chr, _ := chronology.NewChronology(arg)

	chr.SetSubroundId(0)
	chr.StartRound()

	assert.Equal(t, int64(0), roundHandlerMock.Index())
}

func TestChronology_StartRoundShouldUpdateRoundWhenCurrentRoundIsFinished(t *testing.T) {
	t.Parallel()
	arg := getDefaultChronologyArg()
	roundHandlerMock := &round.RoundHandlerMock{}
	arg.RoundHandler = roundHandlerMock
	arg.GenesisTime = arg.SyncTimer.CurrentTime()
	chr, _ := chronology.NewChronology(arg)

	chr.SetSubroundId(-1)
	chr.StartRound()

	assert.Equal(t, int64(1), roundHandlerMock.Index())
}

func TestChronology_CheckIfStatusHandlerWorks(t *testing.T) {
	t.Parallel()

	chanDone := make(chan bool, 2)
	arg := getDefaultChronologyArg()
	arg.GenesisTime = arg.SyncTimer.CurrentTime()
	arg.AppStatusHandler = &statusHandlerMock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			chanDone <- true
		},
	}
	chr, err := chronology.NewChronology(arg)

	assert.Nil(t, err)

	srm := initSubroundHandlerMock()
	srm.DoWorkCalled = func(roundHandler consensus.RoundHandler) bool {
		return true
	}

	chr.AddSubround(srm)
	chr.StartRound()

	select {
	case <-chanDone:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "AppStatusHandler not working")
	}
}

func getDefaultChronologyArg() chronology.ArgChronology {
	return chronology.ArgChronology{
		GenesisTime:         time.Now(),
		RoundHandler:        &round.RoundHandlerMock{},
		SyncTimer:           &consensusMocks.SyncTimerMock{},
		AppStatusHandler:    statusHandlerMock.NewAppStatusHandlerMock(),
		Watchdog:            &mock.WatchdogMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EnableRoundsHandler: &testscommon.EnableRoundsHandlerStub{},
		ConfigsHandler:      testscommon.GetDefaultCommonConfigsHandler(),
	}
}

func TestChronology_CloseWatchDogStop(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	stopCalled := false
	arg.Watchdog = &mock.WatchdogMock{
		StopCalled: func(alarmID string) {
			stopCalled = true
		},
	}

	chr, err := chronology.NewChronology(arg)
	require.Nil(t, err)
	chr.SetCancelFunc(nil)

	err = chr.Close()
	assert.Nil(t, err)
	assert.True(t, stopCalled)
}

func TestChronology_Close(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	stopCalled := false
	arg.Watchdog = &mock.WatchdogMock{
		StopCalled: func(alarmID string) {
			stopCalled = true
		},
	}

	chr, err := chronology.NewChronology(arg)
	require.Nil(t, err)

	cancelCalled := false
	chr.SetCancelFunc(func() {
		cancelCalled = true
	})

	err = chr.Close()
	assert.Nil(t, err)
	assert.True(t, stopCalled)
	assert.True(t, cancelCalled)
}

func TestChronology_StartRounds(t *testing.T) {
	t.Parallel()

	t.Run("before supernova", func(t *testing.T) {
		t.Parallel()

		arg := getDefaultChronologyArg()

		chr, err := chronology.NewChronology(arg)
		require.Nil(t, err)
		doneFuncCalled := false

		ctx := &mock.ContextMock{
			DoneFunc: func() <-chan struct{} {
				done := make(chan struct{})
				close(done)
				doneFuncCalled = true
				return done
			},
		}
		chr.StartRoundsTest(ctx)
		assert.True(t, doneFuncCalled)
	})

	t.Run("with goroutine call, after supernova", func(t *testing.T) {
		t.Parallel()

		arg := getDefaultChronologyArg()
		arg.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.SupernovaFlag
			},
		}

		updateRoundCalled := &atomic.Bool{}
		updateRoundCalled.Store(false)

		arg.RoundHandler = &round.RoundHandlerMock{
			UpdateRoundCalled: func(t1, t2 time.Time) {
				updateRoundCalled.Store(true)
			},
		}

		chr, err := chronology.NewChronology(arg)
		require.Nil(t, err)

		chr.StartRounds()

		time.Sleep(5 * time.Millisecond)

		require.True(t, updateRoundCalled.Load())
	})
}

func TestChronology_StartRoundsShouldWork(t *testing.T) {
	t.Parallel()

	arg := getDefaultChronologyArg()
	roundHandlerMock := &round.RoundHandlerMock{}
	roundHandlerMock.UpdateRound(roundHandlerMock.TimeStamp(), roundHandlerMock.TimeStamp().Add(roundHandlerMock.TimeDuration()))
	arg.RoundHandler = roundHandlerMock
	chr, _ := chronology.NewChronology(arg)

	srm := initSubroundHandlerMock()
	srm.DoWorkCalled = func(roundHandler consensus.RoundHandler) bool {
		return true
	}
	chr.AddSubround(srm)
	chr.SetSubroundId(1)
	chr.StartRounds()
	defer chr.Close()

	assert.Equal(t, srm.Next(), chr.SubroundId())
	time.Sleep(time.Millisecond * 10)
}

func TestChronology_HandleRoundChangedIfNeeded(t *testing.T) {
	t.Parallel()

	timingBeforeBoundary := config.ConsensusConfigByRound{
		EnableRound: 0,
		SubroundsTiming: []config.SubroundTiming{
			{StartTime: 0.0, EndTime: 0.05},
			{StartTime: 0.05, EndTime: 0.25},
			{StartTime: 0.25, EndTime: 0.85},
			{StartTime: 0.85, EndTime: 0.95},
		},
		ProcessingThresholdPercent: 85,
	}
	timingAfterBoundary := config.ConsensusConfigByRound{
		EnableRound: 10,
		SubroundsTiming: []config.SubroundTiming{
			{StartTime: 0.0, EndTime: 0.10},
			{StartTime: 0.10, EndTime: 0.30},
			{StartTime: 0.30, EndTime: 0.80},
			{StartTime: 0.80, EndTime: 0.90},
		},
		ProcessingThresholdPercent: 90,
	}

	configsHandlerStub := &testscommon.CommonConfigsHandlerStub{
		GetActiveTimingBoundaryRoundCalled: func(round uint64) uint64 {
			if round >= 10 {
				return 10
			}

			return 0
		},
		GetSubroundsTimingByRoundCalled: func(round uint64) config.ConsensusConfigByRound {
			if round >= 10 {
				return timingAfterBoundary
			}

			return timingBeforeBoundary
		},
	}

	arg := getDefaultChronologyArg()
	arg.ConfigsHandler = configsHandlerStub
	roundHandlerMock := &round.RoundHandlerMock{}
	arg.RoundHandler = roundHandlerMock

	chr, err := chronology.NewChronology(arg)
	require.Nil(t, err)

	var setStartCalls, setEndCalls, sigPercentCalls, thresholdCalls int32
	var lastStart, lastEnd, lastSigPercent float64
	var lastThreshold int32

	// register the four subrounds in id order (StartRound, Block, Signature, EndRound)
	for i := 0; i < 4; i++ {
		current := i
		srm := &mock.SubroundHandlerMock{}
		srm.CurrentCalled = func() int { return current }
		srm.NextCalled = func() int { return current + 1 }
		srm.DoWorkCalled = func(consensus.RoundHandler) bool { return false }
		srm.NameCalled = func() string { return "(TEST)" }

		// count timing applications on the start-round subround only
		if current == bls.SrStartRound {
			srm.SetStartTimePercentageCalled = func(startTimePercent float64) {
				atomic.AddInt32(&setStartCalls, 1)
				lastStart = startTimePercent
			}
			srm.SetEndTimePercentageCalled = func(endTimePercent float64) {
				atomic.AddInt32(&setEndCalls, 1)
				lastEnd = endTimePercent
			}
			srm.SetProcessingThresholdPercentCalled = func(percent int) {
				atomic.AddInt32(&thresholdCalls, 1)
				atomic.StoreInt32(&lastThreshold, int32(percent))
			}
		}
		// capture the signature subround end time percentage pushed onto the block subround
		if current == bls.SrBlock {
			srm.SetSignatureSubroundEndTimePercentageCalled = func(percent float64) {
				atomic.AddInt32(&sigPercentCalls, 1)
				lastSigPercent = percent
			}
		}
		chr.AddSubround(srm)
	}

	// round 0 is in the base timing boundary, already applied at subrounds generation -> no reconciliation
	roundHandlerMock.RoundIndex = 0
	chr.InitRound()

	require.Equal(t, int32(0), atomic.LoadInt32(&setStartCalls))
	require.Equal(t, int32(0), atomic.LoadInt32(&setEndCalls))
	require.Equal(t, int32(0), atomic.LoadInt32(&sigPercentCalls))
	require.Equal(t, int32(0), atomic.LoadInt32(&thresholdCalls))

	// round advances but stays within the same timing boundary -> still no reconciliation
	roundHandlerMock.RoundIndex = 5
	chr.InitRound()

	require.Equal(t, int32(0), atomic.LoadInt32(&setStartCalls))
	require.Equal(t, int32(0), atomic.LoadInt32(&setEndCalls))
	require.Equal(t, int32(0), atomic.LoadInt32(&sigPercentCalls))
	require.Equal(t, int32(0), atomic.LoadInt32(&thresholdCalls))

	// round crosses into the new timing boundary -> setters fire with the new values
	roundHandlerMock.RoundIndex = 10
	chr.InitRound()

	require.Equal(t, int32(1), atomic.LoadInt32(&setStartCalls))
	require.Equal(t, int32(1), atomic.LoadInt32(&setEndCalls))
	require.Equal(t, timingAfterBoundary.SubroundsTiming[bls.SrStartRound].StartTime, lastStart)
	require.Equal(t, timingAfterBoundary.SubroundsTiming[bls.SrStartRound].EndTime, lastEnd)
	require.Equal(t, int32(1), atomic.LoadInt32(&sigPercentCalls))
	require.Equal(t, timingAfterBoundary.SubroundsTiming[bls.SrSignature].EndTime, lastSigPercent)
	require.Equal(t, int32(1), atomic.LoadInt32(&thresholdCalls))
	require.Equal(t, int32(timingAfterBoundary.ProcessingThresholdPercent), atomic.LoadInt32(&lastThreshold))

	// round advances further but the boundary is unchanged -> no re-application
	roundHandlerMock.RoundIndex = 20
	chr.InitRound()

	require.Equal(t, int32(1), atomic.LoadInt32(&setStartCalls))
	require.Equal(t, int32(1), atomic.LoadInt32(&setEndCalls))
	require.Equal(t, int32(1), atomic.LoadInt32(&sigPercentCalls))
	require.Equal(t, int32(1), atomic.LoadInt32(&thresholdCalls))
}

func TestChronology_RemoveAllSubroundsResetsTimingBoundary(t *testing.T) {
	t.Parallel()

	timingBeforeBoundary := config.ConsensusConfigByRound{
		EnableRound: 0,
		SubroundsTiming: []config.SubroundTiming{
			{StartTime: 0.0, EndTime: 0.05},
			{StartTime: 0.05, EndTime: 0.25},
			{StartTime: 0.25, EndTime: 0.85},
			{StartTime: 0.85, EndTime: 0.95},
		},
		ProcessingThresholdPercent: 85,
	}
	timingAfterBoundary := config.ConsensusConfigByRound{
		EnableRound: 10,
		SubroundsTiming: []config.SubroundTiming{
			{StartTime: 0.0, EndTime: 0.10},
			{StartTime: 0.10, EndTime: 0.30},
			{StartTime: 0.30, EndTime: 0.80},
			{StartTime: 0.80, EndTime: 0.90},
		},
		ProcessingThresholdPercent: 90,
	}

	configsHandlerStub := &testscommon.CommonConfigsHandlerStub{
		GetActiveTimingBoundaryRoundCalled: func(round uint64) uint64 {
			if round >= 10 {
				return 10
			}

			return 0
		},
		GetSubroundsTimingByRoundCalled: func(round uint64) config.ConsensusConfigByRound {
			if round >= 10 {
				return timingAfterBoundary
			}

			return timingBeforeBoundary
		},
	}

	arg := getDefaultChronologyArg()
	arg.ConfigsHandler = configsHandlerStub
	roundHandlerMock := &round.RoundHandlerMock{}
	arg.RoundHandler = roundHandlerMock

	chr, err := chronology.NewChronology(arg)
	require.Nil(t, err)

	var setStartCalls int32
	var lastStart float64
	addSubrounds := func() {
		for i := 0; i < 4; i++ {
			current := i
			srm := &mock.SubroundHandlerMock{}
			srm.CurrentCalled = func() int { return current }
			srm.NextCalled = func() int { return current + 1 }
			srm.DoWorkCalled = func(consensus.RoundHandler) bool { return false }
			srm.NameCalled = func() string { return "(TEST)" }
			if current == bls.SrStartRound {
				srm.SetStartTimePercentageCalled = func(startTimePercent float64) {
					atomic.AddInt32(&setStartCalls, 1)
					lastStart = startTimePercent
				}
			}
			chr.AddSubround(srm)
		}
	}

	addSubrounds()

	// cross into the new timing boundary -> setters fire with the boundary values
	roundHandlerMock.RoundIndex = 10
	chr.InitRound()
	require.Equal(t, int32(1), atomic.LoadInt32(&setStartCalls))
	require.Equal(t, timingAfterBoundary.SubroundsTiming[bls.SrStartRound].StartTime, lastStart)

	// regenerate the subrounds (as done on an epoch / consensus-type switch); the freshly generated
	// subrounds use base timing and must be reconciled again to the currently active boundary
	chr.RemoveAllSubrounds()
	addSubrounds()

	roundHandlerMock.RoundIndex = 11
	chr.InitRound()

	// without the boundary reset in RemoveAllSubrounds this would early-return and leave base timing
	require.Equal(t, int32(2), atomic.LoadInt32(&setStartCalls))
	require.Equal(t, timingAfterBoundary.SubroundsTiming[bls.SrStartRound].StartTime, lastStart)
}
