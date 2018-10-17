package spos

import (
	"errors"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	//"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	//"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

const roundTimeDuration = time.Duration(4000 * time.Millisecond)
const sleepTime = 5

type Subround int

const (
	SrStartRound Subround = iota
	SrBlock
	SrComitmentHash
	SrBitmap
	SrComitment
	SrSignature
	SrEndRound
)

type Response int

const (
	rFalse Response = iota
	rTrue
	rNone
)

type RoundStatus struct {
	Block         SubroundStatus
	ComitmentHash SubroundStatus
	Bitmap        SubroundStatus
	Comitment     SubroundStatus
	Signature     SubroundStatus
}

func NewRoundStatus(block SubroundStatus, comitmentHash SubroundStatus, bitmap SubroundStatus, comitment SubroundStatus, signature SubroundStatus) *RoundStatus {
	rs := RoundStatus{Block: block, ComitmentHash: comitmentHash, Bitmap: bitmap, Comitment: comitment, Signature: signature}
	return &rs
}

func (rs *RoundStatus) ResetRoundStatus() {
	rs.Block = SsNotFinished
	rs.ComitmentHash = SsNotFinished
	rs.Bitmap = SsNotFinished
	rs.Comitment = SsNotFinished
	rs.Signature = SsNotFinished
}

// A SubroundStatus specifies what kind of status could have a subround
type SubroundStatus int

const (
	SsNotFinished SubroundStatus = iota
	SsExtended
	SsFinished
)

type Threshold struct {
	Block         int
	ComitmentHash int
	Bitmap        int
	Comitment     int
	Signature     int
}

func NewThreshold(block int, comitmentHash int, bitmap int, comitment int, signature int) *Threshold {
	th := Threshold{Block: block, ComitmentHash: comitmentHash, Bitmap: bitmap, Comitment: comitment, Signature: signature}
	return &th
}

type Consensus struct {
	doLog bool

	*Validators
	*Threshold
	*RoundStatus

	Chr      *chronology.Chronology
	ChRcvMsg chan []byte

	//block      *block.block
	//blockChain *blockchain.blockChain
}

func NewConsensus(doLog bool, validators *Validators, threshold *Threshold, roundStatus *RoundStatus, chr *chronology.Chronology) *Consensus {
	cns := Consensus{doLog: doLog, Validators: validators, Threshold: threshold, RoundStatus: roundStatus, Chr: chr}

	if cns.Validators == nil || len(cns.Validators.ConsensusGroup) == 0 {
		cns.ChRcvMsg = make(chan []byte)
	} else {
		cns.ChRcvMsg = make(chan []byte, len(cns.Validators.ConsensusGroup))
	}

	return &cns
}

func (cns *Consensus) CheckConsensus(startRoundState chronology.Subround, endRoundState chronology.Subround) (bool, int) {
	var n int
	var ok bool

	for i := startRoundState; i <= endRoundState; i++ {
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

func (cns *Consensus) IsNodeLeaderInCurrentRound(node string) bool {
	leader, err := cns.GetLeader()

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	return leader == node
}

func (cns *Consensus) GetLeader() (string, error) {
	if cns.Chr == nil {
		return "", errors.New("Chronology is null")
	}

	if cns.Chr.Round() == nil {
		return "", errors.New("Round is null")
	}

	if cns.Chr.Round().Index() < 0 {
		return "", errors.New("Round index is negative")
	}

	if cns.ConsensusGroup == nil {
		return "", errors.New("ConsensusGroup is null")
	}

	if len(cns.ConsensusGroup) == 0 {
		return "", errors.New("ConsensusGroup is empty")
	}

	index := cns.Chr.Round().Index() % len(cns.ConsensusGroup)
	return cns.ConsensusGroup[index], nil
}

func (cns *Consensus) Log(message string) {
	if cns.doLog {
		fmt.Printf(message + "\n")
	}
}
