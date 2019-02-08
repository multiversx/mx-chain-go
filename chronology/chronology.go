package chronology

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

// sleepTime defines the time in milliseconds between each iteration made in StartRounds method
const sleepTime = time.Duration(5 * time.Millisecond)

var log = logger.NewDefaultLogger()

// SubroundId defines the type used to refer the current subround
type SubroundId int

// Chronology defines the data needed by the chronology
type Chronology struct {
	DoRun      bool
	doSyncMode bool

	round       *Round
	genesisTime time.Time

	selfSubround SubroundId
	timeSubround SubroundId
	clockOffset  time.Duration

	subroundHandlers []SubroundHandler
	subrounds        map[SubroundId]int
	mutSubrounds     sync.RWMutex
	syncTimer        ntp.SyncTimer
}

// NewChronology defines a new Chr object
func NewChronology(
	doSyncMode bool,
	round *Round,
	genesisTime time.Time,
	syncTimer ntp.SyncTimer) *Chronology {

	chr := Chronology{
		doSyncMode:  doSyncMode,
		round:       round,
		genesisTime: genesisTime,
		syncTimer:   syncTimer}

	chr.DoRun = true

	chr.SetSelfSubround(-1)
	chr.timeSubround = -1
	chr.clockOffset = syncTimer.ClockOffset()

	chr.subroundHandlers = make([]SubroundHandler, 0)
	chr.subrounds = make(map[SubroundId]int)

	return &chr
}

// initRound is called when a new round begins and do the necessary initialization
func (chr *Chronology) initRound() {
	chr.SetSelfSubround(-1)

	chr.mutSubrounds.RLock()

	if len(chr.subroundHandlers) > 0 {
		chr.SetSelfSubround(chr.subroundHandlers[0].Current())
	}

	chr.mutSubrounds.RUnlock()

	chr.clockOffset = chr.syncTimer.ClockOffset()
}

// RemoveAllSubrounds removes all the SubroundHandler implementations added to the chronology
func (chr *Chronology) RemoveAllSubrounds() {
	chr.mutSubrounds.Lock()

	chr.subroundHandlers = make([]SubroundHandler, 0)
	chr.subrounds = make(map[SubroundId]int)

	chr.mutSubrounds.Unlock()
}

// AddSubround adds new SubroundHandler implementation to the chronology. This method **must** be called in
// chronological order of the subroundHandlers.EndTime()
func (chr *Chronology) AddSubround(subroundHandler SubroundHandler) {
	chr.mutSubrounds.Lock()

	if chr.subrounds == nil {
		chr.subrounds = make(map[SubroundId]int)
	}

	chr.subrounds[subroundHandler.Current()] = len(chr.subroundHandlers)
	chr.subroundHandlers = append(chr.subroundHandlers, subroundHandler)

	chr.mutSubrounds.Unlock()
}

// StartRounds actually starts the chronology and runs the subroundHandlers loaded
func (chr *Chronology) StartRounds() {
	for chr.DoRun {
		time.Sleep(sleepTime)
		chr.StartRound()
	}
}

// StartRound calls the current subround, given by the current time or by the finished tasks in this round
func (chr *Chronology) StartRound() {
	subRoundId := chr.updateRound()

	chr.updateSelfSubroundIfNeeded(subRoundId)
}

func (chr *Chronology) updateSelfSubroundIfNeeded(subRoundId SubroundId) {
	if chr.SelfSubround() != subRoundId {
		return
	}

	sr := chr.LoadSubroundHandler(subRoundId)
	if sr == nil {
		return
	}

	if chr.Round().Index() < 0 {
		return
	}

	if !sr.DoWork(chr.ComputeSubRoundId, chr.IsCancelled) {
		return
	}

	if chr.IsCancelled() {
		return
	}

	chr.SetSelfSubround(sr.Next())
}

