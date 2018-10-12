package chronology

import (
	"fmt"
	"math"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
)

const SLEEP_TIME = 0

type Subround int

const (
	SR_UNKNOWN               = math.MinInt32
	SR_ABORDED               = math.MinInt32 + 1
	SR_BEFORE_ROUND Subround = math.MinInt32 + 2
	SR_AFTER_ROUND  Subround = math.MaxInt32
)

type Subrounder interface {
	DoWork() bool
	Next() Subround
	Current() Subround
	EndTime() int64
	Name() string
}

type Chronology struct {
	doLog      bool
	DoRun      bool
	doSyncMode bool

	round       *Round
	genesisTime time.Time

	selfSubRound Subround
	timeSubRound Subround
	clockOffset  time.Duration

	subRounders []Subrounder
	subRounds   map[Subround]int

	syncTime ntp.SyncTimer
	rounds   int // only for statistic
}

func NewChronology(doLog bool, doSyncMode bool, round *Round, genesisTime time.Time, syncTime ntp.SyncTimer) Chronology {
	chr := Chronology{}

	chr.doLog = doLog
	chr.DoRun = true
	chr.doSyncMode = doSyncMode
	chr.round = round
	chr.genesisTime = genesisTime
	chr.selfSubRound = SR_BEFORE_ROUND
	chr.timeSubRound = SR_BEFORE_ROUND
	chr.clockOffset = syncTime.GetClockOffset()
	chr.syncTime = syncTime
	chr.subRounders = make([]Subrounder, 0)
	chr.subRounds = make(map[Subround]int)

	return chr
}

func (chr *Chronology) initRound() {
	chr.selfSubRound = SR_BEFORE_ROUND

	if len(chr.subRounders) > 0 {
		chr.selfSubRound = chr.subRounders[0].Current()
	}

	chr.clockOffset = chr.syncTime.GetClockOffset()
}

func (chr *Chronology) AddSubRounder(subRounder Subrounder) {
	if chr.subRounds == nil {
		chr.subRounds = make(map[Subround]int)
	}

	chr.subRounds[subRounder.Current()] = len(chr.subRounders)
	chr.subRounders = append(chr.subRounders, subRounder)
}

func (chr *Chronology) StartRounds() {
	for chr.DoRun {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		chr.startRound()
	}
}

func (chr *Chronology) startRound() {
	subRound := chr.UpdateRound()

	if chr.selfSubRound == subRound {
		sr := chr.LoadSubRounder(subRound)
		if sr != nil {
			if sr.DoWork() {
				chr.selfSubRound = sr.Next()
			}
		}
	}
}

func (chr *Chronology) UpdateRound() Subround {
	oldRoundIndex := chr.round.index
	oldTimeSubRound := chr.timeSubRound

	currentTime := chr.syncTime.GetCurrentTime(chr.clockOffset)
	chr.round.UpdateRound(chr.genesisTime, currentTime)
	chr.timeSubRound = chr.GetSubRoundFromDateTime(currentTime)

	if oldRoundIndex != chr.round.index {
		chr.rounds++ // only for statistic

		//leader, err := chr.GetLeader()
		//if err != nil {
		//	leader = "Unknown"
		//}
		//
		//if leader == c.Validators.Self {
		//	leader += " (MY TURN)"
		//}

		//chr.Log(fmt.Sprintf("\n"+chr.GetFormatedCurrentTime()+"############################## ROUND %d BEGINS WITH LEADER  %s  ##############################\n", chr.round.Index, leader))
		chr.Log(fmt.Sprintf("\n"+chr.GetFormatedCurrentTime()+"############################## ROUND %d BEGINS ##############################\n", chr.round.index))
		chr.initRound()
	}

	if oldTimeSubRound != chr.timeSubRound {
		sr := chr.LoadSubRounder(chr.timeSubRound)
		if sr != nil {
			chr.Log(fmt.Sprintf("\n" + chr.GetFormatedCurrentTime() + ".................... SUBROUND " + sr.Name() + " BEGINS ....................\n"))
		}
	}

	subRound := chr.selfSubRound

	if chr.doSyncMode {
		subRound = chr.timeSubRound
	}

	return subRound
}

func (chr *Chronology) LoadSubRounder(subRound Subround) Subrounder {
	index := chr.subRounds[subRound]
	if index < 0 || index >= len(chr.subRounders) {
		return nil
	}

	return chr.subRounders[index]
}

func (chr *Chronology) GetSubRoundFromDateTime(timeStamp time.Time) Subround {

	delta := timeStamp.Sub(chr.round.timeStamp).Nanoseconds()

	if delta < 0 {
		return SR_BEFORE_ROUND
	}

	if delta > chr.round.timeDuration.Nanoseconds() {
		return SR_AFTER_ROUND
	}

	for i := 0; i < len(chr.subRounders); i++ {
		if delta <= chr.subRounders[i].EndTime() {
			return chr.subRounders[i].Current()
		}
	}

	return SR_UNKNOWN
}

func (chr *Chronology) Log(message string) {
	if chr.doLog {
		fmt.Printf(message + "\n")
	}
}

func (chr *Chronology) GetRoundIndex() int {
	return chr.round.index
}

func (chr *Chronology) GetSelfSubRound() Subround {
	return chr.selfSubRound
}

func (chr *Chronology) SetSelfSubRound(subRound Subround) {
	chr.selfSubRound = subRound
}

func (chr *Chronology) GetTimeSubRound() Subround {
	return chr.timeSubRound
}

func (chr *Chronology) GetClockOffset() time.Duration {
	return chr.clockOffset
}

func (chr *Chronology) GetSyncTimer() ntp.SyncTimer {
	return chr.syncTime
}

func (chr *Chronology) GetFormatedCurrentTime() string {
	return chr.syncTime.GetFormatedCurrentTime(chr.clockOffset)
}

func (chr *Chronology) GetCurrentTime() time.Time {
	return chr.syncTime.GetCurrentTime(chr.clockOffset)
}
