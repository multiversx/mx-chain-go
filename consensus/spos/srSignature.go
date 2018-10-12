package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRSignature struct {
	doLog   bool
	endTime int64
	cns     *Consensus
}

func NewSRSignature(doLog bool, endTime int64, cns *Consensus) SRSignature {
	sr := SRSignature{doLog: doLog, endTime: endTime, cns: cns}
	return sr
}

func (sr *SRSignature) DoWork() bool {
	for sr.cns.chr.GetSelfSubRound() != chronology.SR_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		switch sr.doSignature() {
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

	sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime()+"Step 5: Aborded round %d in subround %s", sr.cns.chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SRSignature) doSignature() Response {
	bActionDone := sr.cns.SendMessage(chronology.Subround(SR_SIGNATURE))

	timeSubRound := sr.cns.chr.GetSubRoundFromDateTime(sr.cns.chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(SR_SIGNATURE) {
		sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime()+"Step 5: Extended the "+sr.Name()+" subround. Got only %d from %d sigantures which are not enough", sr.cns.GetSignaturesCount(), len(sr.cns.ConsensusGroup)))
		sr.cns.RoundStatus.Signature = SS_EXTENDED
		return R_True // Try to give a chance to this round if the necesary signatures will arrive later
	}

	select {
	case rcvMsg := <-sr.cns.ChRcvMsg:
		if sr.cns.ConsumeReceivedMessage(&rcvMsg, timeSubRound) {
			bActionDone = true
		}
	default:
	}

	if bActionDone {
		bActionDone = false
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(SR_BLOCK), chronology.Subround(SR_SIGNATURE)); ok {
			sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime()+"Step 5: Received %d from %d signatures, which are matching with bitmap and are enough", n, len(sr.cns.ConsensusGroup)))
			return R_True
		}
	}

	return R_None
}

func (sr *SRSignature) Current() chronology.Subround {
	return chronology.Subround(SR_SIGNATURE)
}

func (sr *SRSignature) Next() chronology.Subround {
	return chronology.Subround(SR_END_ROUND)
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
