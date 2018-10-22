package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SREndRound defines the data needed by the end round subround
type SREndRound struct {
	doLog      bool
	endTime    int64
	Cns        *Consensus
	OnEndRound func()
}

// NewSREndRound creates a new SREndRound object
func NewSREndRound(doLog bool, endTime int64, cns *Consensus, onEndRound func()) *SREndRound {
	sr := SREndRound{doLog: doLog, endTime: endTime, Cns: cns, OnEndRound: onEndRound}
	return &sr
}

// DoWork method calls repeatedly DoEndRound method, which is in charge to do the job of this subround, until RTrue or RFalse is return
// or until this subround is put in the canceled mode
func (sr *SREndRound) DoWork(chr *chronology.Chronology) bool {
	sr.Cns.SetReceivedMessage(true)
	for chr.SelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.DoEndRound(chr) {
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

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 6: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

// DoEndRound method actually do the job of this subround. First, it checks the consensus and if it is not done yet it waits for a new message to be
// received, and than it will check the consensus again. If the upper time limit of this subround is reached and the consensus is not done, this round
// no block will be added to the blockchain
func (sr *SREndRound) DoEndRound(chr *chronology.Chronology) Response {
	timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

	if timeSubRound > chronology.Subround(SrEndRound) {
		sr.Log(fmt.Sprintf("\n" + chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) + ">>>>>>>>>>>>>>>>>>>> THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n"))
		return RTrue
	}

	if sr.Cns.ReceivedMessage() {
		sr.Cns.SetReceivedMessage(false)
		if ok, _ := sr.Cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrSignature)); ok {

			sr.OnEndRound()

			return RTrue
		}
	}

	return RNone
}

// Current method returns the ID of this subround
func (sr *SREndRound) Current() chronology.Subround {
	return chronology.Subround(SrEndRound)
}

// Next method returns the ID of the next subround
func (sr *SREndRound) Next() chronology.Subround {
	return chronology.Subround(SrStartRound)
}

// EndTime method returns the upper time limit of this subround
func (sr *SREndRound) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of this subround
func (sr *SREndRound) Name() string {
	return "<END_ROUND>"
}

// Log method prints info about this subrond (if doLog is true)
func (sr *SREndRound) Log(message string) {
	if sr.doLog {
		fmt.Printf(message + "\n")
	}
}