// updateRound updates Rounds and subrounds inside round depending of the current time and sync mode
func (chr *Chronology) updateRound() SubroundId {
	oldRoundIndex := chr.round.index
	oldTimeSubRound := chr.timeSubround

	currentTime := chr.syncTimer.CurrentTime(chr.clockOffset)
	chr.round.UpdateRound(chr.genesisTime, currentTime)
	chr.timeSubround = chr.GetSubroundFromDateTime(currentTime)

	if oldRoundIndex != chr.round.index {
		log.Info(fmt.Sprintf(
			"\n%s############################## ROUND %d BEGINS (%d) ##############################\n\n",
			chr.SyncTimer().FormattedCurrentTime(chr.ClockOffset()),
			chr.round.index,
			chr.SyncTimer().CurrentTime(chr.ClockOffset()).Unix()))

		chr.initRound()
	}

	if oldTimeSubRound != chr.timeSubround {
		sr := chr.LoadSubroundHandler(chr.timeSubround)
		if sr != nil {
			log.Info(fmt.Sprintf(
				"\n%s.................... SUBROUND %s BEGINS ....................\n\n",
				chr.SyncTimer().FormattedCurrentTime(chr.ClockOffset()),
				sr.Name(),
			))
		}
	}

	subRound := chr.SelfSubround()

	if chr.doSyncMode {
		subRound = chr.timeSubround
	}

	return subRound
}

// LoadSubroundHandler returns the implementation of SubroundHandler attached to the subround given
func (chr *Chronology) LoadSubroundHandler(subround SubroundId) SubroundHandler {
	chr.mutSubrounds.RLock()
	defer chr.mutSubrounds.RUnlock()

	index, ok := chr.subrounds[subround]

	if !ok || index < 0 || index >= len(chr.subroundHandlers) {
		return nil
	}

	return chr.subroundHandlers[index]
}

// GetSubround method returns current subround taking in consideration the current time
func (chr *Chronology) GetSubround() SubroundId {
	currentTime := chr.syncTimer.CurrentTime(chr.clockOffset)

	return chr.GetSubroundFromDateTime(currentTime)
}

// GetSubroundFromDateTime returns subround in the current round related to the time given
func (chr *Chronology) GetSubroundFromDateTime(timeStamp time.Time) SubroundId {
	chr.mutSubrounds.RLock()
	defer chr.mutSubrounds.RUnlock()

	delta := timeStamp.Sub(chr.round.timeStamp).Nanoseconds()

	if delta < 0 || delta > chr.round.timeDuration.Nanoseconds() {
		return -1
	}

	for i := 0; i < len(chr.subroundHandlers); i++ {
		if delta <= chr.subroundHandlers[i].EndTime() {
			return chr.subroundHandlers[i].Current()
		}
	}

	return -1
}

// GetFormattedTime method returns formatted current time
func (chr *Chronology) GetFormattedTime() string {
	return chr.syncTimer.FormattedCurrentTime(chr.clockOffset)
}

// RoundTimeStamp method returns time stamp of the current round
func (chr *Chronology) RoundTimeStamp() uint64 {
	return chr.RoundTimeStampFromIndex(chr.Round().Index())
}

// RoundTimeStampFromIndex method returns time stamp of a round from a given index
func (chr *Chronology) RoundTimeStampFromIndex(index int32) uint64 {
	return uint64(chr.genesisTime.Add(time.Duration(int64(index) * int64(chr.round.timeDuration))).Unix())
}

// Round returns the current round object
func (chr *Chronology) Round() *Round {
	return chr.round
}

// SelfSubround returns the subround, related to the finished tasks in the current round
func (chr *Chronology) SelfSubround() SubroundId {
	return chr.selfSubround
}

// SetSelfSubround set self subround depending of the finished tasks in the current round
func (chr *Chronology) SetSelfSubround(subRound SubroundId) {
	chr.selfSubround = subRound
}

// TimeSubround returns the subround, related to the current time
func (chr *Chronology) TimeSubround() SubroundId {
	return chr.timeSubround
}

// ClockOffset returns the current offset between local time and NTP
func (chr *Chronology) ClockOffset() time.Duration {
	return chr.clockOffset
}

// SetClockOffset set current offset
func (chr *Chronology) SetClockOffset(clockOffset time.Duration) {
	chr.clockOffset = clockOffset
}

// SyncTimer returns the current implementation of SynchTimer interface
func (chr *Chronology) SyncTimer() ntp.SyncTimer {
	return chr.syncTimer
}

// SubroundHandlers returns the array of subrounds loaded
func (chr *Chronology) SubroundHandlers() []SubroundHandler {
	return chr.subroundHandlers
}

// ComputeSubRoundId gets the current subround id from the current time
func (chr *Chronology) ComputeSubRoundId() SubroundId {
	return chr.GetSubroundFromDateTime(chr.SyncTimer().CurrentTime(chr.ClockOffset()))
}

// IsCancelled checks if this round is canceled
func (chr *Chronology) IsCancelled() bool {
	return chr.SelfSubround() == SubroundId(-1)
}
