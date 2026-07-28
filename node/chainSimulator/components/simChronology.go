package components

import (
	"context"
	goErrors "errors"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/ntp"
)

var _ consensus.ChronologyHandler = (*simChronology)(nil)

type consensusWaitInterrupter interface {
	interruptConsensusWait()
}

var (
	errNilRoundHandler     = goErrors.New("nil round handler")
	errNilSyncTimer        = goErrors.New("nil sync timer")
	errNilAppStatusHandler = goErrors.New("nil app status handler")
	errNilTickChan         = goErrors.New("nil tick channel")
)

// srBeforeStartRound mirrors the production chronology's before-start-round subround state
const srBeforeStartRound = -1

// ArgsSimChronology holds the dependencies of the simulator-owned chronology
type ArgsSimChronology struct {
	GenesisTime         time.Time
	RoundHandler        consensus.RoundHandler
	SyncTimer           ntp.SyncTimer
	AppStatusHandler    core.AppStatusHandler
	EnableEpochsHandler common.EnableEpochsHandler
	EnableRoundsHandler common.EnableRoundsHandler
	ConfigsHandler      common.CommonConfigsHandler
	// TickChan is the wake-up source of the rounds processing loop: one received tick executes
	// one startRound pass (at most one subround), replacing the production chronology's 1ms
	// real-time polling so the drive can step consensus explicitly
	TickChan <-chan time.Time
	// StepDoneChan acknowledges that the subround started by one tick has returned.
	StepDoneChan chan<- struct{}
}

// simChronology is the simulator-owned implementation of consensus.ChronologyHandler. It mirrors
// the production chronology's round/subround stepping semantics (round initialization, DoWork
// chaining, dynamic timing boundaries and the Supernova base-duration transition) but wakes up
// only on the injected tick channel, so consensus advances exactly when the drive says so. Any
// behavioral change in the production chronology must be mirrored here; the parity test guards
// the overlap.
type simChronology struct {
	genesisTime time.Time

	roundHandler consensus.RoundHandler
	syncTimer    ntp.SyncTimer

	subroundId int

	// mutRun guards running and subroundId: the proxy (re)starts the chronology from the consensus
	// goroutine on every version switch while the same goroutine also reads/writes subroundId
	mutRun     sync.Mutex
	running    bool
	generation uint64
	runContext context.Context
	cancelRun  context.CancelFunc
	// subroundInFlight is true only while the chronology goroutine is inside DoWork. A
	// consensus-version switch has to wake that obsolete DoWork, but Close can also be called
	// while the chronology is idle. Signalling the shared consensus channel in the latter case
	// leaves a stale token for the newly installed consensus version.
	subroundInFlight bool

	subrounds        map[int]int
	subroundHandlers []consensus.SubroundHandler
	mutSubrounds     sync.RWMutex
	appStatusHandler core.AppStatusHandler
	cancelFunc       func()

	enableEpochsHandler           common.EnableEpochsHandler
	enableRoundsHandler           common.EnableRoundsHandler
	configsHandler                common.CommonConfigsHandler
	supernovaTransitionDone       bool
	lastTimingBoundaryEnableRound uint64

	tickChan <-chan time.Time
	stepDone chan<- struct{}
}

// NewSimChronology creates a simulator-owned, tick-driven chronology
func NewSimChronology(args ArgsSimChronology) (*simChronology, error) {
	err := checkSimChronologyArgs(args)
	if err != nil {
		return nil, err
	}

	chr := &simChronology{
		genesisTime:         args.GenesisTime,
		roundHandler:        args.RoundHandler,
		syncTimer:           args.SyncTimer,
		appStatusHandler:    args.AppStatusHandler,
		enableEpochsHandler: args.EnableEpochsHandler,
		enableRoundsHandler: args.EnableRoundsHandler,
		configsHandler:      args.ConfigsHandler,
		tickChan:            args.TickChan,
		stepDone:            args.StepDoneChan,
	}

	chr.subroundId = srBeforeStartRound
	chr.subrounds = make(map[int]int)
	chr.subroundHandlers = make([]consensus.SubroundHandler, 0)

	return chr, nil
}

