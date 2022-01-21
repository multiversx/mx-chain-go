package chronology

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/closing"
	"github.com/ElrondNetwork/elrond-go-core/display"
	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/ntp"
)

var _ consensus.ChronologyHandler = (*chronology)(nil)
var _ closing.Closer = (*chronology)(nil)

var log = logger.GetOrCreate("consensus/chronology")

// srBeforeStartRound defines the state which exist before the start of the round
const srBeforeStartRound = -1

const numRoundsToWaitBeforeSignalingChronologyStuck = 10
const chronologyAlarmID = "chronology"

// chronology defines the data needed by the chronology
type chronology struct {
	genesisTime time.Time

	roundHandler consensus.RoundHandler
	syncTimer    ntp.SyncTimer

	subroundId int

	subrounds        map[int]int
	subroundHandlers []consensus.SubroundHandler
	mutSubrounds     sync.RWMutex
	appStatusHandler core.AppStatusHandler
	cancelFunc       func()

	watchdog core.WatchdogTimer
}

// NewChronology creates a new chronology object
func NewChronology(arg ArgChronology) (*chronology, error) {

	err := checkNewChronologyParams(arg)
	if err != nil {
		return nil, err
	}

	chr := chronology{
		genesisTime:      arg.GenesisTime,
		roundHandler:     arg.RoundHandler,
		syncTimer:        arg.SyncTimer,
		appStatusHandler: arg.AppStatusHandler,
		watchdog:         arg.Watchdog,
	}

	chr.subroundId = srBeforeStartRound

	chr.subrounds = make(map[int]int)
	chr.subroundHandlers = make([]consensus.SubroundHandler, 0)

	return &chr, nil
}

func checkNewChronologyParams(arg ArgChronology) error {

	if check.IfNil(arg.RoundHandler) {
		return ErrNilRoundHandler
	}
	if check.IfNil(arg.SyncTimer) {
		return ErrNilSyncTimer
	}
	if check.IfNil(arg.Watchdog) {
		return ErrNilWatchdog
	}
	if check.IfNil(arg.AppStatusHandler) {
		return ErrNilAppStatusHandler
	}

	return nil
}

// AddSubround adds new SubroundHandler implementation to the chronology
func (chr *chronology) AddSubround(subroundHandler consensus.SubroundHandler) {
	chr.mutSubrounds.Lock()

	chr.subrounds[subroundHandler.Current()] = len(chr.subroundHandlers)
	chr.subroundHandlers = append(chr.subroundHandlers, subroundHandler)

	chr.mutSubrounds.Unlock()
}

// RemoveAllSubrounds removes all the SubroundHandler implementations added to the chronology
func (chr *chronology) RemoveAllSubrounds() {
	chr.mutSubrounds.Lock()

	chr.subrounds = make(map[int]int)
	chr.subroundHandlers = make([]consensus.SubroundHandler, 0)

	chr.mutSubrounds.Unlock()
}

// StartRounds actually starts the chronology and calls the DoWork() method of the subroundHandlers loaded
func (chr *chronology) StartRounds() {
	watchdogAlarmDuration := chr.roundHandler.TimeDuration() * numRoundsToWaitBeforeSignalingChronologyStuck
	chr.watchdog.SetDefault(watchdogAlarmDuration, chronologyAlarmID)

	var ctx context.Context
	ctx, chr.cancelFunc = context.WithCancel(context.Background())
	go chr.startRounds(ctx)
}

func (chr *chronology) startRounds(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("chronology's go routine is stopping...")
			return
		case <-time.After(time.Millisecond):
		}

		chr.startRound(ctx)
	}
}

// startRound calls the current subround, given by the finished tasks in this round
func (chr *chronology) startRound(ctx context.Context) {
	if chr.subroundId == srBeforeStartRound {
		chr.updateRound()
	}

	if chr.roundHandler.BeforeGenesis() {
		return
	}

	sr := chr.loadSubroundHandler(chr.subroundId)
	if sr == nil {
		return
	}

	msg := fmt.Sprintf("SUBROUND %s BEGINS", sr.Name())
	log.Debug(display.Headline(msg, chr.syncTimer.FormattedCurrentTime(), "."))
	logger.SetCorrelationSubround(sr.Name())

	if !sr.DoWork(ctx, chr.roundHandler) {
		chr.subroundId = srBeforeStartRound
		return
	}

	chr.subroundId = sr.Next()
}

// updateRound updates rounds and subrounds depending of the current time and the finished tasks
func (chr *chronology) updateRound() {
	oldRoundIndex := chr.roundHandler.Index()
	chr.roundHandler.UpdateRound(chr.genesisTime, chr.syncTimer.CurrentTime())

	if oldRoundIndex != chr.roundHandler.Index() {
		chr.watchdog.Reset(chronologyAlarmID)
		msg := fmt.Sprintf("ROUND %d BEGINS (%d)", chr.roundHandler.Index(), chr.roundHandler.TimeStamp().Unix())
		log.Debug(display.Headline(msg, chr.syncTimer.FormattedCurrentTime(), "#"))
		logger.SetCorrelationRound(chr.roundHandler.Index())

		chr.initRound()
	}
}

// initRound is called when a new round begins and it does the necessary initialization
func (chr *chronology) initRound() {
	chr.subroundId = srBeforeStartRound

	chr.mutSubrounds.RLock()

	hasSubroundsAndGenesisTimePassed := !chr.roundHandler.BeforeGenesis() && len(chr.subroundHandlers) > 0

	if hasSubroundsAndGenesisTimePassed {
		chr.subroundId = chr.subroundHandlers[0].Current()
		chr.appStatusHandler.SetUInt64Value(common.MetricCurrentRound, uint64(chr.roundHandler.Index()))
		chr.appStatusHandler.SetUInt64Value(common.MetricCurrentRoundTimestamp, uint64(chr.roundHandler.TimeStamp().Unix()))
	}

	chr.mutSubrounds.RUnlock()
}

// loadSubroundHandler returns the implementation of SubroundHandler given by the subroundId
func (chr *chronology) loadSubroundHandler(subroundId int) consensus.SubroundHandler {
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

// Close will close the endless running go routine
func (chr *chronology) Close() error {
	if chr.cancelFunc != nil {
		chr.cancelFunc()
	}

	chr.watchdog.Stop(chronologyAlarmID)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (chr *chronology) IsInterfaceNil() bool {
	return chr == nil
}
