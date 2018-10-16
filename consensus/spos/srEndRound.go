package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SREndRound struct {
	doLog             bool
	endTime           int64
	cns               *Consensus
	roundsWithBlocks  int // only for statistic
	OnReceivedMessage func(*[]byte, *chronology.Chronology) bool
	OnSendMessage     func(chronology.Subround) bool
}

func NewSREndRound(doLog bool, endTime int64, cns *Consensus, onReceivedMessage func(*[]byte, *chronology.Chronology) bool, onSendMessage func(chronology.Subround) bool) *SREndRound {
	sr := SREndRound{doLog: doLog, endTime: endTime, cns: cns, OnReceivedMessage: onReceivedMessage, OnSendMessage: onSendMessage}
	return &sr
}

func (sr *SREndRound) DoWork(chr *chronology.Chronology) bool {
	bActionDone := true
	for chr.GetSelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.doEndRound(chr, &bActionDone) {
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

	sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 6: Canceled round %d in subround %s", chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SREndRound) doEndRound(chr *chronology.Chronology, bActionDone *bool) Response {
	timeSubRound := chr.GetSubroundFromDateTime(chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(srEndRound) {
		sr.Log(fmt.Sprintf("\n" + chr.GetFormatedCurrentTime() + ">>>>>>>>>>>>>>>>>>>> THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n"))
		return rTrue
	}

	select {
	case rcvMsg := <-sr.cns.ChRcvMsg:
		if sr.OnReceivedMessage(&rcvMsg, chr) {
			*bActionDone = true
		}
	default:
	}

	if *bActionDone {
		*bActionDone = false
		if ok, _ := sr.cns.CheckConsensus(chronology.Subround(srBlock), chronology.Subround(srSignature)); ok {
			sr.roundsWithBlocks++ // only for statistic

			sr.cns.BlockChain.AddBlock(*sr.cns.Block)

			if sr.cns.IsNodeLeaderInCurrentRound(sr.cns.Self) {
				sr.Log(fmt.Sprintf("\n"+chr.GetFormatedCurrentTime()+">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", sr.cns.Block.Nonce))
			} else {
				sr.Log(fmt.Sprintf("\n"+chr.GetFormatedCurrentTime()+">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", sr.cns.Block.Nonce))
			}

			return rTrue
		}
	}

	return rNone
}

func (sr *SREndRound) Current() chronology.Subround {
	return chronology.Subround(srEndRound)
}

func (sr *SREndRound) Next() chronology.Subround {
	return chronology.Subround(srStartRound)
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
