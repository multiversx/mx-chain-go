package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRStartRound struct {
	doLog             bool
	endTime           int64
	cns               *Consensus
	OnReceivedMessage func(*[]byte, *chronology.Chronology) bool
	OnSendMessage     func(chronology.Subround) bool
}

func NewSRStartRound(doLog bool, endTime int64, cns *Consensus, onReceivedMessage func(*[]byte, *chronology.Chronology) bool, onSendMessage func(chronology.Subround) bool) *SRStartRound {
	sr := SRStartRound{doLog: doLog, endTime: endTime, cns: cns, OnReceivedMessage: onReceivedMessage, OnSendMessage: onSendMessage}
	return &sr
}

func (sr *SRStartRound) DoWork(chr *chronology.Chronology) bool {
	for chr.SelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.doStartRound(chr) {
		case rNone:
			continue
		case rFalse:
			return false
		case rTrue:
			return true
		default:
			return false
		}
	}

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 0: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

func (sr *SRStartRound) doStartRound(chr *chronology.Chronology) Response {
	leader, err := sr.cns.GetLeader()

	if err != nil {
		return rNone
	}

	if leader == sr.cns.Self {
		leader += " (MY TURN)"
	}

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 0: Preparing for this round with leader %s ", leader))

	//sr.cns.block.ResetBlock()
	sr.cns.ResetRoundStatus()
	sr.cns.ResetValidationMap()

	return rTrue
}

func (sr *SRStartRound) Current() chronology.Subround {
	return chronology.Subround(SrStartRound)
}

func (sr *SRStartRound) Next() chronology.Subround {
	return chronology.Subround(SrBlock)
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