func checkSimChronologyArgs(args ArgsSimChronology) error {
	if check.IfNil(args.RoundHandler) {
		return errNilRoundHandler
	}
	if check.IfNil(args.SyncTimer) {
		return errNilSyncTimer
	}
	if check.IfNil(args.AppStatusHandler) {
		return errNilAppStatusHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return errors.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.EnableRoundsHandler) {
		return errors.ErrNilEnableRoundsHandler
	}
	if check.IfNil(args.ConfigsHandler) {
		return common.ErrNilCommonConfigsHandler
	}
	if args.TickChan == nil {
		return errNilTickChan
	}
	if args.StepDoneChan == nil {
		return errNilTickChan
	}

	return nil
}

// AddSubround adds a new SubroundHandler implementation, mirroring the production chronology
func (chr *simChronology) AddSubround(subroundHandler consensus.SubroundHandler) {
	chr.mutSubrounds.Lock()

	chr.subrounds[subroundHandler.Current()] = len(chr.subroundHandlers)
	chr.subroundHandlers = append(chr.subroundHandlers, subroundHandler)

	chr.mutSubrounds.Unlock()
}

// RemoveAllSubrounds removes all the SubroundHandler implementations, mirroring the production chronology
func (chr *simChronology) RemoveAllSubrounds() {
	chr.mutSubrounds.Lock()

	chr.subrounds = make(map[int]int)
	chr.subroundHandlers = make([]consensus.SubroundHandler, 0)
	chr.setSubroundId(srBeforeStartRound)
	chr.lastTimingBoundaryEnableRound = 0

	chr.mutSubrounds.Unlock()
}

// StartRounds (re)arms the tick-driven rounds processing loop. The proxy restarts the chronology
// (Close then StartRounds) on every consensus-version switch; a SINGLE long-lived goroutine is
// reused across restarts rather than spawned anew each time. Spawning a fresh goroutine per restart
// would leave the superseded one briefly alive and — sharing the unbuffered tick channel and the
// subround state with the new one — racing on subroundId and stealing ticks. Gating the one
// goroutine on `running` keeps tick processing single-threaded across restarts and avoids the
// deadlock a blocking Close would hit (the switch is re-entrant from the chronology goroutine itself).
func (chr *simChronology) StartRounds() {
	chr.mutRun.Lock()
	defer chr.mutRun.Unlock()

	if chr.cancelRun != nil {
		chr.cancelRun()
	}
	chr.runContext, chr.cancelRun = context.WithCancel(context.Background())
	chr.running = true
	chr.subroundId = srBeforeStartRound
	chr.generation++

	// force a round update to initialize the round on the first tick after (re)start, as the
	// production chronology does at each goroutine (re)start
	if roundHandlerWithRevert, ok := chr.roundHandler.(consensus.RoundHandlerConsensusSwitch); ok {
		roundHandlerWithRevert.RevertOneRound()
	}

	if chr.cancelFunc != nil {
		return
	}

	var ctx context.Context
	ctx, chr.cancelFunc = context.WithCancel(context.Background())
	go chr.startRounds(ctx)
}

func (chr *simChronology) startRounds(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("simChronology's go routine is stopping...")
			return
		case <-chr.tickChan:
		}

		runContext, isRunning := chr.currentRunContext()
		if !isRunning {
			// closed for a consensus-version switch (the proxy will re-arm via StartRounds) or for
			// teardown: the tick is drained so the drive's step never blocks, but no subround runs
			chr.signalStepDone(ctx)
			continue
		}

		chr.startRound(runContext)
		chr.signalStepDone(ctx)
	}
}

func (chr *simChronology) signalStepDone(ctx context.Context) {
	select {
	case chr.stepDone <- struct{}{}:
	case <-ctx.Done():
	}
}

