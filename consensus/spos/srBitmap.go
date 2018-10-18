package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRBitmap struct {
	doLog            bool
	endTime          int64
	cns              *Consensus
	OnReceivedBitmap func(*[]byte, *chronology.Chronology) bool
	OnSendBitmap     func(chronology.Subround) bool
}

func NewSRBitmap(doLog bool, endTime int64, cns *Consensus, onReceivedBitmap func(*[]byte, *chronology.Chronology) bool, onSendBitmap func(chronology.Subround) bool) *SRBitmap {
	sr := SRBitmap{doLog: doLog, endTime: endTime, cns: cns, OnReceivedBitmap: onReceivedBitmap, OnSendBitmap: onSendBitmap}
	return &sr
}

func (sr *SRBitmap) DoWork(chr *chronology.Chronology) bool {
	for chr.SelfSubround() != chronology.SrCanceled {
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

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 3: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

func (sr *SRBitmap) doBitmap(chr *chronology.Chronology) Response {
	sr.cns.SetSentMessage(sr.OnSendBitmap(chronology.Subround(SrBitmap)))

	if sr.cns.SentMessage() {
		sr.cns.SetSentMessage(false)
		if ok, _ := sr.cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrBitmap)); ok {
			return rTrue
		}
	}

	timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

	if timeSubRound > chronology.Subround(SrBitmap) {
		sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) + "Step 3: Extended the " + sr.Name() + " subround"))
		sr.cns.RoundStatus.Bitmap = SsExtended
		return rTrue // Try to give a chance to this round if the bitmap from leader will arrive later
	}

	if sr.cns.ReceivedMessage() {
		sr.cns.SetReceivedMessage(false)
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrBitmap)); ok {
			addMessage := "BUT I WAS NOT selected in this bitmap"
			if sr.cns.IsNodeInBitmapGroup(sr.cns.Self) {
				addMessage = "AND I WAS selected in this bitmap"
			}
			sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 3: Received bitmap from leader, matching with my own, and it got %d from %d comitment hashes, which are enough, %s", n, len(sr.cns.ConsensusGroup), addMessage))
			return rTrue
		}
	}

	return rNone
}

func (sr *SRBitmap) Current() chronology.Subround {
	return chronology.Subround(SrBitmap)
}

func (sr *SRBitmap) Next() chronology.Subround {
	return chronology.Subround(SrComitment)
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
