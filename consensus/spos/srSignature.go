package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRSignature struct {
	doLog             bool
	endTime           int64
	cns               *Consensus
	OnReceivedMessage func(*[]byte, *chronology.Chronology) bool
	OnSendMessage     func(chronology.Subround) bool
}

func NewSRSignature(doLog bool, endTime int64, cns *Consensus, onReceivedMessage func(*[]byte, *chronology.Chronology) bool, onSendMessage func(chronology.Subround) bool) *SRSignature {
	sr := SRSignature{doLog: doLog, endTime: endTime, cns: cns, OnReceivedMessage: onReceivedMessage, OnSendMessage: onSendMessage}
	return &sr
}

func (sr *SRSignature) DoWork(chr *chronology.Chronology) bool {
	for chr.GetSelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.doSignature(chr) {
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

	sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 5: Canceled round %d in subround %s", chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SRSignature) doSignature(chr *chronology.Chronology) Response {
	bActionDone := sr.OnSendMessage(chronology.Subround(srSignature))

	timeSubRound := chr.GetSubroundFromDateTime(chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(srSignature) {
		sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 5: Extended the "+sr.Name()+" subround. Got only %d from %d sigantures which are not enough", sr.cns.GetSignaturesCount(), len(sr.cns.ConsensusGroup)))
		sr.cns.RoundStatus.Signature = ssExtended
		return rTrue // Try to give a chance to this round if the necesary signatures will arrive later
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
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(srBlock), chronology.Subround(srSignature)); ok {
			sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 5: Received %d from %d signatures, which are matching with bitmap and are enough", n, len(sr.cns.ConsensusGroup)))
			return rTrue
		}
	}

	return rNone
}

func (sr *SRSignature) Current() chronology.Subround {
	return chronology.Subround(srSignature)
}

func (sr *SRSignature) Next() chronology.Subround {
	return chronology.Subround(srEndRound)
}

func (sr *SRSignature) EndTime() int64 {
	return int64(sr.endTime)
}

func (sr *SRSignature) Name() string {
	return "<SIGNATURE>"
}

func (sr *SRSignature) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