// startRound mirrors the production chronology's startRound: at most one subround DoWork per pass
func (chr *simChronology) startRound(ctx context.Context) {
	if chr.getSubroundId() == srBeforeStartRound {
		chr.updateRound()
	}

	if chr.roundHandler.BeforeGenesis() {
		return
	}

	generation := chr.getGeneration()
	sr := chr.loadSubroundHandler(chr.getSubroundId())
	if sr == nil {
		return
	}

	log.Trace("simChronology: subround begins", "name", sr.Name(), "time", chr.syncTimer.FormattedCurrentTime())

	chr.setSubroundInFlight(true)
	defer chr.setSubroundInFlight(false)

	if !sr.DoWork(ctx, chr.roundHandler) {
		chr.setSubroundIdForGeneration(srBeforeStartRound, generation)
		return
	}

	chr.setSubroundIdForGeneration(sr.Next(), generation)
}

func (chr *simChronology) currentRunContext() (context.Context, bool) {
	chr.mutRun.Lock()
	defer chr.mutRun.Unlock()

	return chr.runContext, chr.running
}

func (chr *simChronology) getSubroundId() int {
	chr.mutRun.Lock()
	defer chr.mutRun.Unlock()

	return chr.subroundId
}

func (chr *simChronology) setSubroundId(subroundId int) {
	chr.mutRun.Lock()
	defer chr.mutRun.Unlock()

	chr.subroundId = subroundId
}

func (chr *simChronology) getGeneration() uint64 {
	chr.mutRun.Lock()
	defer chr.mutRun.Unlock()

	return chr.generation
}

func (chr *simChronology) setSubroundIdForGeneration(subroundId int, generation uint64) {
	chr.mutRun.Lock()
	defer chr.mutRun.Unlock()

	if chr.generation != generation {
		return
	}

	chr.subroundId = subroundId
}

func (chr *simChronology) setSubroundInFlight(inFlight bool) {
	chr.mutRun.Lock()
	chr.subroundInFlight = inFlight
	chr.mutRun.Unlock()
}

func (chr *simChronology) updateRound() {
	oldRoundIndex := chr.roundHandler.Index()
	chr.roundHandler.UpdateRound(chr.genesisTime, chr.syncTimer.CurrentTime())

	if oldRoundIndex != chr.roundHandler.Index() {
		log.Trace("simChronology: round begins", "round", chr.roundHandler.Index(), "time", chr.syncTimer.FormattedCurrentTime())
		chr.initRound()
	}
}

func (chr *simChronology) getRoundUnixTimeStamp() int64 {
	if chr.enableEpochsHandler.IsFlagEnabled(common.SupernovaFlag) {
		return chr.roundHandler.TimeStamp().UnixMilli()
	}

	return chr.roundHandler.TimeStamp().Unix()
}

func (chr *simChronology) initRound() {
	chr.setSubroundId(srBeforeStartRound)

	chr.mutSubrounds.Lock()

	hasSubroundsAndGenesisTimePassed := !chr.roundHandler.BeforeGenesis() && len(chr.subroundHandlers) > 0

	if hasSubroundsAndGenesisTimePassed {
		chr.setSubroundId(chr.subroundHandlers[0].Current())

		roundIndex := uint64(chr.roundHandler.Index())
		chr.appStatusHandler.SetUInt64Value(common.MetricCurrentRound, roundIndex)
		chr.appStatusHandler.SetUInt64Value(common.MetricCurrentRoundTimestamp, uint64(chr.getRoundUnixTimeStamp()))

		chr.handleRoundChangedIfNeeded()
		chr.handleSupernovaTransitionIfNeeded()
	}

	chr.mutSubrounds.Unlock()
}

func (chr *simChronology) handleRoundChangedIfNeeded() {
	roundIndex := uint64(chr.roundHandler.Index())
	activeBoundary := chr.configsHandler.GetActiveTimingBoundaryRound(roundIndex)
	if activeBoundary == chr.lastTimingBoundaryEnableRound {
		return
	}

	chr.lastTimingBoundaryEnableRound = activeBoundary

	timing := chr.configsHandler.GetSubroundsTimingByRound(roundIndex)
	for _, subroundHandler := range chr.subroundHandlers {
		subroundHandler.SetProcessingThresholdPercent(int(timing.ProcessingThresholdPercent))

		idx := subroundHandler.Current()
		if idx < 0 || idx >= len(timing.SubroundsTiming) {
			log.Warn("simChronology: found subround handler with unknown index", "idx", idx, "name", subroundHandler.Name())
			continue
		}

		subroundHandler.SetTimingPercentage(timing.SubroundsTiming[idx].StartTime, timing.SubroundsTiming[idx].EndTime)
	}

	// the block subround needs the signature subround end time for managed-key signature deadline
	if len(chr.subroundHandlers) > bls.SrBlock && len(timing.SubroundsTiming) > bls.SrSignature {
		chr.subroundHandlers[bls.SrBlock].SetSignatureSubroundEndTimePercentage(timing.SubroundsTiming[bls.SrSignature].EndTime)
	}
}

