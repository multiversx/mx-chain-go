package chronology

import (
	"fmt"
	"math"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
)

// sleepTime defines the time in milliseconds between each iteration made in StartRounds method
const sleepTime = 5

// Subround defines the type used to reffer the current subround
type Subround int

const (
	// SrUnknown defines an unknown state of the round
	SrUnknown Subround = math.MinInt64
	// SrCanceled defines an canceled state of the round
	SrCanceled Subround = math.MinInt64 + 1
	// SrBeforeRound defines the state of the round before it's start
	SrBeforeRound Subround = math.MinInt64 + 2
	// SrAfterRound defines the state of the round after it's finish
	SrAfterRound Subround = math.MaxInt64
)

// SubroundHandler defines the actions that can be handled in a sub-round
type SubroundHandler interface {
	DoWork(*Chronology) bool // DoWork implements of the subround's job
	Next() Subround          // Next returns the ID of the next subround
	Current() Subround       // Current returns the ID of the current subround
	EndTime() int64          // EndTime returns the top limit time, in the round time, of the current subround
	Name() string            // Name returns the name of the current round
}

// Chronology defines the data needed by the chronology
type Chronology struct {
	doLog      bool
	DoRun      bool
	doSyncMode bool

	round       *Round
	genesisTime time.Time

	selfSubround Subround
	timeSubround Subround
	clockOffset  time.Duration

	subroundHandlers []SubroundHandler
	subrounds        map[Subround]int

	syncTime ntp.SyncTimer

	rounds int // only for statistic
}

// NewChronology defines a new Chronology object
func NewChronology(doLog bool, doSyncMode bool, round *Round, genesisTime time.Time, syncTime ntp.SyncTimer) *Chronology {
	chr := Chronology{doLog: doLog, doSyncMode: doSyncMode, round: round, genesisTime: genesisTime, syncTime: syncTime}

	chr.DoRun = true

	chr.selfSubround = SrBeforeRound
	chr.timeSubround = SrBeforeRound
	chr.clockOffset = syncTime.ClockOffset()

	chr.subroundHandlers = make([]SubroundHandler, 0)
	chr.subrounds = make(map[Subround]int)

	chr.rounds = 0

	return &chr
}

// initRound is call when a new round begins and do the necesary initialization
func (chr *Chronology) initRound() {
	chr.selfSubround = SrBeforeRound

	if len(chr.subroundHandlers) > 0 {
		chr.selfSubround = chr.subroundHandlers[0].Current()
	}

	chr.clockOffset = chr.syncTime.ClockOffset()
}

// AddSubround adds new SubroundHandler implementation to the chronology
func (chr *Chronology) AddSubround(subroundHandler SubroundHandler) {
	if chr.subrounds == nil {
		chr.subrounds = make(map[Subround]int)
	}

	chr.subrounds[subroundHandler.Current()] = len(chr.subroundHandlers)
	chr.subroundHandlers = append(chr.subroundHandlers, subroundHandler)
}

// StartRounds actually starts the chronology and runs the subroundHandlers loaded
func (chr *Chronology) StartRounds() {
	for chr.DoRun {
		time.Sleep(sleepTime * time.Millisecond)
		chr.StartRound()
	}
}

// StartRound calls the current subround, given by the current time or by the finished tasks in this round
func (chr *Chronology) StartRound() {
	subRound := chr.updateRound()

	if chr.selfSubround == subRound {
		sr := chr.LoadSubroundHandler(subRound)
		if sr != nil {
			if sr.DoWork(chr) {
				chr.selfSubround = sr.Next()
			}
		}
	}
}

// updateRound updates rounds and subrounds inside round depending of the current time and sync mode
func (chr *Chronology) updateRound() Subround {
	oldRoundIndex := chr.round.index
	oldTimeSubRound := chr.timeSubround

	currentTime := chr.syncTime.CurrentTime(chr.clockOffset)
	chr.round.UpdateRound(chr.genesisTime, currentTime)
	chr.timeSubround = chr.GetSubroundFromDateTime(currentTime)

	if oldRoundIndex != chr.round.index {
		chr.rounds++ // only for statistic
		chr.log(fmt.Sprintf("\n"+chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"############################## ROUND %d BEGINS ##############################\n", chr.round.index))
		chr.initRound()
	}

	if oldTimeSubRound != chr.timeSubround {
		sr := chr.LoadSubroundHandler(chr.timeSubround)
		if sr != nil {
			chr.log(fmt.Sprintf("\n" + chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) + ".................... SUBROUND " + sr.Name() + " BEGINS ....................\n"))
		}
	}

	subRound := chr.selfSubround

	if chr.doSyncMode {
		subRound = chr.timeSubround
	}

	return subRound
}

// LoadSubroundHandler returns the implementation of SubroundHandler attached to the subround given
func (chr *Chronology) LoadSubroundHandler(subround Subround) SubroundHandler {
	index := chr.subrounds[subround]
	if index < 0 || index >= len(chr.subroundHandlers) {
		return nil
	}

	return chr.subroundHandlers[index]
}

// GetSubroundFromDateTime returns subround in the current round related to the time given
func (chr *Chronology) GetSubroundFromDateTime(timeStamp time.Time) Subround {

	delta := timeStamp.Sub(chr.round.timeStamp).Nanoseconds()

	if delta < 0 {
		return SrBeforeRound
	}

	if delta > chr.round.timeDuration.Nanoseconds() {
		return SrAfterRound
	}

	for i := 0; i < len(chr.subroundHandlers); i++ {
		if delta <= chr.subroundHandlers[i].EndTime() {
			return chr.subroundHandlers[i].Current()
		}
	}

	return SrUnknown
}

// log do logs of the chronology if doLog variable is set on true
func (chr *Chronology) log(message string) {
	if chr.doLog {
		fmt.Printf(message + "\n")
	}
}

// Round returns the current round object
func (chr *Chronology) Round() *Round {
	return chr.round
}

// SelfSubround returns the subround, related to the finished tasks in the current round
func (chr *Chronology) SelfSubround() Subround {
	return chr.selfSubround
}

// SetSelfSubround set self subround depending of the finished tasks in the current round
func (chr *Chronology) SetSelfSubround(subRound Subround) {
	chr.selfSubround = subRound
}

// TimeSubround returns the subround, related to the current time
func (chr *Chronology) TimeSubround() Subround {
	return chr.timeSubround
}

// clockOffset returns the current offset between local time and NTP
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
