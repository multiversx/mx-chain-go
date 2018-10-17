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
	SrUnknown Subround = math.MinInt64 // xzxzx
	// SrCanceled defines an aborded state of the round
	SrCanceled Subround = math.MinInt64 + 1 // xzxzxz
	// SrBeforeRound defines the state of the round before it's start
	SrBeforeRound Subround = math.MinInt64 + 2 // xzxzxz
	// SrAfterRound defines the state of the round after it's finish
	SrAfterRound Subround = math.MaxInt64
)

// Subrounder defines an interface which chronology deals with
type Subrounder interface {
	DoWork(*Chronology) bool
	Next() Subround
	Current() Subround
	EndTime() int64
	Name() string
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

	subrounders []Subrounder
	subrounds   map[Subround]int

	syncTime ntp.SyncTimer

	rounds int // only for statistic
}

// NewChronology defines a new Chronology object
func NewChronology(doLog bool, doSyncMode bool, round *Round, genesisTime time.Time, syncTime ntp.SyncTimer) *Chronology {
	chr := Chronology{}

	chr.doLog = doLog
	chr.DoRun = true
	chr.doSyncMode = doSyncMode

	chr.round = round
	chr.genesisTime = genesisTime

	chr.selfSubround = SrBeforeRound
	chr.timeSubround = SrBeforeRound
	chr.clockOffset = syncTime.ClockOffset()

	chr.subrounders = make([]Subrounder, 0)
	chr.subrounds = make(map[Subround]int)

	chr.syncTime = syncTime

	chr.rounds = 0

	return &chr
}

// initRound is call when a new round begins and do some init stuff
func (chr *Chronology) initRound() {
	chr.selfSubround = SrBeforeRound

	if len(chr.subrounders) > 0 {
		chr.selfSubround = chr.subrounders[0].Current()
	}

	chr.clockOffset = chr.syncTime.ClockOffset()
}

// AddSubrounder adds new Subrounder implementation to the chronology
func (chr *Chronology) AddSubrounder(subRounder Subrounder) {
	if chr.subrounds == nil {
		chr.subrounds = make(map[Subround]int)
	}

	chr.subrounds[subRounder.Current()] = len(chr.subrounders)
	chr.subrounders = append(chr.subrounders, subRounder)
}

// StartRounds actually starts the chronology and runs the subrounders loaded
func (chr *Chronology) StartRounds() {
	for chr.DoRun {
		time.Sleep(sleepTime * time.Millisecond)
		chr.startRound()
	}
}

// startRound calls the current subrounder, given by the current time or by the finished tasks in this round
func (chr *Chronology) startRound() {
	subRound := chr.updateRound()

	if chr.selfSubround == subRound {
		sr := chr.loadSubrounder(subRound)
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
		sr := chr.loadSubrounder(chr.timeSubround)
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

// loadSubrounder returns the implementation of Subrounder attached to the subround given
func (chr *Chronology) loadSubrounder(subround Subround) Subrounder {
	index := chr.subrounds[subround]
	if index < 0 || index >= len(chr.subrounders) {
		return nil
	}

	return chr.subrounders[index]
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

	for i := 0; i < len(chr.subrounders); i++ {
		if delta <= chr.subrounders[i].EndTime() {
			return chr.subrounders[i].Current()
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
