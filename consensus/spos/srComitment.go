package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRComitment struct {
	doLog   bool
	endTime int64
	cns     *Consensus
}

func NewSRComitment(doLog bool, endTime int64, cns *Consensus) SRComitment {
	sr := SRComitment{doLog: doLog, endTime: endTime, cns: cns}
	return sr
}

func (sr *SRComitment) DoWork(chr *chronology.Chronology) bool {
	for chr.GetSelfSubround() != chronology.SR_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		switch sr.doComitment(chr) {
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

	sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 4: Aborded round %d in subround %s", chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SRComitment) doComitment(chr *chronology.Chronology) Response {
	bActionDone := sr.cns.SendMessage(chronology.Subround(SR_COMITMENT))

	timeSubRound := chr.GetSubroundFromDateTime(chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(SR_COMITMENT) {
		sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 4: Extended the "+sr.Name()+" subround. Got only %d from %d commitments which are not enough", sr.cns.GetComitmentsCount(), len(sr.cns.ConsensusGroup)))
		sr.cns.RoundStatus.Comitment = SS_EXTENDED
		return R_True // Try to give a chance to this round if the necesary comitments will arrive later
	}

	select {
	case rcvMsg := <-sr.cns.ChRcvMsg:
		if sr.cns.ConsumeReceivedMessage(&rcvMsg, chr) {
			bActionDone = true
		}
	default:
	}

	if bActionDone {
		bActionDone = false
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(SR_BLOCK), chronology.Subround(SR_COMITMENT)); ok {
			sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 4: Received %d from %d comitments, which are matching with bitmap and are enough", n, len(sr.cns.ConsensusGroup)))
			return R_True
		}
	}

	return R_None
}

func (sr *SRComitment) Current() chronology.Subround {
	return chronology.Subround(SR_COMITMENT)
}

func (sr *SRComitment) Next() chronology.Subround {
	return chronology.Subround(SR_SIGNATURE)
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
