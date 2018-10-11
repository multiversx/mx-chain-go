package chronology

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
)

const SLEEP_TIME = 0

const (
	RS_BEFORE_ROUND = -1
	RS_AFTER_ROUND  = -2
	RS_UNKNOWN      = -3
)

type SubRounder interface {
	DoWork(*Chronology) bool
	Next() int
	EndTime() int64
	Name() string
}

type Chronology struct {
	doLog      bool
	DoRun      bool
	doSyncMode bool

	round       *Round
	genesisTime time.Time

	selfRoundState int
	timeRoundState int
	clockOffset    time.Duration

	subRounders []SubRounder

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
	chr.selfRoundState = 0
	chr.timeRoundState = RS_BEFORE_ROUND
	chr.clockOffset = syncTime.GetClockOffset()
	chr.syncTime = syncTime
	chr.subRounders = make([]SubRounder, 0)

	return chr
}

func (chr *Chronology) initRound() {
	chr.selfRoundState = 0
	chr.clockOffset = chr.syncTime.GetClockOffset()
}

func (chr *Chronology) AddSubRounder(subRounder SubRounder) {
	chr.subRounders = append(chr.subRounders, subRounder)
}

func (chr *Chronology) StartRounds() {
	for chr.DoRun {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		chr.startRound()
	}
}

func (chr *Chronology) startRound() {
	roundState := chr.UpdateRound()

	sr := chr.LoadSubRounder(roundState)

	if chr.selfRoundState == roundState {
		if sr != nil {
			if sr.DoWork(chr) {
				chr.selfRoundState = sr.Next()
			}
		}
	}
}

func (chr *Chronology) UpdateRound() int {
	oldRoundIndex := chr.round.index
	oldTimeRoundState := chr.timeRoundState

	currentTime := chr.syncTime.GetCurrentTime(chr.clockOffset)
	chr.round.UpdateRound(chr.genesisTime, currentTime)
	chr.timeRoundState = chr.GetRoundStateFromDateTime(currentTime)

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

		//chr.Log(fmt.Sprintf("\n"+chr.syncTime.GetFormatedCurrentTimeWithOffset(chr.clockOffset)+"############################## ROUND %d BEGINS WITH LEADER  %s  ##############################\n", chr.round.Index, leader))
		chr.Log(fmt.Sprintf("\n"+chr.syncTime.GetFormatedCurrentTime(chr.clockOffset)+"############################## ROUND %d BEGINS ##############################\n", chr.round.index))
		chr.initRound()
	}

	if oldTimeRoundState != chr.timeRoundState {
		sr := chr.LoadSubRounder(chr.timeRoundState)
		if sr != nil {
			chr.Log(fmt.Sprintf("\n" + chr.syncTime.GetFormatedCurrentTime(chr.clockOffset) + ".................... SUBROUND " + sr.Name() + " BEGINS ....................\n"))
		}
	}

	roundState := chr.selfRoundState

	if chr.doSyncMode {
		roundState = chr.timeRoundState
	}

	return roundState
}

func (chr *Chronology) LoadSubRounder(roundState int) SubRounder {
	if roundState < 0 || roundState >= len(chr.subRounders) {
		return nil
	}

	return chr.subRounders[roundState]
}

func (chr *Chronology) GetRoundStateFromDateTime(timeStamp time.Time) int {

	delta := timeStamp.Sub(chr.round.timeStamp).Nanoseconds()

	if delta < 0 {
		return RS_BEFORE_ROUND
	}

	if delta > chr.round.timeDuration.Nanoseconds() {
		return RS_AFTER_ROUND
	}

	for i := 0; i < len(chr.subRounders); i++ {
		if delta <= chr.subRounders[i].EndTime() {
			return i
		}
	}

	return RS_UNKNOWN
}

func (chr *Chronology) Log(message string) {
	if chr.doLog {
		fmt.Printf(message + "\n")
	}
}

//func (chr *Chronology) GetRoundIndex() int {
//	return chr.round.index
//}
//
//func (chr *Chronology) GetSelfRoundState() int {
//	return chr.selfRoundState
//}
//
//func (chr *Chronology) GetTimeRoundState() int {
//	return chr.timeRoundState
//}
