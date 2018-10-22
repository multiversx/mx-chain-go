package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SRStartRound defines the data needed by the start round subround
type SRStartRound struct {
	doLog                bool
	endTime              int64
	cns                  *Consensus
	OnReceivedStartRound func(*[]byte, *chronology.Chronology) bool
	OnSendStartRound     func() bool
}

// NewSRStartRound creates a new SRStartRound object
func NewSRStartRound(doLog bool, endTime int64, cns *Consensus, onReceivedStartRound func(*[]byte, *chronology.Chronology) bool, onSendStartRound func() bool) *SRStartRound {
	sr := SRStartRound{doLog: doLog, endTime: endTime, cns: cns, OnReceivedStartRound: onReceivedStartRound, OnSendStartRound: onSendStartRound}
	return &sr
}

// DoWork method calls repeatedly doStartRound method, which is in charge to do the job of this subround, until rTrue or rFalse is return
// or until this subround is put in the canceled mode
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

// doStartRound method actually do the initialization of the new round
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

// Current method returns the ID of this subround
func (sr *SRStartRound) Current() chronology.Subround {
	return chronology.Subround(SrStartRound)
}

// Next method returns the ID of the next subround
func (sr *SRStartRound) Next() chronology.Subround {
	return chronology.Subround(SrBlock)
}

// EndTime method returns the upper time limit of this subround
func (sr *SRStartRound) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of this subround
func (sr *SRStartRound) Name() string {
	return "<START_ROUND>"
}

// Log method prints info about this subrond (if doLog is true)
func (sr *SRStartRound) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
