package components

import (
	"context"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/round"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

// The tests in this file prove that simChronology mirrors the production chronology's
// round/subround stepping semantics (consensus/chronology/chronology.go): round initialization on
// a round-index change only, at most one subround DoWork per pass, DoWork-false reset to the
// before-start-round state, subround chaining through Next(), the timing-boundary application on
// round change, its reset by RemoveAllSubrounds, and the one-shot Supernova base-duration
// transition. Any behavioral change in the production chronology must be mirrored in
// simChronology and guarded here.

const parityTestRoundDuration = time.Second

var parityTestGenesisTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

// newParityRoundHandler builds the PRODUCTION round handler (consensus/round) over the manual sync
// timer, exactly as the consensus-mode core components do, so the parity tests exercise the same
// round arithmetic the simulator runs with.
func newParityRoundHandler(t *testing.T, timer *manualSyncTimer) consensus.RoundHandler {
	roundHandler, err := round.NewRound(round.ArgsRound{
		GenesisTimeStamp:          parityTestGenesisTime,
		SupernovaGenesisTimeStamp: parityTestGenesisTime,
		CurrentTimeStamp:          timer.CurrentTime(),
		RoundTimeDuration:         parityTestRoundDuration,
		SupernovaTimeDuration:     parityTestRoundDuration,
		SyncTimer:                 timer,
		EnableRoundsHandler:       &testscommon.EnableRoundsHandlerStub{},
	})
	require.Nil(t, err)

	return roundHandler
}

func newParityArgs(timer *manualSyncTimer, roundHandler consensus.RoundHandler, tickChan <-chan time.Time) ArgsSimChronology {
	return ArgsSimChronology{
		GenesisTime:         parityTestGenesisTime,
		RoundHandler:        roundHandler,
		SyncTimer:           timer,
		AppStatusHandler:    statusHandlerMock.NewAppStatusHandlerMock(),
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EnableRoundsHandler: &testscommon.EnableRoundsHandlerStub{},
		ConfigsHandler:      testscommon.GetDefaultCommonConfigsHandler(),
		TickChan:            tickChan,
		StepDoneChan:        make(chan struct{}, 100),
	}
}

func newParitySubround(current int, next int, doWork func() bool) *mock.SubroundHandlerMock {
	return &mock.SubroundHandlerMock{
		CurrentCalled: func() int { return current },
		NextCalled:    func() int { return next },
		DoWorkCalled: func(_ consensus.RoundHandler) bool {
			return doWork()
		},
		NameCalled:             func() string { return "(TEST)" },
		ConsensusChannelCalled: func() chan bool { return make(chan bool, 1) },
	}
}

// TestSimChronology_TickDrivenRounds mirrors the production chronology's manual-drive semantics
// through the real tick loop (StartRounds' goroutine body): the injected tick channel replaces the
// production 1ms polling and the manual sync timer replaces the NTP clock. One round passes exactly
// when the test advances the timer by one round duration and sends one tick; a tick without time
// advancement must not start a new round (the DoWork-false reset leaves the chronology in the
// before-start-round state, and re-initialization happens only on a round-index change).
func TestSimChronology_TickDrivenRounds(t *testing.T) {
	t.Parallel()

	manualTimer := NewManualSyncTimer(parityTestGenesisTime)
	roundHandler := newParityRoundHandler(t, manualTimer)

	tick := make(chan time.Time)
	chr, err := NewSimChronology(newParityArgs(manualTimer, roundHandler, tick))
	require.Nil(t, err)

	// arm the loop the way StartRounds does, but drive its goroutine directly so the test controls
	// its lifecycle deterministically (StartRounds owns the goroutine internally)
	chr.running = true

	roundStarted := make(chan struct{})
	numDoWorkCalls := 0
	subround := newParitySubround(0, 1, func() bool {
		numDoWorkCalls++
		roundStarted <- struct{}{}
		// returning false sends the chronology back to the before-start-round state, so the
		// next loop iteration checks for a new round again
		return false
	})
	chr.AddSubround(subround)

	// a chained subround that must never run: DoWork returns false, so the chronology resets
	// instead of advancing to Next()
	numChainedCalls := 0
	chained := newParitySubround(1, 2, func() bool {
		numChainedCalls++
		return false
	})
	chr.AddSubround(chained)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		chr.startRounds(ctx)
		close(done)
	}()

	// each (advance round duration + tick) pair executes exactly one round
	numRounds := 3
	for i := 0; i < numRounds; i++ {
		manualTimer.AdvanceTime(parityTestRoundDuration)
		tick <- time.Time{}
		<-roundStarted
	}

	// a tick without time advancement must not start a new round; if it wrongly ran DoWork the
	// send on roundStarted would block forever and the test would time out on <-done below
	tick <- time.Time{}

	cancel()
	<-done

	assert.Equal(t, int64(numRounds), roundHandler.Index())
	assert.Equal(t, numRounds, numDoWorkCalls)
	assert.Zero(t, numChainedCalls, "DoWork false must reset to before-start-round, not chain to Next()")
}

