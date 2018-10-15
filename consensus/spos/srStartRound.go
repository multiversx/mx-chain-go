package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRStartRound struct {
	doLog   bool
	endTime int64
	cns     *Consensus
}

func NewSRStartRound(doLog bool, endTime int64, cns *Consensus) SRStartRound {
	sr := SRStartRound{doLog: doLog, endTime: endTime, cns: cns}
	return sr
}

func (sr *SRStartRound) DoWork(chr *chronology.Chronology) bool {
	for chr.GetSelfSubround() != chronology.SR_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		switch sr.doStartRound(chr) {
		case R_None:
			continue
		case R_False:
			return false
		case R_True:
			return true
		default:
			return false
		}
	}

	sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 0: Aborded round %d in subround %s", chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SRStartRound) doStartRound(chr *chronology.Chronology) Response {
	leader, err := sr.cns.GetLeader()
	if err != nil {
		leader = "Unknown"
	}

	if leader == sr.cns.Self {
		leader += " (MY TURN)"
	}

	sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 0: Preparing for this round with leader %s ", leader))

	sr.cns.Block.ResetBlock()
	sr.cns.ResetRoundStatus()
	sr.cns.ResetValidationMap()

	return R_True
}

func (sr *SRStartRound) Current() chronology.Subround {
	return chronology.Subround(SR_START_ROUND)
}

func (sr *SRStartRound) Next() chronology.Subround {
	return chronology.Subround(SR_BLOCK)
}

func (sr *SRStartRound) EndTime() int64 {
	return int64(sr.endTime)
}

func (sr *SRStartRound) Name() string {
	return "<START_ROUND>"
}

func (sr *SRStartRound) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
