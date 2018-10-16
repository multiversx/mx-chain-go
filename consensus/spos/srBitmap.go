package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRBitmap struct {
	doLog             bool
	endTime           int64
	cns               *Consensus
	OnReceivedMessage func(*[]byte, *chronology.Chronology) bool
	OnSendMessage     func(chronology.Subround) bool
}

func NewSRBitmap(doLog bool, endTime int64, cns *Consensus, onReceivedMessage func(*[]byte, *chronology.Chronology) bool, onSendMessage func(chronology.Subround) bool) *SRBitmap {
	sr := SRBitmap{doLog: doLog, endTime: endTime, cns: cns, OnReceivedMessage: onReceivedMessage, OnSendMessage: onSendMessage}
	return &sr
}

func (sr *SRBitmap) DoWork(chr *chronology.Chronology) bool {
	for chr.GetSelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.doBitmap(chr) {
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

	sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 3: Canceled round %d in subround %s", chr.GetRoundIndex(), sr.Name()))
	return false
}

func (sr *SRBitmap) doBitmap(chr *chronology.Chronology) Response {
	bActionDone := sr.OnSendMessage(chronology.Subround(srBitmap))

	if bActionDone {
		bActionDone = false
		if ok, _ := sr.cns.CheckConsensus(chronology.Subround(srBlock), chronology.Subround(srBitmap)); ok {
			return rTrue
		}
	}

	timeSubRound := chr.GetSubroundFromDateTime(chr.GetCurrentTime())

	if timeSubRound > chronology.Subround(srBitmap) {
		sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime() + "Step 3: Extended the " + sr.Name() + " subround"))
		sr.cns.RoundStatus.Bitmap = ssExtended
		return rTrue // Try to give a chance to this round if the bitmap from leader will arrive later
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
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(srBlock), chronology.Subround(srBitmap)); ok {
			addMessage := "BUT I WAS NOT selected in this bitmap"
			if sr.cns.IsNodeInBitmapGroup(sr.cns.Self) {
				addMessage = "AND I WAS selected in this bitmap"
			}
			sr.Log(fmt.Sprintf(chr.GetFormatedCurrentTime()+"Step 3: Received bitmap from leader, matching with my own, and it got %d from %d comitment hashes, which are enough, %s", n, len(sr.cns.ConsensusGroup), addMessage))
			return rTrue
		}
	}

	return rNone
}

func (sr *SRBitmap) Current() chronology.Subround {
	return chronology.Subround(srBitmap)
}

func (sr *SRBitmap) Next() chronology.Subround {
	return chronology.Subround(srComitment)
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