// TestSimChronology_OneTickOneSubroundAndChaining drives startRound synchronously and proves:
// (a) one pass executes AT MOST one subround DoWork, even when DoWork succeeds (the next subround
// runs only on the next tick); (c) the subround executed next is the one announced by Next();
// (b) a DoWork failure resets to before-start-round, a tick without a round change does nothing,
// and the next round re-initializes from the first subround.
func TestSimChronology_OneTickOneSubroundAndChaining(t *testing.T) {
	t.Parallel()

	manualTimer := NewManualSyncTimer(parityTestGenesisTime)
	roundHandler := newParityRoundHandler(t, manualTimer)

	// the tick channel is unused: startRound is invoked directly, one call == one tick
	chr, err := NewSimChronology(newParityArgs(manualTimer, roundHandler, make(chan time.Time)))
	require.Nil(t, err)

	callOrder := make([]string, 0)
	numACalls, numBCalls := 0, 0
	subroundA := newParitySubround(0, 1, func() bool {
		numACalls++
		callOrder = append(callOrder, "A")
		return true // success: the chronology must advance to Next() == 1, but only on the NEXT tick
	})
	subroundB := newParitySubround(1, 2, func() bool {
		numBCalls++
		callOrder = append(callOrder, "B")
		return false // failure: reset to before-start-round
	})
	chr.AddSubround(subroundA)
	chr.AddSubround(subroundB)

	ctx := context.Background()

	// tick 1: the round index changes (0 -> 1), the round initializes and ONLY subround A runs,
	// even though its DoWork succeeded — one tick, at most one subround
	manualTimer.AdvanceTime(parityTestRoundDuration)
	chr.startRound(ctx)
	assert.Equal(t, 1, numACalls)
	assert.Equal(t, 0, numBCalls, "a successful DoWork must not cascade into the next subround within the same tick")

	// tick 2 (no time advance): the chained subround B — A's Next() — runs
	chr.startRound(ctx)
	assert.Equal(t, 1, numACalls)
	assert.Equal(t, 1, numBCalls)

	// tick 3 (no time advance): B returned false, so the chronology sits before-start-round;
	// with an unchanged round index nothing runs
	chr.startRound(ctx)
	assert.Equal(t, 1, numACalls)
	assert.Equal(t, 1, numBCalls)

	// tick 4, new round: re-initialization starts again from the FIRST subround
	manualTimer.AdvanceTime(parityTestRoundDuration)
	chr.startRound(ctx)
	assert.Equal(t, 2, numACalls)
	assert.Equal(t, 1, numBCalls)

	assert.Equal(t, []string{"A", "B", "A"}, callOrder)
	assert.Equal(t, int64(2), roundHandler.Index())
}

func TestSimChronology_RearmRetriesSameManualClockRound(t *testing.T) {
	t.Parallel()

	manualTimer := NewManualSyncTimer(parityTestGenesisTime)
	roundHandler := newParityRoundHandler(t, manualTimer)
	chr, err := NewSimChronology(newParityArgs(manualTimer, roundHandler, make(chan time.Time)))
	require.Nil(t, err)

	numCalls := 0
	chr.AddSubround(newParitySubround(0, 1, func() bool {
		numCalls++
		return false
	}))

	manualTimer.AdvanceTime(parityTestRoundDuration)
	chr.startRound(context.Background())
	require.Equal(t, 1, numCalls)
	require.Equal(t, int64(1), roundHandler.Index())

	// A retry at the unchanged manual time is a no-op until the production round handler is
	// reverted once. The simulator's RearmConsensusRound uses this exact operation.
	chr.startRound(context.Background())
	require.Equal(t, 1, numCalls)

	rearmable, ok := roundHandler.(consensus.RoundHandlerConsensusSwitch)
	require.True(t, ok)
	rearmable.RevertOneRound()
	chr.startRound(context.Background())

	assert.Equal(t, 2, numCalls)
	assert.Equal(t, int64(1), roundHandler.Index(), "rearming must retry, not advance, the round")
}

