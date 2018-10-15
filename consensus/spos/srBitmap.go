package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRBitmap struct {
	doLog   bool
	endTime int64
	cns     *Consensus
}

func NewSRBitmap(doLog bool, endTime int64, cns *Consensus) SRBitmap {
	sr := SRBitmap{doLog: doLog, endTime: endTime, cns: cns}
	return sr
}

func (sr *SRBitmap) DoWork(chr *chronology.Chronology) bool {
	for chr.GetSelfSubround() != chronology.SR_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		switch sr.doBitmap(chr) {
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

	sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 3: Aborded round %d in subround %s", chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SRBitmap) doBitmap(chr *chronology.Chronology) Response {
	bActionDone := sr.cns.SendMessage(chronology.Subround(SR_BITMAP))

	if bActionDone {
		bActionDone = false
		if ok, _ := sr.cns.CheckConsensus(chronology.Subround(SR_BLOCK), chronology.Subround(SR_BITMAP)); ok {
			return R_True
		}
	}

	timeSubRound := chr.GetSubroundFromDateTime(chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(SR_BITMAP) {
		sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime() + "Step 3: Extended the " + sr.Name() + " subround"))
		sr.cns.RoundStatus.Bitmap = SS_EXTENDED
		return R_True // Try to give a chance to this round if the bitmap from leader will arrive later
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
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(SR_BLOCK), chronology.Subround(SR_BITMAP)); ok {
			addMessage := "BUT I WAS NOT selected in this bitmap"
			if sr.cns.IsNodeInBitmapGroup(sr.cns.Self) {
				addMessage = "AND I WAS selected in this bitmap"
			}
			sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 3: Received bitmap from leader, matching with my own, and it got %d from %d comitment hashes, which are enough, %s", n, len(sr.cns.ConsensusGroup), addMessage))
			return R_True
		}
	}

	return R_None
}

func (sr *SRBitmap) Current() chronology.Subround {
	return chronology.Subround(SR_BITMAP)
}

func (sr *SRBitmap) Next() chronology.Subround {
	return chronology.Subround(SR_COMITMENT)
}

func (sr *SRBitmap) EndTime() int64 {
	return int64(sr.endTime)
}

func (sr *SRBitmap) Name() string {
	return "<BITMAP>"
}

func (sr *SRBitmap) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
