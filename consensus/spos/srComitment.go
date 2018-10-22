package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SRComitment defines the data needed by the comitment subround
type SRComitment struct {
	doLog           bool
	endTime         int64
	Cns             *Consensus
	OnSendComitment func() bool
}

// NewSRComitment creates a new SRComitment object
func NewSRComitment(doLog bool, endTime int64, cns *Consensus, onSendComitment func() bool) *SRComitment {
	sr := SRComitment{doLog: doLog, endTime: endTime, Cns: cns, OnSendComitment: onSendComitment}
	return &sr
}

// DoWork method calls repeatedly DoComitment method, which is in charge to do the job of this subround, until RTrue or RFalse is return
// or until this subround is put in the canceled mode
func (sr *SRComitment) DoWork(chr *chronology.Chronology) bool {
	for chr.SelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.DoComitment(chr) {
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

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 4: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

// DoComitment method actually do the job of this subround. First, it tries to send the comitment and than if the comitment was succesfully sent or
// if meantime a new message was received, it will check the consensus again. If the upper time limit of this subround is reached, it's state
// is set to extended and the chronology will advance to the next subround
func (sr *SRComitment) DoComitment(chr *chronology.Chronology) Response {
	sr.Cns.SetSentMessage(sr.OnSendComitment())

	timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

	if timeSubRound > chronology.Subround(SrComitment) {
		sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 4: Extended the "+sr.Name()+" subround. Got only %d from %d commitments which are not enough", sr.Cns.GetComitmentsCount(), len(sr.Cns.ConsensusGroup)))
		sr.Cns.RoundStatus.Comitment = SsExtended
		return RTrue // Try to give a chance to this round if the necesary comitments will arrive later
	}

	if sr.Cns.SentMessage() || sr.Cns.ReceivedMessage() {
		sr.Cns.SetSentMessage(false)
		sr.Cns.SetReceivedMessage(false)
		if ok, n := sr.Cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrComitment)); ok {
			sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 4: Received %d from %d comitments, which are matching with bitmap and are enough", n, len(sr.Cns.ConsensusGroup)))
			return RTrue
		}
	}

	return RNone
}

// Current method returns the ID of this subround
func (sr *SRComitment) Current() chronology.Subround {
	return chronology.Subround(SrComitment)
}

// Next method returns the ID of the next subround
func (sr *SRComitment) Next() chronology.Subround {
	return chronology.Subround(SrSignature)
}

// EndTime method returns the upper time limit of this subround
func (sr *SRComitment) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of this subround
func (sr *SRComitment) Name() string {
	return "<COMITMENT>"
}

// Log method prints info about this subrond (if doLog is true)
func (sr *SRComitment) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
