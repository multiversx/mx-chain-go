package spos

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	//"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	//"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

// roundTimeDuration defines the time duration in milliseconds of each round
const roundTimeDuration = time.Duration(4000 * time.Millisecond)

// sleepTime defines the time in milliseconds between each iteration made in DoWork methods of the subrounds
const sleepTime = 5

// Subround defines the type used to reffer the current subround
type Subround int

const (
	// SrStartRound defines ID of subround "Start round"
	SrStartRound Subround = iota
	// SrBlock defines ID of subround "Block"
	SrBlock
	// SrComitmentHash defines ID of subround "Comitment hash"
	SrComitmentHash
	// SrBitmap defines ID of subround "Bitmap"
	SrBitmap
	// SrComitment defines ID of subround "Comitment"
	SrComitment
	// SrSignature defines ID of subround "Signature"
	SrSignature
	// SrEndRound defines ID of subround "End round"
	SrEndRound
)

// Response defines the type used to reffer the subround job state in the DoWork methods of the subrounds
type Response int

const (
	// rFalse defines a "finished subround job without success" state in the DoWork methods of the subrounds
	rFalse Response = iota
	// rTrue defines a "finished subround job with success" state in the DoWork methods of the subrounds
	rTrue
	// rNone defines a "continue subround job" state in the DoWork methods of the subrounds
	rNone
)

// RoundStatus defines the data needed by spos to know the state of each subround of the current round
type RoundStatus struct {
	Block         SubroundStatus
	ComitmentHash SubroundStatus
	Bitmap        SubroundStatus
	Comitment     SubroundStatus
	Signature     SubroundStatus
}

// NewRoundStatus creates a new RoundStatus object
func NewRoundStatus(block SubroundStatus, comitmentHash SubroundStatus, bitmap SubroundStatus, comitment SubroundStatus, signature SubroundStatus) *RoundStatus {
	rs := RoundStatus{Block: block, ComitmentHash: comitmentHash, Bitmap: bitmap, Comitment: comitment, Signature: signature}
	return &rs
}

// ResetRoundStatus method resets the state of each subround
func (rs *RoundStatus) ResetRoundStatus() {
	rs.Block = SsNotFinished
	rs.ComitmentHash = SsNotFinished
	rs.Bitmap = SsNotFinished
	rs.Comitment = SsNotFinished
	rs.Signature = SsNotFinished
}

// SubroundStatus defines the type used to reffer the state of the current subround
type SubroundStatus int

const (
	// SsNotFinished defines the un-finished state of the subround
	SsNotFinished SubroundStatus = iota
	// SsExtended defines the extended state of the subround
	SsExtended
	// SsFinished defines the finished state of the subround
	SsFinished
)

// Threshold defines the minimum agreements needed for each subround to consider the subround finished. (Ex: PBFT threshold has 2 / 3 + 1 agreements)
type Threshold struct {
	Block         int
	ComitmentHash int
	Bitmap        int
	Comitment     int
	Signature     int
}

// NewThreshold creates a new Threshold object
func NewThreshold(block int, comitmentHash int, bitmap int, comitment int, signature int) *Threshold {
	th := Threshold{Block: block, ComitmentHash: comitmentHash, Bitmap: bitmap, Comitment: comitment, Signature: signature}
	return &th
}

// Consensus defines the data needed by spos to do the consensus in each round
type Consensus struct {
	doLog bool

	mutReceived sync.RWMutex
	mutSent     sync.RWMutex

	*Validators
	*Threshold
	*RoundStatus

	receivedMessage bool
	sentMessage     bool

	Chr *chronology.Chronology
}

// NewConsensus creates a new Consensus object
func NewConsensus(doLog bool, validators *Validators, threshold *Threshold, roundStatus *RoundStatus, chr *chronology.Chronology) *Consensus {
	cns := Consensus{doLog: doLog, Validators: validators, Threshold: threshold, RoundStatus: roundStatus, Chr: chr}

	return &cns
}

// SetReceivedMessage sets the flag which reveals the fact that this node received a message in the current subround
func (cns *Consensus) SetReceivedMessage(receivedMessage bool) {
	cns.mutReceived.Lock()
	defer cns.mutReceived.Unlock()

	cns.receivedMessage = receivedMessage
}

// ReceivedMessage gets the flag which reveals if this node received a message in the current subround
func (cns *Consensus) ReceivedMessage() bool {
	cns.mutReceived.RLock()
	defer cns.mutReceived.RUnlock()

	return cns.receivedMessage
}