func TestSimChronology_ConsensusSwitchDoesNotLetInFlightSubroundOverwriteRestart(t *testing.T) {
	t.Parallel()

	manualTimer := NewManualSyncTimer(parityTestGenesisTime)
	roundHandler := newParityRoundHandler(t, manualTimer)
	chr, err := NewSimChronology(newParityArgs(manualTimer, roundHandler, make(chan time.Time)))
	require.NoError(t, err)

	newSubround := newParitySubround(0, 1, func() bool { return false })
	oldSubround := newParitySubround(0, 1, func() bool {
		require.NoError(t, chr.Close())
		chr.RemoveAllSubrounds()
		chr.AddSubround(newSubround)
		chr.StartRounds()

		return true
	})
	chr.AddSubround(oldSubround)
	chr.StartRounds()
	defer chr.stop()

	manualTimer.AdvanceTime(parityTestRoundDuration)
	chr.startRound(context.Background())

	assert.Equal(t, srBeforeStartRound, chr.getSubroundId(),
		"the old subround must not advance the freshly restarted chronology")
}

// TestSimChronology_RemoveAllSubroundsResetsTimingBoundaryState proves (d): the dynamic
// subround-timing boundary is applied when the active boundary differs from the last applied one,
// skipped while it is unchanged, and re-applied after RemoveAllSubrounds resets the boundary
// state — exactly like the production chronology's RemoveAllSubrounds.
func TestSimChronology_RemoveAllSubroundsResetsTimingBoundaryState(t *testing.T) {
	t.Parallel()

	manualTimer := NewManualSyncTimer(parityTestGenesisTime)
	roundHandler := newParityRoundHandler(t, manualTimer)

	args := newParityArgs(manualTimer, roundHandler, make(chan time.Time))
	args.ConfigsHandler = &testscommon.CommonConfigsHandlerStub{
		// a non-genesis boundary so the very first round-change already differs from the
		// zero-value lastTimingBoundaryEnableRound and must apply the timing
		GetActiveTimingBoundaryRoundCalled: func(_ uint64) uint64 { return 7 },
	}
	chr, err := NewSimChronology(args)
	require.Nil(t, err)

	numTimingApplications := 0
	numThresholdApplications := 0
	subround := newParitySubround(0, 1, func() bool { return false })
	subround.SetTimingPercentageCalled = func(_ float64, _ float64) {
		numTimingApplications++
	}
	subround.SetProcessingThresholdPercentCalled = func(_ int) {
		numThresholdApplications++
	}
	chr.AddSubround(subround)

	ctx := context.Background()

	// round 1: boundary 7 != last applied 0 -> the timing is pushed to the subround handlers
	manualTimer.AdvanceTime(parityTestRoundDuration)
	chr.startRound(ctx)
	assert.Equal(t, 1, numTimingApplications)
	assert.Equal(t, 1, numThresholdApplications)

	// round 2: boundary unchanged -> no re-application
	manualTimer.AdvanceTime(parityTestRoundDuration)
	chr.startRound(ctx)
	assert.Equal(t, 1, numTimingApplications)

	// RemoveAllSubrounds resets the boundary state (and the subround position); after re-adding
	// the subrounds the next round change must re-apply the timing
	chr.RemoveAllSubrounds()
	chr.AddSubround(subround)

	manualTimer.AdvanceTime(parityTestRoundDuration)
	chr.startRound(ctx)
	assert.Equal(t, 2, numTimingApplications)
	assert.Equal(t, 2, numThresholdApplications)
}

// TestSimChronology_SupernovaTransitionAppliesBaseDurationOnce proves the Supernova transition
// parity: once the Supernova flag is active and the activation round is reached, every subround
// gets SetBaseDuration(roundHandler.TimeDuration()) exactly once — the transition is one-shot,
// exactly like the production chronology's handleSupernovaTransitionIfNeeded.
func TestSimChronology_SupernovaTransitionAppliesBaseDurationOnce(t *testing.T) {
	t.Parallel()

	manualTimer := NewManualSyncTimer(parityTestGenesisTime)
	roundHandler := newParityRoundHandler(t, manualTimer)

	args := newParityArgs(manualTimer, roundHandler, make(chan time.Time))
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.SupernovaFlag
		},
	}
	// activation round 0 (the EnableRoundsHandlerStub default) is always <= the current round,
	// so the transition triggers on the first initialized round
	chr, err := NewSimChronology(args)
	require.Nil(t, err)

	numBaseDurationCalls := 0
	subround := newParitySubround(0, 1, func() bool { return false })
	subround.SetBaseDurationCalled = func(baseDuration time.Duration) {
		numBaseDurationCalls++
		assert.Equal(t, roundHandler.TimeDuration(), baseDuration)
	}
	chr.AddSubround(subround)

	ctx := context.Background()

	manualTimer.AdvanceTime(parityTestRoundDuration)
	chr.startRound(ctx)
	assert.Equal(t, 1, numBaseDurationCalls)

	// the transition is one-shot: the next round must not re-apply the base duration
	manualTimer.AdvanceTime(parityTestRoundDuration)
	chr.startRound(ctx)
	assert.Equal(t, 1, numBaseDurationCalls)
}
