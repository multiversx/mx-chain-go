package spos

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// RoundConsensus defines the data needed by spos to do the consensus in each round
type RoundConsensus struct {
	consensusGroup       []string
	selfId               string
	validatorRoundStates map[string]*RoundState
	mut                  sync.RWMutex
}

// ConsensusGroup returns the consensus group ID's
func (vld *RoundConsensus) ConsensusGroup() []string {
	return vld.consensusGroup
}

// SetConsensusGroup sets the consensus group ID's
func (vld *RoundConsensus) SetConsensusGroup(consensusGroup []string) {
	vld.consensusGroup = consensusGroup

	vld.mut.Lock()

	vld.validatorRoundStates = make(map[string]*RoundState)

	for i := 0; i < len(consensusGroup); i++ {
		vld.validatorRoundStates[vld.consensusGroup[i]] = NewRoundState()
	}

	vld.mut.Unlock()
}

// SelfId returns selfId ID
func (vld *RoundConsensus) SelfId() string {
	return vld.selfId
}

// SetSelfId sets selfId ID
func (vld *RoundConsensus) SetSelfId(selfId string) {
	vld.selfId = selfId
}

// NewRoundConsensus creates a new RoundConsensus object
func NewRoundConsensus(
	consensusGroup []string,
	selfId string,
) *RoundConsensus {

	v := RoundConsensus{
		consensusGroup: consensusGroup,
		selfId:         selfId,
	}

	v.validatorRoundStates = make(map[string]*RoundState)

	for i := 0; i < len(consensusGroup); i++ {
		v.validatorRoundStates[v.consensusGroup[i]] = NewRoundState()
	}

	return &v
}

// GetJobDone returns the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (vld *RoundConsensus) GetJobDone(key string, subroundId chronology.SubroundId) bool {
	retcode := false
	vld.mut.RLock()
	if vld.validatorRoundStates[key] != nil {
		retcode = vld.validatorRoundStates[key].JobDone(subroundId)
	}
	vld.mut.RUnlock()
	return retcode
}

// SetJobDone set the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (vld *RoundConsensus) SetJobDone(key string, subroundId chronology.SubroundId, value bool) {
	vld.mut.Lock()
	if vld.validatorRoundStates[key] != nil {
		vld.validatorRoundStates[key].SetJobDone(subroundId, value)
	}
	vld.mut.Unlock()
}

// ResetRoundState method resets the state of each node from the current jobDone group, regarding to the
// consensus validatorRoundStates
func (vld *RoundConsensus) ResetRoundState() {
	for i := 0; i < len(vld.consensusGroup); i++ {
		vld.mut.Lock()
		if vld.validatorRoundStates[vld.consensusGroup[i]] != nil {
			vld.validatorRoundStates[vld.consensusGroup[i]].ResetJobsDone()
		}
		vld.mut.Unlock()
	}
}

// IsValidatorInBitmap method checks if the node is part of the bitmap received from leader
func (vld *RoundConsensus) IsValidatorInBitmap(validator string) bool {
	return vld.GetJobDone(validator, SrBitmap)
}

// IsNodeInConsensusGroup method checks if the node is part of the jobDone group of the current round
func (vld *RoundConsensus) IsNodeInConsensusGroup(node string) bool {
	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.consensusGroup[i] == node {
			return true
		}
	}

	return false
}

// IsBlockReceived method checks if the block was received from the leader in the current round
func (vld *RoundConsensus) IsBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetJobDone(vld.consensusGroup[i], SrBlock) {
			n++
		}
	}

	return n >= threshold
}

// IsCommitmentHashReceived method checks if the commitment hashes from the nodes, belonging to the current jobDone
// group, was received in current round
func (vld *RoundConsensus) IsCommitmentHashReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetJobDone(vld.consensusGroup[i], SrCommitmentHash) {
			n++
		}
	}

	return n >= threshold
}

// CommitmentHashesCollected method checks if the commitment hashes received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (vld *RoundConsensus) CommitmentHashesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetJobDone(vld.consensusGroup[i], SrBitmap) {
			if !vld.GetJobDone(vld.consensusGroup[i], SrCommitmentHash) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// CommitmentsCollected method checks if the commitments received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (vld *RoundConsensus) CommitmentsCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetJobDone(vld.consensusGroup[i], SrBitmap) {
			if !vld.GetJobDone(vld.consensusGroup[i], SrCommitment) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// SignaturesCollected method checks if the signatures received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (vld *RoundConsensus) SignaturesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetJobDone(vld.consensusGroup[i], SrBitmap) {
			if !vld.GetJobDone(vld.consensusGroup[i], SrSignature) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// ComputeSize method returns the number of messages received from the nodes belonging to the current jobDone group
// related to this subround
func (vld *RoundConsensus) ComputeSize(subroundId chronology.SubroundId) int {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetJobDone(vld.consensusGroup[i], subroundId) {
			n++
		}
	}

	return n
}
