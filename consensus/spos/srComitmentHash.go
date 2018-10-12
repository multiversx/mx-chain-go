package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRComitmentHash struct {
	doLog   bool
	endTime int64
	cns     *Consensus
}

func NewSRComitmentHash(doLog bool, endTime int64, cns *Consensus) SRComitmentHash {
	sr := SRComitmentHash{doLog: doLog, endTime: endTime, cns: cns}
	return sr
}

func (sr *SRComitmentHash) DoWork() bool {
	for sr.cns.chr.GetSelfSubRound() != chronology.SR_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		switch sr.doComitmentHash() {
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

	sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime()+"Step 2: Aborded round %d in subround %s", sr.cns.chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SRComitmentHash) doComitmentHash() Response {
	bActionDone := sr.cns.SendMessage(chronology.Subround(SR_COMITMENT_HASH))

	timeSubRound := sr.cns.chr.GetSubRoundFromDateTime(sr.cns.chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(SR_COMITMENT_HASH) {
		if sr.cns.GetComitmentHashesCount() < sr.cns.Threshold.ComitmentHash {
			sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime()+"Step 2: Extended the "+sr.Name()+" subround. Got only %d from %d commitment hashes which are not enough", sr.cns.GetComitmentHashesCount(), len(sr.cns.ConsensusGroup)))
		} else {
			sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime() + "Step 2: Extended the " + sr.Name() + " subround"))
		}
		sr.cns.RoundStatus.ComitmentHash = SS_EXTENDED
		return R_True // Try to give a chance to this round if the necesary comitment hashes will arrive later
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
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(SR_BLOCK), chronology.Subround(SR_COMITMENT_HASH)); ok {
			if n == len(sr.cns.ConsensusGroup) {
				sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime()+"Step 2: Received all (%d from %d) comitment hashes", n, len(sr.cns.ConsensusGroup)))
			} else {
				sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime()+"Step 2: Received %d from %d comitment hashes, which are enough", n, len(sr.cns.ConsensusGroup)))
			}
			return R_True
		}
	}

	return R_None
}

func (sr *SRComitmentHash) Current() chronology.Subround {
	return chronology.Subround(SR_COMITMENT_HASH)
}

func (sr *SRComitmentHash) Next() chronology.Subround {
	return chronology.Subround(SR_BITMAP)
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
