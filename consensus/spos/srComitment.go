package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRComitment struct {
	doLog             bool
	endTime           int64
	cns               *Consensus
	OnReceivedMessage func(*[]byte, *chronology.Chronology) bool
	OnSendMessage     func(chronology.Subround) bool
}

func NewSRComitment(doLog bool, endTime int64, cns *Consensus, onReceivedMessage func(*[]byte, *chronology.Chronology) bool, onSendMessage func(chronology.Subround) bool) *SRComitment {
	sr := SRComitment{doLog: doLog, endTime: endTime, cns: cns, OnReceivedMessage: onReceivedMessage, OnSendMessage: onSendMessage}
	return &sr
}

func (sr *SRComitment) DoWork(chr *chronology.Chronology) bool {
	for chr.GetSelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.doComitment(chr) {
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

	sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 4: Canceled round %d in subround %s", chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SRComitment) doComitment(chr *chronology.Chronology) Response {
	bActionDone := sr.OnSendMessage(chronology.Subround(srComitment))

	timeSubRound := chr.GetSubroundFromDateTime(chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(srComitment) {
		sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 4: Extended the "+sr.Name()+" subround. Got only %d from %d commitments which are not enough", sr.cns.GetComitmentsCount(), len(sr.cns.ConsensusGroup)))
		sr.cns.RoundStatus.Comitment = ssExtended
		return rTrue // Try to give a chance to this round if the necesary comitments will arrive later
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
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(srBlock), chronology.Subround(srComitment)); ok {
			sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 4: Received %d from %d comitments, which are matching with bitmap and are enough", n, len(sr.cns.ConsensusGroup)))
			return rTrue
		}
	}

	return rNone
}

func (sr *SRComitment) Current() chronology.Subround {
	return chronology.Subround(srComitment)
}

func (sr *SRComitment) Next() chronology.Subround {
	return chronology.Subround(srSignature)
}

func (sr *SRComitment) EndTime() int64 {
	return int64(sr.endTime)
}

func (sr *SRComitment) Name() string {
	return "<COMITMENT>"
}

func (sr *SRComitment) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
