package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SRSignature defines the data needed by the signature subround
type SRSignature struct {
	doLog           bool
	endTime         int64
	Cns             *Consensus
	OnSendSignature func() bool
}

// NewSRSignature creates a new SRSignature object
func NewSRSignature(doLog bool, endTime int64, cns *Consensus, onSendSignature func() bool) *SRSignature {
	sr := SRSignature{doLog: doLog, endTime: endTime, Cns: cns, OnSendSignature: onSendSignature}
	return &sr
}

// DoWork method calls repeatedly DoSignature method, which is in charge to do the job of this subround, until RTrue or RFalse is return
// or until this subround is put in the canceled mode
func (sr *SRSignature) DoWork(chr *chronology.Chronology) bool {
	for chr.SelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.DoSignature(chr) {
		case RNone:
			continue
		case RFalse:
			return false
		case RTrue:
			return true
		default:
			return false
		}
	}

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 5: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

// DoSignature method actually do the job of this subround. First, it tries to send the signature and than if the signature was succesfully sent or
// if meantime a new message was received, it will check the consensus again. If the upper time limit of this subround is reached, it's state
// is set to extended and the chronology will advance to the next subround
func (sr *SRSignature) DoSignature(chr *chronology.Chronology) Response {
	sr.Cns.SetSentMessage(sr.OnSendSignature())

	timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

	if timeSubRound > chronology.Subround(SrSignature) {
		sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 5: Extended the "+sr.Name()+" subround. Got only %d from %d sigantures which are not enough", sr.Cns.GetSignaturesCount(), len(sr.Cns.ConsensusGroup)))
		sr.Cns.RoundStatus.Signature = SsExtended
		return RTrue // Try to give a chance to this round if the necesary signatures will arrive later
	}

	if sr.Cns.SentMessage() || sr.Cns.ReceivedMessage() {
		sr.Cns.SetSentMessage(false)
		sr.Cns.SetReceivedMessage(false)
		if ok, n := sr.Cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrSignature)); ok {
			sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 5: Received %d from %d signatures, which are matching with bitmap and are enough", n, len(sr.Cns.ConsensusGroup)))
			return RTrue
		}
	}

	return RNone
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
