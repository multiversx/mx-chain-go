package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SRBlock defines the data needed by the block subround
type SRBlock struct {
	doLog       bool
	endTime     int64
	Cns         *Consensus
	OnSendBlock func() bool
}

// NewSRBlock creates a new SRBlock object
func NewSRBlock(doLog bool, endTime int64, cns *Consensus, onSendBlock func() bool) *SRBlock {
	sr := SRBlock{doLog: doLog, endTime: endTime, Cns: cns, OnSendBlock: onSendBlock}
	return &sr
}

// DoWork method calls repeatedly DoBlock method, which is in charge to do the job of this subround, until RTrue or RFalse is return
// or until this subround is put in the canceled mode
func (sr *SRBlock) DoWork(chr *chronology.Chronology) bool {
	for chr.SelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.DoBlock(chr) {
		case RNone:
			continue
		case RFalse:
			return false
		case RTrue:
			return true
		default:
			return false
		}
	}

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 1: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

// DoBlock method actually do the job of this subround. First, it tries to send the block (if this node is in charge with this action)
// and than if the block was succesfully sent or if meantime a new message was received, it will check the consensus again.
// If the upper time limit of this subround is reached, it's state is set to extended and the chronology will advance to the next subround
func (sr *SRBlock) DoBlock(chr *chronology.Chronology) Response {
	sr.Cns.SetSentMessage(sr.OnSendBlock())

	if sr.Cns.SentMessage() {
		sr.Cns.SetSentMessage(false)
		if ok, _ := sr.Cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrBlock)); ok {
			return RTrue
		}
	}

	timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

	if timeSubRound > chronology.Subround(SrBlock) {
		sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) + "Step 1: Extended the " + sr.Name() + " subround"))
		sr.Cns.RoundStatus.Block = SsExtended
		return RTrue // Try to give a chance to this round if the block from leader will arrive later
	}

	if sr.Cns.ReceivedMessage() {
		sr.Cns.SetReceivedMessage(false)
		if ok, _ := sr.Cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrBlock)); ok {
			sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) + "Step 1: Synchronized block"))
			return RTrue
		}
	}

	return RNone
}

// Current method returns the ID of this subround
func (sr *SRBlock) Current() chronology.Subround {
	return chronology.Subround(SrBlock)
}

// Next method returns the ID of the next subround
func (sr *SRBlock) Next() chronology.Subround {
	return chronology.Subround(SrComitmentHash)
}

// EndTime method returns the upper time limit of this subround
func (sr *SRBlock) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of this subround
func (sr *SRBlock) Name() string {
	return "<BLOCK>"
}

// Log method prints info about this subrond (if doLog is true)
func (sr *SRBlock) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
