package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

type SRComitmentHash struct {
	doLog                   bool
	endTime                 int64
	cns                     *Consensus
	OnReceivedComitmentHash func(*[]byte, *chronology.Chronology) bool
	OnSendComitmentHash     func(chronology.Subround) bool
}

func NewSRComitmentHash(doLog bool, endTime int64, cns *Consensus, onReceivedComitmentHash func(*[]byte, *chronology.Chronology) bool, onSendComitmentHash func(chronology.Subround) bool) *SRComitmentHash {
	sr := SRComitmentHash{doLog: doLog, endTime: endTime, cns: cns, OnReceivedComitmentHash: onReceivedComitmentHash, OnSendComitmentHash: onSendComitmentHash}
	return &sr
}

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

func (sr *SRComitmentHash) doComitmentHash(chr *chronology.Chronology) Response {
	sr.cns.SetSentMessage(sr.OnSendComitmentHash(chronology.Subround(SrComitmentHash)))

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

func (sr *SRComitmentHash) Current() chronology.Subround {
	return chronology.Subround(SrComitmentHash)
}

func (sr *SRComitmentHash) Next() chronology.Subround {
	return chronology.Subround(SrBitmap)
}

func (sr *SRComitmentHash) EndTime() int64 {
	return int64(sr.endTime)
}

func (sr *SRComitmentHash) Name() string {
	return "<COMITMENT_HASH>"
}

func (sr *SRComitmentHash) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