func (chr *simChronology) handleSupernovaTransitionIfNeeded() {
	if chr.supernovaTransitionDone {
		return
	}

	if !chr.enableEpochsHandler.IsFlagEnabled(common.SupernovaFlag) {
		return
	}

	roundIndex := uint64(chr.roundHandler.Index())
	supernovaActivationRound := chr.enableRoundsHandler.GetActivationRound(common.SupernovaRoundFlag)
	if supernovaActivationRound > roundIndex {
		return
	}

	chr.appStatusHandler.SetUInt64Value(common.MetricRoundDuration, uint64(chr.roundHandler.TimeDuration().Milliseconds()))

	// update time duration on each subround
	for _, subroundHandler := range chr.subroundHandlers {
		subroundHandler.SetBaseDuration(chr.roundHandler.TimeDuration())
	}

	chr.supernovaTransitionDone = true
}

func (chr *simChronology) loadSubroundHandler(subroundId int) consensus.SubroundHandler {
	chr.mutSubrounds.RLock()
	defer chr.mutSubrounds.RUnlock()

	index, exist := chr.subrounds[subroundId]
	if !exist {
		return nil
	}

	indexIsOutOfBounds := index < 0 || index >= len(chr.subroundHandlers)
	if indexIsOutOfBounds {
		return nil
	}

	return chr.subroundHandlers[index]
}

// Close pauses tick processing without tearing down the goroutine. The proxy calls Close then
// StartRounds on every consensus-version switch; keeping the single goroutine alive across that
// restart keeps tick processing single-threaded (see StartRounds). The goroutine itself is
// stopped only at node teardown, via stop().
func (chr *simChronology) Close() error {
	chr.mutRun.Lock()
	chr.running = false
	if chr.cancelRun != nil {
		chr.cancelRun()
	}
	chr.mutRun.Unlock()

	chr.interruptCurrentSubround()

	return nil
}

// interruptCurrentSubround wakes an obsolete in-flight simulator subround without leaving a token
// in the shared consensus channel for the replacement consensus version.
func (chr *simChronology) interruptCurrentSubround() {
	chr.mutRun.Lock()
	subroundID := chr.subroundId
	generation := chr.generation
	subroundInFlight := chr.subroundInFlight
	chr.mutRun.Unlock()
	if !subroundInFlight {
		return
	}

	subroundHandler := chr.loadSubroundHandler(subroundID)
	if subroundHandler == nil {
		return
	}

	// Recheck while holding mutRun. DoWork's deferred in-flight reset takes the same lock, so it
	// cannot finish between this check and the non-blocking signal and turn the signal into a stale
	// token consumed by the following subround.
	chr.mutRun.Lock()
	defer chr.mutRun.Unlock()
	if !chr.subroundInFlight || chr.subroundId != subroundID || chr.generation != generation {
		return
	}

	if roundHandler, ok := chr.roundHandler.(consensusWaitInterrupter); ok {
		roundHandler.interruptConsensusWait()
	}

	select {
	case subroundHandler.ConsensusChannel() <- true:
	default:
	}
}

// stop permanently terminates the rounds processing goroutine. Used only at node teardown; the
// proxy's epoch-switch restart uses Close, which must not kill the goroutine.
func (chr *simChronology) stop() {
	chr.mutRun.Lock()
	chr.running = false
	if chr.cancelRun != nil {
		chr.cancelRun()
	}
	if chr.cancelFunc != nil {
		chr.cancelFunc()
	}
	chr.mutRun.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (chr *simChronology) IsInterfaceNil() bool {
	return chr == nil
}
