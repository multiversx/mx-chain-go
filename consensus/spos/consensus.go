package spos

import (
	"errors"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

const roundTimeDuration = time.Duration(4000 * time.Millisecond)
const sleepTime = 5

type Subround int

const (
	srStartRound Subround = iota
	srBlock
	srComitmentHash
	srBitmap
	srComitment
	srSignature
	srEndRound
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

// A SubroundStatus specifies what kind of status could have a subround
type SubroundStatus int

const (
	ssNotFinished SubroundStatus = iota
	ssExtended
	ssFinished
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

	Block      *block.Block
	BlockChain *blockchain.BlockChain
	chr        *chronology.Chronology
	P2PNode    *p2p.Messenger

	ChRcvMsg chan []byte
}

func NewConsensus(doLog bool, validators *Validators, threshold *Threshold, roundStatus *RoundStatus, chr *chronology.Chronology) *Consensus {
	cns := Consensus{doLog: doLog, Validators: validators, Threshold: threshold, RoundStatus: roundStatus, chr: chr}

	cns.ChRcvMsg = make(chan []byte, len(cns.Validators.ConsensusGroup))
	//	(*cns.P2PNode).SetOnRecvMsg(cns.ReceiveMessage)
	return &cns
}

func (cns *Consensus) CheckConsensus(startRoundState chronology.Subround, endRoundState chronology.Subround) (bool, int) {
	var n int
	var ok bool

	for i := startRoundState; i <= endRoundState; i++ {
		switch i {
		case chronology.Subround(srBlock):
			if cns.RoundStatus.Block != ssFinished {
				if ok, n = cns.IsBlockReceived(cns.Threshold.Block); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 1: Subround <BLOCK> has been finished"))
					cns.RoundStatus.Block = ssFinished
				} else {
					return false, n
				}
			}
		case chronology.Subround(srComitmentHash):
			if cns.RoundStatus.ComitmentHash != ssFinished {
				threshold := cns.Threshold.ComitmentHash
				if !cns.IsNodeLeaderInCurrentRound(cns.Self) {
					threshold = len(cns.ConsensusGroup)
				}
				if ok, n = cns.IsComitmentHashReceived(threshold); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 2: Subround <COMITMENT_HASH> has been finished"))
					cns.RoundStatus.ComitmentHash = ssFinished
				} else if ok, n = cns.IsComitmentHashInBitmap(cns.Threshold.Bitmap); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 2: Subround <COMITMENT_HASH> has been finished"))
					cns.RoundStatus.ComitmentHash = ssFinished
				} else {
					return false, n
				}
			}
		case chronology.Subround(srBitmap):
			if cns.RoundStatus.Bitmap != ssFinished {
				if ok, n = cns.IsComitmentHashInBitmap(cns.Threshold.Bitmap); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 3: Subround <BITMAP> has been finished"))
					cns.RoundStatus.Bitmap = ssFinished
				} else {
					return false, n
				}
			}
		case chronology.Subround(srComitment):
			if cns.RoundStatus.Comitment != ssFinished {
				if ok, n = cns.IsBitmapInComitment(cns.Threshold.Comitment); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 4: Subround <COMITMENT> has been finished"))
					cns.RoundStatus.Comitment = ssFinished
				} else {
					return false, n
				}
			}
		case chronology.Subround(srSignature):
			if cns.RoundStatus.Signature != ssFinished {
				if ok, n = cns.IsComitmentInSignature(cns.Threshold.Signature); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 5: Subround <SIGNATURE> has been finished"))
					cns.RoundStatus.Signature = ssFinished
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
	if cns.chr.GetRoundIndex() == -1 {
		return "", errors.New("Round is not set")
	}

	if cns.ConsensusGroup == nil {
		return "", errors.New("List of Validators.ConsensusGroup is null")
	}

	if len(cns.ConsensusGroup) == 0 {
		return "", errors.New("List of nodes is empty")
	}

	index := cns.chr.GetRoundIndex() % len(cns.ConsensusGroup)
	return cns.ConsensusGroup[index], nil
}

func (cns *Consensus) ResetRoundStatus() {
	cns.RoundStatus.Block = ssNotFinished
	cns.RoundStatus.ComitmentHash = ssNotFinished
	cns.RoundStatus.Bitmap = ssNotFinished
	cns.RoundStatus.Comitment = ssNotFinished
	cns.RoundStatus.Signature = ssNotFinished
}

func (cns *Consensus) Log(message string) {
	if cns.doLog {
		fmt.Printf(message + "\n")
	}
}

//func (cns *Consensus) IsNodeLeader(node string, nodes []string, round *chronology.Round) (bool, error) {
//	v, err := cns.ComputeLeader(nodes, round)
//
//	if err != nil {
//		fmt.Println(err)
//		return false, err
//	}
//
//	return v == node, nil
//}
//
//func (cns *Consensus) ComputeLeader(nodes []string, round *chronology.Round) (string, error) {
//	if round == nil {
//		return "", errors.New("Round is null")
//	}
//
//	if nodes == nil {
//		return "", errors.New("List of nodes is null")
//	}
//
//	if len(nodes) == 0 {
//		return "", errors.New("List of nodes is empty")
//	}
//
//	index := cns.chr.GetRoundIndex() % len(nodes)
//	return nodes[index], nil
//}
