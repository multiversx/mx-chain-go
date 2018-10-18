package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRBlock struct {
	doLog           bool
	endTime         int64
	cns             *Consensus
	OnReceivedBlock func(*[]byte, *chronology.Chronology) bool
	OnSendBlock     func(chronology.Subround) bool
}

func NewSRBlock(doLog bool, endTime int64, cns *Consensus, onReceivedBlock func(*[]byte, *chronology.Chronology) bool, onSendBlock func(chronology.Subround) bool) *SRBlock {
	sr := SRBlock{doLog: doLog, endTime: endTime, cns: cns, OnReceivedBlock: onReceivedBlock, OnSendBlock: onSendBlock}
	return &sr
}

func (sr *SRBlock) DoWork(chr *chronology.Chronology) bool {
	for chr.SelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.doBlock(chr) {
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

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 1: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

func (sr *SRBlock) doBlock(chr *chronology.Chronology) Response {
	sr.cns.SetSentMessage(sr.OnSendBlock(chronology.Subround(SrBlock)))

	if sr.cns.SentMessage() {
		sr.cns.SetSentMessage(false)
		if ok, _ := sr.cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrBlock)); ok {
			return rTrue
		}
	}

	timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

	if timeSubRound > chronology.Subround(SrBlock) {
		sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) + "Step 1: Extended the " + sr.Name() + " subround"))
		sr.cns.RoundStatus.Block = SsExtended
		return rTrue // Try to give a chance to this round if the block from leader will arrive later
	}

	if sr.cns.ReceivedMessage() {
		sr.cns.SetReceivedMessage(false)
		if ok, _ := sr.cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrBlock)); ok {
			sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) + "Step 1: Synchronized block"))
			return rTrue
		}
	}

	return rNone
}

func (sr *SRBlock) Current() chronology.Subround {
	return chronology.Subround(SrBlock)
}

func (sr *SRBlock) Next() chronology.Subround {
	return chronology.Subround(SrComitmentHash)
}

func (sr *SRBlock) EndTime() int64 {
	return int64(sr.endTime)
}

func (sr *SRBlock) Name() string {
	return "<BLOCK>"
}

func (sr *SRBlock) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
