package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRComitmentHash struct {
	doLog             bool
	endTime           int64
	cns               *Consensus
	OnReceivedMessage func(*[]byte, *chronology.Chronology) bool
	OnSendMessage     func(chronology.Subround) bool
}

func NewSRComitmentHash(doLog bool, endTime int64, cns *Consensus, onReceivedMessage func(*[]byte, *chronology.Chronology) bool, onSendMessage func(chronology.Subround) bool) *SRComitmentHash {
	sr := SRComitmentHash{doLog: doLog, endTime: endTime, cns: cns, OnReceivedMessage: onReceivedMessage, OnSendMessage: onSendMessage}
	return &sr
}

func (sr *SRComitmentHash) DoWork(chr *chronology.Chronology) bool {
	for chr.GetSelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.doComitmentHash(chr) {
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

	sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 2: Canceled round %d in subround %s", chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SRComitmentHash) doComitmentHash(chr *chronology.Chronology) Response {
	bActionDone := sr.OnSendMessage(chronology.Subround(srComitmentHash))

	timeSubRound := chr.GetSubroundFromDateTime(chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(srComitmentHash) {
		if sr.cns.GetComitmentHashesCount() < sr.cns.Threshold.ComitmentHash {
			sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 2: Extended the "+sr.Name()+" subround. Got only %d from %d commitment hashes which are not enough", sr.cns.GetComitmentHashesCount(), len(sr.cns.ConsensusGroup)))
		} else {
			sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime() + "Step 2: Extended the " + sr.Name() + " subround"))
		}
		sr.cns.RoundStatus.ComitmentHash = ssExtended
		return rTrue // Try to give a chance to this round if the necesary comitment hashes will arrive later
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
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(srBlock), chronology.Subround(srComitmentHash)); ok {
			if n == len(sr.cns.ConsensusGroup) {
				sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 2: Received all (%d from %d) comitment hashes", n, len(sr.cns.ConsensusGroup)))
			} else {
				sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 2: Received %d from %d comitment hashes, which are enough", n, len(sr.cns.ConsensusGroup)))
			}
			return rTrue
		}
	}

	return rNone
}

func (sr *SRComitmentHash) Current() chronology.Subround {
	return chronology.Subround(srComitmentHash)
}

func (sr *SRComitmentHash) Next() chronology.Subround {
	return chronology.Subround(srBitmap)
}

func (sr *SRComitmentHash) EndTime() int64 {
	return int64(sr.endTime)
}

func (sr *SRComitmentHash) Name() string {
	return "<COMITMENT_HASH>"
}

func (sr *SRComitmentHash) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
