package chronology

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
)

var log = logger.NewDefaultLogger()

// chronology defines the data needed by the chronology
type chronology struct {
	genesisTime time.Time

	rounder   round.Rounder
	syncTimer ntp.SyncTimer

	subroundId int

	subrounds        map[int]int
	subroundHandlers []SubroundHandler
	mutSubrounds     sync.RWMutex
}

// NewChronology defines a new chronology object
func NewChronology(
	genesisTime time.Time,
	rounder round.Rounder,
	syncTimer ntp.SyncTimer,
) (*chronology, error) {

	err := checkNewChronologyParams(
		rounder,
		syncTimer,
	)

	if err != nil {
		return nil, err
	}

	chr := chronology{
		genesisTime: genesisTime,
		rounder:     rounder,
		syncTimer:   syncTimer}

	chr.subroundId = -1

	chr.subrounds = make(map[int]int)
	chr.subroundHandlers = make([]SubroundHandler, 0)

	return &chr, nil
}

func checkNewChronologyParams(
	rounder round.Rounder,
	syncTimer ntp.SyncTimer,
) error {

	if rounder == nil {
		return spos.ErrNilRounder
	}

	if syncTimer == nil {
		return spos.ErrNilSyncTimer
	}

	return nil
}

// AddSubround adds new SubroundHandler implementation to the chronology
func (chr *chronology) AddSubround(subroundHandler SubroundHandler) {
	chr.mutSubrounds.Lock()

	chr.subrounds[subroundHandler.Current()] = len(chr.subroundHandlers)
	chr.subroundHandlers = append(chr.subroundHandlers, subroundHandler)

	chr.mutSubrounds.Unlock()
}

// RemoveAllSubrounds removes all the SubroundHandler implementations added to the chronology
func (chr *chronology) RemoveAllSubrounds() {
	chr.mutSubrounds.Lock()

	chr.subrounds = make(map[int]int)
	chr.subroundHandlers = make([]SubroundHandler, 0)

	chr.mutSubrounds.Unlock()
}

// StartRounds actually starts the chronology and calls the DoWork() method of the subroundHandlers loaded
func (chr *chronology) StartRounds() {
	for {
		time.Sleep(time.Millisecond)
		chr.startRound()
	}
}

// startRound calls the current subround, given by the finished tasks in this round
func (chr *chronology) startRound() {
	chr.updateRound()

	if chr.rounder.Index() < 0 {
		return
	}

	sr := chr.loadSubroundHandler(chr.subroundId)

	if sr == nil {
		return
	}

	log.Info(fmt.Sprintf(
		"\n%s.................... SUBROUND %s BEGINS ....................\n\n",
		chr.syncTimer.FormattedCurrentTime(),
		sr.Name(),
	))

	if !sr.DoWork(chr.haveTimeInCurrentRound) {
		return
	}

	chr.subroundId = sr.Next()
}

// updateRound updates rounds and subrounds depending of the current time and the finished tasks
func (chr *chronology) updateRound() {
	oldRoundIndex := chr.rounder.Index()
	chr.rounder.UpdateRound(chr.genesisTime, chr.syncTimer.CurrentTime())

	if oldRoundIndex != chr.rounder.Index() {
		log.Info(fmt.Sprintf(
			"\n%s############################## ROUND %d BEGINS (%d) ##############################\n\n",
			chr.syncTimer.FormattedCurrentTime(),
			chr.rounder.Index(),
			chr.rounder.TimeStamp().Unix()))

		chr.initRound()
	}
}

// initRound is called when a new round begins and it does the necessary initialization
func (chr *chronology) initRound() {
	chr.subroundId = -1

	chr.mutSubrounds.RLock()

	if len(chr.subroundHandlers) > 0 {
		chr.subroundId = chr.subroundHandlers[0].Current()
	}

	chr.mutSubrounds.RUnlock()
}

// loadSubroundHandler returns the implementation of SubroundHandler given by the subroundId
func (chr *chronology) loadSubroundHandler(subroundId int) SubroundHandler {
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

func (chr *chronology) haveTimeInCurrentRound() time.Duration {
	roundStartTime := chr.rounder.TimeStamp()
	currentTime := chr.syncTimer.CurrentTime()
	elapsedTime := currentTime.Sub(roundStartTime)
	haveTime := float64(chr.rounder.TimeDuration()) - float64(elapsedTime)

	return time.Duration(haveTime)
}
