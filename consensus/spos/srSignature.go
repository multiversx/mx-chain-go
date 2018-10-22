package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SRSignature defines the data needed by the signature subround
type SRSignature struct {
	doLog               bool
	endTime             int64
	cns                 *Consensus
	OnReceivedSignature func(*[]byte, *chronology.Chronology) bool
	OnSendSignature     func() bool
}

// NewSRSignature creates a new SRSignature object
func NewSRSignature(doLog bool, endTime int64, cns *Consensus, onReceivedSignature func(*[]byte, *chronology.Chronology) bool, onSendSignature func() bool) *SRSignature {
	sr := SRSignature{doLog: doLog, endTime: endTime, cns: cns, OnReceivedSignature: onReceivedSignature, OnSendSignature: onSendSignature}
	return &sr
}

// DoWork method calls repeatedly doSignature method, which is in charge to do the job of this subround, until rTrue or rFalse is return
// or until this subround is put in the canceled mode
func (sr *SRSignature) DoWork(chr *chronology.Chronology) bool {
	for chr.SelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.doSignature(chr) {
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

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 5: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

// doSignature method actually do the job of this subround. First, it tries to send the signature and than if the signature was succesfully sent or
// if meantime a new message was received, it will check the consensus again. If the upper time limit of this subround is reached, it's state
// is set to extended and the chronology will advance to the next subround
func (sr *SRSignature) doSignature(chr *chronology.Chronology) Response {
	sr.cns.SetSentMessage(sr.OnSendSignature())

	timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

	if timeSubRound > chronology.Subround(SrSignature) {
		sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 5: Extended the "+sr.Name()+" subround. Got only %d from %d sigantures which are not enough", sr.cns.GetSignaturesCount(), len(sr.cns.ConsensusGroup)))
		sr.cns.RoundStatus.Signature = SsExtended
		return rTrue // Try to give a chance to this round if the necesary signatures will arrive later
	}

	if sr.cns.SentMessage() || sr.cns.ReceivedMessage() {
		sr.cns.SetSentMessage(false)
		sr.cns.SetReceivedMessage(false)
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrSignature)); ok {
			sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 5: Received %d from %d signatures, which are matching with bitmap and are enough", n, len(sr.cns.ConsensusGroup)))
			return rTrue
		}
	}

	return rNone
}

// Current method returns the ID of this subround
func (sr *SRSignature) Current() chronology.Subround {
	return chronology.Subround(SrSignature)
}

// Next method returns the ID of the next subround
func (sr *SRSignature) Next() chronology.Subround {
	return chronology.Subround(SrEndRound)
}

// EndTime method returns the upper time limit of this subround
func (sr *SRSignature) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of this subround
func (sr *SRSignature) Name() string {
	return "<SIGNATURE>"
}

// Log method prints info about this subrond (if doLog is true)
func (sr *SRSignature) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
