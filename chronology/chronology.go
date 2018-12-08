package chronology

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
)

// sleepTime defines the time in milliseconds between each iteration made in StartRounds method
const sleepTime = time.Duration(5 * time.Millisecond)

// SubroundId defines the type used to refer the current subround
type SubroundId int

// SubroundHandler defines the actions that can be handled in a sub-round
type SubroundHandler interface {
	DoWork(func() SubroundId, func() bool) bool // DoWork implements of the subround's job
	Next() SubroundId                           // Next returns the ID of the next subround
	Current() SubroundId                        // Current returns the ID of the current subround
	EndTime() int64                             // EndTime returns the top limit time, in the round time, of the current subround
	Name() string                               // Name returns the name of the current round
	Check() bool
}

// Chronology defines the data needed by the chronology
type Chronology struct {
	log        bool
	DoRun      bool
	doSyncMode bool

	round       *Round
	genesisTime time.Time

	selfSubround SubroundId
	timeSubround SubroundId
	clockOffset  time.Duration

	subroundHandlers []SubroundHandler
	subrounds        map[SubroundId]int

	syncTime ntp.SyncTimer

	mut sync.RWMutex
}

// NewChronology defines a new Chr object
func NewChronology(
	log bool,
	doSyncMode bool,
	round *Round,
	genesisTime time.Time,
	syncTime ntp.SyncTimer) *Chronology {

	chr := Chronology{
		log:         log,
		doSyncMode:  doSyncMode,
		round:       round,
		genesisTime: genesisTime,
		syncTime:    syncTime}

	chr.DoRun = true

	chr.SetSelfSubround(-1)
	chr.timeSubround = -1
	chr.clockOffset = syncTime.ClockOffset()

	chr.subroundHandlers = make([]SubroundHandler, 0)
	chr.subrounds = make(map[SubroundId]int)

	return &chr
}

// initRound is called when a new round begins and do the necesary initialization
func (chr *Chronology) initRound() {
	chr.SetSelfSubround(-1)

	if len(chr.subroundHandlers) > 0 {
		chr.SetSelfSubround(chr.subroundHandlers[0].Current())
	}

	chr.clockOffset = chr.syncTime.ClockOffset()
}

// AddSubround adds new SubroundHandler implementation to the chronology
func (chr *Chronology) AddSubround(subroundHandler SubroundHandler) {
	if chr.subrounds == nil {
		chr.subrounds = make(map[SubroundId]int)
	}

	chr.subrounds[subroundHandler.Current()] = len(chr.subroundHandlers)
	chr.subroundHandlers = append(chr.subroundHandlers, subroundHandler)
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
	subRound := chr.updateRound()

	if chr.SelfSubround() == subRound {
		sr := chr.LoadSubroundHandler(subRound)
		if sr != nil {
			if sr.DoWork(chr.computeSubRoundId, chr.isCancelled) {
				chr.SetSelfSubround(sr.Next())
			}
		}
	}
}

// updateRound updates Rounds and subrounds inside round depending of the current time and sync mode
func (chr *Chronology) updateRound() SubroundId {
	oldRoundIndex := chr.round.index
	oldTimeSubRound := chr.timeSubround

	currentTime := chr.syncTime.CurrentTime(chr.clockOffset)
	chr.round.UpdateRound(chr.genesisTime, currentTime)
	chr.timeSubround = chr.GetSubroundFromDateTime(currentTime)

	if oldRoundIndex != chr.round.index {
		chr.Log(fmt.Sprintf("\n"+chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+
			"############################## ROUND %d BEGINS ##############################\n", chr.round.index))
		chr.initRound()
	}

	if oldTimeSubRound != chr.timeSubround {
		sr := chr.LoadSubroundHandler(chr.timeSubround)
		if sr != nil {
			chr.Log(fmt.Sprintf("\n" + chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) +
				".................... SUBROUND " + sr.Name() + " BEGINS ....................\n"))
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
	index, ok := chr.subrounds[subround]
	if !ok || index < 0 || index >= len(chr.subroundHandlers) {
		return nil
	}

	return chr.subroundHandlers[index]
}

// GetSubroundFromDateTime returns subround in the current round related to the time given
func (chr *Chronology) GetSubroundFromDateTime(timeStamp time.Time) SubroundId {
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

// Log do logs of the chronology if log variable is set on true
func (chr *Chronology) Log(message string) {
	if chr.log {
		fmt.Printf(message + "\n")
	}
}

// Round returns the current round object
func (chr *Chronology) Round() *Round {
	return chr.round
}

// SelfSubround returns the subround, related to the finished tasks in the current round
func (chr *Chronology) SelfSubround() SubroundId {
	chr.mut.RLock()
	defer chr.mut.RUnlock()

	return chr.selfSubround
}

// SetSelfSubround set self subround depending of the finished tasks in the current round
func (chr *Chronology) SetSelfSubround(subRound SubroundId) {
	chr.mut.Lock()
	defer chr.mut.Unlock()

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

// SyncTime returns the current implementation of SynchTimer interface
func (chr *Chronology) SyncTime() ntp.SyncTimer {
	return chr.syncTime
}

// SubroundHandlers returns the array of subrounds loaded
func (chr *Chronology) SubroundHandlers() []SubroundHandler {
	return chr.subroundHandlers
}

func (chr *Chronology) computeSubRoundId() SubroundId {
	return chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))
}

func (chr *Chronology) isCancelled() bool {
	return chr.SelfSubround() == SubroundId(-1)
}
