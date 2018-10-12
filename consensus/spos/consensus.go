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

const ROUND_TIME_DURATION = time.Duration(4000 * time.Millisecond)
const SLEEP_TIME = 5

type Subround int

const (
	SR_START_ROUND Subround = iota
	SR_BLOCK
	SR_COMITMENT_HASH
	SR_BITMAP
	SR_COMITMENT
	SR_SIGNATURE
	SR_END_ROUND
)

type Response int

const (
	R_False Response = iota
	R_True
	R_None
)

type RoundStatus struct {
	Block         SubroundStatus
	ComitmentHash SubroundStatus
	Bitmap        SubroundStatus
	Comitment     SubroundStatus
	Signature     SubroundStatus
}

func NewRoundStatus(block SubroundStatus, comitmentHash SubroundStatus, bitmap SubroundStatus, comitment SubroundStatus, signature SubroundStatus) RoundStatus {
	rs := RoundStatus{Block: block, ComitmentHash: comitmentHash, Bitmap: bitmap, Comitment: comitment, Signature: signature}
	return rs
}

// A SubroundStatus specifies what kind of status could have a subround
type SubroundStatus int

const (
	SS_NOTFINISHED SubroundStatus = iota
	SS_EXTENDED
	SS_FINISHED
)

type Threshold struct {
	Block         int
	ComitmentHash int
	Bitmap        int
	Comitment     int
	Signature     int
}

func NewThreshold(block int, comitmentHash int, bitmap int, comitment int, signature int) Threshold {
	th := Threshold{Block: block, ComitmentHash: comitmentHash, Bitmap: bitmap, Comitment: comitment, Signature: signature}
	return th
}

type Consensus struct {
	doLog bool

	Validators
	Threshold
	RoundStatus

	Block      *block.Block
	BlockChain *blockchain.BlockChain
	chr        *chronology.Chronology
	P2PNode    *p2p.Messenger

	ChRcvMsg chan []byte
}

func NewConsensus(doLog bool, validators Validators, threshold Threshold, roundStatus RoundStatus) Consensus {
	cns := Consensus{doLog: doLog, Validators: validators, Threshold: threshold, RoundStatus: roundStatus}

	cns.ChRcvMsg = make(chan []byte, len(cns.Validators.ConsensusGroup))
	//(*cns.P2PNode).SetOnRecvMsg(cns.ReceiveMessage)
	return cns
}

func (cns *Consensus) CheckConsensus(startRoundState chronology.Subround, endRoundState chronology.Subround) (bool, int) {
	var n int
	var ok bool

	for i := startRoundState; i <= endRoundState; i++ {
		switch i {
		case chronology.Subround(SR_BLOCK):
			if cns.RoundStatus.Block != SS_FINISHED {
				if ok, n = cns.IsBlockReceived(cns.Threshold.Block); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 1: Subround <BLOCK> has been finished"))
					cns.RoundStatus.Block = SS_FINISHED
				} else {
					return false, n
				}
			}
		case chronology.Subround(SR_COMITMENT_HASH):
			if cns.RoundStatus.ComitmentHash != SS_FINISHED {
				threshold := cns.Threshold.ComitmentHash
				if !cns.IsNodeLeaderInCurrentRound(cns.Self) {
					threshold = len(cns.ConsensusGroup)
				}
				if ok, n = cns.IsComitmentHashReceived(threshold); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 2: Subround <COMITMENT_HASH> has been finished"))
					cns.RoundStatus.ComitmentHash = SS_FINISHED
				} else if ok, n = cns.IsComitmentHashInBitmap(cns.Threshold.Bitmap); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 2: Subround <COMITMENT_HASH> has been finished"))
					cns.RoundStatus.ComitmentHash = SS_FINISHED
				} else {
					return false, n
				}
			}
		case chronology.Subround(SR_BITMAP):
			if cns.RoundStatus.Bitmap != SS_FINISHED {
				if ok, n = cns.IsComitmentHashInBitmap(cns.Threshold.Bitmap); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 3: Subround <BITMAP> has been finished"))
					cns.RoundStatus.Bitmap = SS_FINISHED
				} else {
					return false, n
				}
			}
		case chronology.Subround(SR_COMITMENT):
			if cns.RoundStatus.Comitment != SS_FINISHED {
				if ok, n = cns.IsBitmapInComitment(cns.Threshold.Comitment); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 4: Subround <COMITMENT> has been finished"))
					cns.RoundStatus.Comitment = SS_FINISHED
				} else {
					return false, n
				}
			}
		case chronology.Subround(SR_SIGNATURE):
			if cns.RoundStatus.Signature != SS_FINISHED {
				if ok, n = cns.IsComitmentInSignature(cns.Threshold.Signature); ok {
					cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 5: Subround <SIGNATURE> has been finished"))
					cns.RoundStatus.Signature = SS_FINISHED
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
	cns.RoundStatus = RoundStatus{SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED}
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
