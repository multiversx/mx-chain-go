package spos

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SREndRound defines the data needed by the end round subround
type SREndRound struct {
	doLog              bool
	endTime            int64
	cns                *Consensus
	roundsWithBlocks   int // only for statistic
	OnReceivedEndRound func(*[]byte, *chronology.Chronology) bool
	OnSendEndRound     func() bool
}

// NewSREndRound creates a new SREndRound object
func NewSREndRound(doLog bool, endTime int64, cns *Consensus, onReceivedEndRound func(*[]byte, *chronology.Chronology) bool, onSendEndRound func() bool) *SREndRound {
	sr := SREndRound{doLog: doLog, endTime: endTime, cns: cns, OnReceivedEndRound: onReceivedEndRound, OnSendEndRound: onSendEndRound}
	return &sr
}

// DoWork method calls repeatedly doEndRound method, which is in charge to do the job of this subround, until rTrue or rFalse is return
// or until this subround is put in the canceled mode
func (sr *SREndRound) DoWork(chr *chronology.Chronology) bool {
	sr.cns.SetReceivedMessage(true)
	for chr.SelfSubround() != chronology.SrCanceled {
		time.Sleep(sleepTime * time.Millisecond)
		switch sr.doEndRound(chr) {
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

	sr.Log(fmt.Sprintf(chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+"Step 6: Canceled round %d in subround %s", chr.Round().Index(), sr.Name()))
	return false
}

// doEndRound method actually do the job of this subround. First, it checks the consensus and if it is not done yet it waits for a new message to be
// received, and than it will check the consensus again. If the upper time limit of this subround is reached and the consensus is not done, this round
// no block will be added to the blockchain
func (sr *SREndRound) doEndRound(chr *chronology.Chronology) Response {
	timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

	if timeSubRound > chronology.Subround(SrEndRound) {
		sr.Log(fmt.Sprintf("\n" + chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()) + ">>>>>>>>>>>>>>>>>>>> THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n"))
		return rTrue
	}

	if sr.cns.ReceivedMessage() {
		sr.cns.SetReceivedMessage(false)
		if ok, _ := sr.cns.CheckConsensus(chronology.Subround(SrBlock), chronology.Subround(SrSignature)); ok {
			sr.roundsWithBlocks++ // only for statistic

			//sr.cns.blockChain.AddBlock(*sr.cns.block)

			if sr.cns.IsNodeLeaderInCurrentRound(sr.cns.Self) {
				//sr.Log(fmt.Sprintf("\n"+Chr.SyncTime().FormatedCurrentTime(Chr.ClockOffset())+">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", sr.cns.block.Nonce))
				sr.Log(fmt.Sprintf("\n"+chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", sr.roundsWithBlocks-1))
			} else {
				//sr.Log(fmt.Sprintf("\n"+Chr.SyncTime().FormatedCurrentTime(Chr.ClockOffset())+">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", sr.cns.block.Nonce))
				sr.Log(fmt.Sprintf("\n"+chr.SyncTime().FormatedCurrentTime(chr.ClockOffset())+">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", sr.roundsWithBlocks-1))
			}

			return rTrue
		}
	}

	return rNone
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