// SetSentMessage sets the flag which reveals the fact that this node sent a message in the current subround
func (cns *Consensus) SetSentMessage(sentMessage bool) {
	cns.mutSent.Lock()
	defer cns.mutSent.Unlock()

	cns.sentMessage = sentMessage
}

// SentMessage gets the flag which reveals if this node sent a message in the current subround
func (cns *Consensus) SentMessage() bool {
	cns.mutSent.RLock()
	defer cns.mutSent.RUnlock()

	return cns.sentMessage
}

// CheckConsensus method checks if the consensus is achieved in each subround given from startSubround to endSubround. If the consensus
// is achieved in one subround, the subround status is marked as finished
func (cns *Consensus) CheckConsensus(startSubround chronology.Subround, endSubround chronology.Subround) (bool, int) {
	var n int
	var ok bool

	for i := startSubround; i <= endSubround; i++ {
		switch i {
		case chronology.Subround(SrBlock):
			if cns.RoundStatus.Block != SsFinished {
				if ok, n = cns.IsBlockReceived(cns.Threshold.Block); ok {
					cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) + "Step 1: Subround <BLOCK> has been finished"))
					cns.RoundStatus.Block = SsFinished
				} else {
					return false, n
				}
			}
		case chronology.Subround(SrComitmentHash):
			if cns.RoundStatus.ComitmentHash != SsFinished {
				threshold := cns.Threshold.ComitmentHash
				if !cns.IsNodeLeaderInCurrentRound(cns.Self) {
					threshold = len(cns.ConsensusGroup)
				}
				if ok, n = cns.IsComitmentHashReceived(threshold); ok {
					cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) + "Step 2: Subround <COMITMENT_HASH> has been finished"))
					cns.RoundStatus.ComitmentHash = SsFinished
				} else if ok, n = cns.IsBitmapInComitmentHash(cns.Threshold.Bitmap); ok {
					cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) + "Step 2: Subround <COMITMENT_HASH> has been finished"))
					cns.RoundStatus.ComitmentHash = SsFinished
				} else {
					return false, n
				}
			}
		case chronology.Subround(SrBitmap):
			if cns.RoundStatus.Bitmap != SsFinished {
				if ok, n = cns.IsBitmapInComitmentHash(cns.Threshold.Bitmap); ok {
					cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) + "Step 3: Subround <BITMAP> has been finished"))
					cns.RoundStatus.Bitmap = SsFinished
				} else {
					return false, n
				}
			}
		case chronology.Subround(SrComitment):
			if cns.RoundStatus.Comitment != SsFinished {
				if ok, n = cns.IsBitmapInComitment(cns.Threshold.Comitment); ok {
					cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) + "Step 4: Subround <COMITMENT> has been finished"))
					cns.RoundStatus.Comitment = SsFinished
				} else {
					return false, n
				}
			}
		case chronology.Subround(SrSignature):
			if cns.RoundStatus.Signature != SsFinished {
				if ok, n = cns.IsBitmapInSignature(cns.Threshold.Signature); ok {
					cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) + "Step 5: Subround <SIGNATURE> has been finished"))
					cns.RoundStatus.Signature = SsFinished
				} else {
					return false, n
				}
			}
		default:
			return false, -1
		}
	}

	return true, n
}

// IsNodeLeaderInCurrentRound method checks if the node is leader in the current round
func (cns *Consensus) IsNodeLeaderInCurrentRound(node string) bool {
	leader, err := cns.GetLeader()

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	return leader == node
}

// GetLeader method gets the leader of the current round
func (cns *Consensus) GetLeader() (string, error) {
	if cns.Chr == nil {
		return "", errors.New("chronology is null")
	}

	if cns.Chr.Round() == nil {
		return "", errors.New("round is null")
	}

	if cns.Chr.Round().Index() < 0 {
		return "", errors.New("round index is negative")
	}

	if cns.ConsensusGroup == nil {
		return "", errors.New("consensusGroup is null")
	}

	if len(cns.ConsensusGroup) == 0 {
		return "", errors.New("consensusGroup is empty")
	}

	index := cns.Chr.Round().Index() % len(cns.ConsensusGroup)
	return cns.ConsensusGroup[index], nil
}

// Log method prints info about consensus (if doLog is true)
func (cns *Consensus) Log(message string) {
	if cns.doLog {
		fmt.Printf(message + "\n")
	}
}
