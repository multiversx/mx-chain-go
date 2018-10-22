package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SRComitmentHash defines the data needed by the comitment hash subround
type SRComitmentHash struct {
	doLog                   bool
	endTime                 int64
	cns                     *Consensus
	OnReceivedComitmentHash func(*[]byte, *chronology.Chronology) bool
	OnSendComitmentHash     func() bool
}

// NewSRComitmentHash creates a new SRComitmentHash object
func NewSRComitmentHash(doLog bool, endTime int64, cns *Consensus, onReceivedComitmentHash func(*[]byte, *chronology.Chronology) bool, onSendComitmentHash func() bool) *SRComitmentHash {
	sr := SRComitmentHash{doLog: doLog, endTime: endTime, cns: cns, OnReceivedComitmentHash: onReceivedComitmentHash, OnSendComitmentHash: onSendComitmentHash}
	return &sr
}

// DoWork method calls repeatedly doComitmentHash method, which is in charge to do the job of this subround, until rTrue or rFalse is return
// or until this subround is put in the canceled mode
func (sr *SRComitmentHash) DoWork(chr *chronology.Chronology) bool {
	for chr.SelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.doComitmentHash(chr) {
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

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 2: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

// doComitmentHash method actually do the job of this subround. First, it tries to send the comitment hash and than if the comitment hash was succesfully sent or
// if meantime a new message was received, it will check the consensus again. If the upper time limit of this subround is reached, it's state
// is set to extended and the chronology will advance to the next subround
func (sr *SRComitmentHash) doComitmentHash(chr *chronology.Chronology) Response {
	sr.cns.SetSentMessage(sr.OnSendComitmentHash())

	timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

	if timeSubRound > chronology.Subround(SrComitmentHash) {
		if sr.cns.GetComitmentHashesCount() < sr.cns.Threshold.ComitmentHash {
			sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 2: Extended the "+sr.Name()+" subround. Got only %d from %d commitment hashes which are not enough", sr.cns.GetComitmentHashesCount(), len(sr.cns.ConsensusGroup)))
		} else {
			sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) + "Step 2: Extended the " + sr.Name() + " subround"))
		}
		sr.cns.RoundStatus.ComitmentHash = SsExtended
		return rTrue // Try to give a chance to this round if the necesary comitment hashes will arrive later
	}

	if sr.cns.SentMessage() || sr.cns.ReceivedMessage() {
		sr.cns.SetSentMessage(false)
		sr.cns.SetReceivedMessage(false)
		if ok, n := sr.cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrComitmentHash)); ok {
			if n == len(sr.cns.ConsensusGroup) {
				sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 2: Received all (%d from %d) comitment hashes", n, len(sr.cns.ConsensusGroup)))
			} else {
				sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 2: Received %d from %d comitment hashes, which are enough", n, len(sr.cns.ConsensusGroup)))
			}
			return rTrue
		}
	}

	return rNone
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
