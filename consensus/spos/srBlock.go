package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRBlock struct {
	doLog             bool
	endTime           int64
	cns               *Consensus
	OnReceivedMessage func(*[]byte, *chronology.Chronology) bool
	OnSendMessage     func(chronology.Subround) bool
}

func NewSRBlock(doLog bool, endTime int64, cns *Consensus, onReceivedMessage func(*[]byte, *chronology.Chronology) bool, onSendMessage func(chronology.Subround) bool) *SRBlock {
	sr := SRBlock{doLog: doLog, endTime: endTime, cns: cns, OnReceivedMessage: onReceivedMessage, OnSendMessage: onSendMessage}
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
	bActionDone := sr.OnSendMessage(chronology.Subround(SrBlock))

	if bActionDone {
		bActionDone = false
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

	select {
	case rcvMsg := <-sr.cns.ChRcvMsg:
		if sr.OnReceivedMessage(&rcvMsg, chr) {
			bActionDone = true
		}
	default:
	}

	if bActionDone {
		bActionDone = false
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
