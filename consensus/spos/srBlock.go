package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRBlock struct {
	doLog   bool
	endTime int64
	cns     *Consensus
}

func NewSRBlock(doLog bool, endTime int64, cns *Consensus) SRBlock {
	sr := SRBlock{doLog: doLog, endTime: endTime, cns: cns}
	return sr
}

func (sr *SRBlock) DoWork() bool {
	for sr.cns.chr.GetSelfSubRound() != chronology.SR_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		switch sr.doBlock() {
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

	sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime()+"Step 1: Aborded round %d in subround %s", sr.cns.chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SRBlock) doBlock() Response {
	bActionDone := sr.cns.SendMessage(chronology.Subround(SR_BLOCK))

	if bActionDone {
		bActionDone = false
		if ok, _ := sr.cns.CheckConsensus(chronology.Subround(SR_BLOCK), chronology.Subround(SR_BLOCK)); ok {
			return R_True
		}
	}

	timeSubRound := sr.cns.chr.GetSubRoundFromDateTime(sr.cns.chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(SR_BLOCK) {
		sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime() + "Step 1: Extended the " + sr.Name() + " subround"))
		sr.cns.RoundStatus.Block = SS_EXTENDED
		return R_True // Try to give a chance to this round if the block from leader will arrive later
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
		if ok, _ := sr.cns.CheckConsensus(chronology.Subround(SR_BLOCK), chronology.Subround(SR_BLOCK)); ok {
			sr.Log(fmt.Sprintf(sr.cns.chr.GetFormatedCurrentTime() + "Step 1: Synchronized block"))
			return R_True
		}
	}

	return R_None
}

func (sr *SRBlock) Current() chronology.Subround {
	return chronology.Subround(SR_BLOCK)
}

func (sr *SRBlock) Next() chronology.Subround {
	return chronology.Subround(SR_COMITMENT_HASH)
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
