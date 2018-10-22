package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SRBitmap defines the data needed by the bitmap subround
type SRBitmap struct {
	doLog            bool
	endTime          int64
	cns              *Consensus
	OnReceivedBitmap func(*[]byte, *chronology.Chronology) bool
	OnSendBitmap     func() bool
}

// NewSRBitmap creates a new SRBitmap object
func NewSRBitmap(doLog bool, endTime int64, cns *Consensus, onReceivedBitmap func(*[]byte, *chronology.Chronology) bool, onSendBitmap func() bool) *SRBitmap {
	sr := SRBitmap{doLog: doLog, endTime: endTime, cns: cns, OnReceivedBitmap: onReceivedBitmap, OnSendBitmap: onSendBitmap}
	return &sr
}

// DoWork method calls repeatedly doBitmap method, which is in charge to do the job of this subround, until rTrue or rFalse is return
// or until this subround is put in the canceled mode
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

// doBitmap method actually do the job of this subround. First, it tries to send the bitmap (if this node is in charge with this action)
// and than if the bitmap was succesfully sent or if meantime a new message was received, it will check the consensus again.
// If the upper time limit of this subround is reached, it's state is set to extended and the chronology will advance to the next subround
func (sr *SRBitmap) doBitmap(chr *chronology.Chronology) Response {
	sr.cns.SetSentMessage(sr.OnSendBitmap())

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

// Current method returns the ID of this subround
func (sr *SRBitmap) Current() chronology.Subround {
	return chronology.Subround(SrBitmap)
}

// Next method returns the ID of the next subround
func (sr *SRBitmap) Next() chronology.Subround {
	return chronology.Subround(SrComitment)
}

// EndTime method returns the upper time limit of this subround
func (sr *SRBitmap) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of this subround
func (sr *SRBitmap) Name() string {
	return "<BITMAP>"
}

// Log method prints info about this subrond (if doLog is true)
func (sr *SRBitmap) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
