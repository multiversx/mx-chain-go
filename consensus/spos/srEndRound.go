package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SREndRound struct {
	doLog            bool
	endTime          int64
	cns              *Consensus
	roundsWithBlocks int // only for statistic
}

func NewSREndRound(doLog bool, endTime int64, cns *Consensus) SREndRound {
	sr := SREndRound{doLog: doLog, endTime: endTime, cns: cns}
	return sr
}

func (sr *SREndRound) DoWork(chr *chronology.Chronology) bool {
	bActionDone := true
	for chr.GetSelfSubround() != chronology.SR_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)
		switch sr.doEndRound(chr, &bActionDone) {
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

	sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 6: Aborded round %d in subround %s", chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SREndRound) doEndRound(chr *chronology.Chronology, bActionDone *bool) Response {
	timeSubRound := chr.GetSubroundFromDateTime(chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(SR_END_ROUND) {
		sr.Log(fmt.Sprintf("\n" + chr.GetFormatedCurrentTime() + ">>>>>>>>>>>>>>>>>>>> THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n"))
		return R_True
	}

	select {
	case rcvMsg := <-sr.cns.ChRcvMsg:
		if sr.cns.ConsumeReceivedMessage(&rcvMsg, chr) {
			*bActionDone = true
		}
	default:
	}

	if *bActionDone {
		*bActionDone = false
		if ok, _ := sr.cns.CheckConsensus(chronology.Subround(SR_BLOCK), chronology.Subround(SR_SIGNATURE)); ok {
			sr.roundsWithBlocks++ // only for statistic

			sr.cns.BlockChain.AddBlock(*sr.cns.Block)

			if sr.cns.IsNodeLeaderInCurrentRound(sr.cns.Self) {
				sr.Log(fmt.Sprintf("\n"+chr.GetFormatedCurrentTime()+">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", sr.cns.Block.Nonce))
			} else {
				sr.Log(fmt.Sprintf("\n"+chr.GetFormatedCurrentTime()+">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", sr.cns.Block.Nonce))
			}

			return R_True
		}
	}

	return R_None
}

func (sr *SREndRound) Current() chronology.Subround {
	return chronology.Subround(SR_END_ROUND)
}

func (sr *SREndRound) Next() chronology.Subround {
	return chronology.Subround(SR_START_ROUND)
}

func (sr *SREndRound) EndTime() int64 {
	return int64(sr.endTime)
}

func (sr *SREndRound) Name() string {
	return "<END_ROUND>"
}

func (sr *SREndRound) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
