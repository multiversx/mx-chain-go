package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SRComitmentHash defines the data needed by the comitment hash subround
type SRComitmentHash struct {
	doLog               bool
	endTime             int64
	Cns                 *Consensus
	OnSendComitmentHash func() bool
}

// NewSRComitmentHash creates a new SRComitmentHash object
func NewSRComitmentHash(doLog bool, endTime int64, cns *Consensus, onSendComitmentHash func() bool) *SRComitmentHash {
	sr := SRComitmentHash{doLog: doLog, endTime: endTime, Cns: cns, OnSendComitmentHash: onSendComitmentHash}
	return &sr
}

// DoWork method calls repeatedly DoComitmentHash method, which is in charge to do the job of this subround, until RTrue or RFalse is return
// or until this subround is put in the canceled mode
func (sr *SRComitmentHash) DoWork(chr *chronology.Chronology) bool {
	for chr.SelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.DoComitmentHash(chr) {
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

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 2: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

// DoComitmentHash method actually do the job of this subround. First, it tries to send the comitment hash and than if the comitment hash was succesfully sent or
// if meantime a new message was received, it will check the consensus again. If the upper time limit of this subround is reached, it's state
// is set to extended and the chronology will advance to the next subround
func (sr *SRComitmentHash) DoComitmentHash(chr *chronology.Chronology) Response {
	sr.Cns.SetSentMessage(sr.OnSendComitmentHash())

	timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

	if timeSubRound > chronology.Subround(SrComitmentHash) {
		if sr.Cns.GetComitmentHashesCount() < sr.Cns.Threshold.ComitmentHash {
			sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 2: Extended the "+sr.Name()+" subround. Got only %d from %d commitment hashes which are not enough", sr.Cns.GetComitmentHashesCount(), len(sr.Cns.ConsensusGroup)))
		} else {
			sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) + "Step 2: Extended the " + sr.Name() + " subround"))
		}
		sr.Cns.RoundStatus.ComitmentHash = SsExtended
		return RTrue // Try to give a chance to this round if the necesary comitment hashes will arrive later
	}

	if sr.Cns.SentMessage() || sr.Cns.ReceivedMessage() {
		sr.Cns.SetSentMessage(false)
		sr.Cns.SetReceivedMessage(false)
		if ok, n := sr.Cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrComitmentHash)); ok {
			if n == len(sr.Cns.ConsensusGroup) {
				sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 2: Received all (%d from %d) comitment hashes", n, len(sr.Cns.ConsensusGroup)))
			} else {
				sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 2: Received %d from %d comitment hashes, which are enough", n, len(sr.Cns.ConsensusGroup)))
			}
			return RTrue
		}
	}

	return RNone
}

// Current method returns the ID of this subround
func (sr *SRComitmentHash) Current() chronology.Subround {
	return chronology.Subround(SrComitmentHash)
}

// Next method returns the ID of the next subround
func (sr *SRComitmentHash) Next() chronology.Subround {
	return chronology.Subround(SrBitmap)
}

// EndTime method returns the upper time limit of this subround
func (sr *SRComitmentHash) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of this subround
func (sr *SRComitmentHash) Name() string {
	return "<COMITMENT_HASH>"
}

// Log method prints info about this subrond (if doLog is true)
func (sr *SRComitmentHash) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
