package spos

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// RoundConsensus defines the data needed by spos to do the consensus in each round
type RoundConsensus struct {
	consensusGroup       []string
	selfPubKey           string
	validatorRoundStates map[string]*RoundState
	mut                  sync.RWMutex
}

// ConsensusGroup returns the consensus group ID's
func (rCns *RoundConsensus) ConsensusGroup() []string {
	return rCns.consensusGroup
}

// IndexSelfConsensusGroup returns the index of self public key in current consensus group
func (rCns *RoundConsensus) IndexSelfConsensusGroup() (int, error) {
	for i, pubKey := range rCns.consensusGroup {
		if pubKey == rCns.selfPubKey {
			return i, nil
		}
	}
	return 0, ErrSelfNotFoundInConsensus
}

// SelfPubKey returns selfPubKey ID
func (rCns *RoundConsensus) SelfPubKey() string {
	return rCns.selfPubKey
}

// SetSelfPubKey sets selfPubKey ID
func (rCns *RoundConsensus) SetSelfPubKey(selfPubKey string) {
	rCns.selfPubKey = selfPubKey
}

// NewRoundConsensus creates a new RoundConsensus object
func NewRoundConsensus(
	consensusGroup []string,
	selfId string,
) *RoundConsensus {

	v := RoundConsensus{
		consensusGroup: consensusGroup,
		selfPubKey:     selfId,
	}

	v.validatorRoundStates = make(map[string]*RoundState)

	for i := 0; i < len(consensusGroup); i++ {
		v.validatorRoundStates[v.consensusGroup[i]] = NewRoundState()
	}

	return &v
}

// GetJobDone returns the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (rCns *RoundConsensus) GetJobDone(key string, subroundId chronology.SubroundId) bool {
	rCns.mut.RLock()
	retcode := rCns.validatorRoundStates[key].JobDone(subroundId)
	rCns.mut.RUnlock()
	return retcode
}

// SetJobDone set the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (rCns *RoundConsensus) SetJobDone(key string, subroundId chronology.SubroundId, value bool) {
	rCns.mut.Lock()
	rCns.validatorRoundStates[key].SetJobDone(subroundId, value)
	rCns.mut.Unlock()
}

// ResetRoundState method resets the state of each node from the current jobDone group, regarding to the
// consensus validatorRoundStates
func (rCns *RoundConsensus) ResetRoundState() {
	for i := 0; i < len(rCns.consensusGroup); i++ {
		rCns.mut.Lock()
		rCns.validatorRoundStates[rCns.consensusGroup[i]].ResetJobsDone()
		rCns.mut.Unlock()
	}
}

// IsValidatorInBitmap method checks if the node is part of the bitmap received from leader
func (rCns *RoundConsensus) IsValidatorInBitmap(validator string) bool {
	return rCns.GetJobDone(validator, SrBitmap)
}

// IsNodeInConsensusGroup method checks if the node is part of the jobDone group of the current round
func (rCns *RoundConsensus) IsNodeInConsensusGroup(node string) bool {
	for i := 0; i < len(rCns.consensusGroup); i++ {
		if rCns.consensusGroup[i] == node {
			return true
		}
	}

	return false
}

// IsBlockReceived method checks if the block was received from the leader in the current round
func (rCns *RoundConsensus) IsBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		if rCns.GetJobDone(rCns.consensusGroup[i], SrBlock) {
			n++
		}
	}

	return n >= threshold
}

// IsCommitmentHashReceived method checks if the commitment hashes from the nodes, belonging to the current jobDone
// group, was received in current round
func (rCns *RoundConsensus) IsCommitmentHashReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		if rCns.GetJobDone(rCns.consensusGroup[i], SrCommitmentHash) {
			n++
		}
	}

	return n >= threshold
}

// CommitmentHashesCollected method checks if the commitment hashes received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (rCns *RoundConsensus) CommitmentHashesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		if rCns.GetJobDone(rCns.consensusGroup[i], SrBitmap) {
			if !rCns.GetJobDone(rCns.consensusGroup[i], SrCommitmentHash) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// CommitmentsCollected method checks if the commitments received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (rCns *RoundConsensus) CommitmentsCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		if rCns.GetJobDone(rCns.consensusGroup[i], SrBitmap) {
			if !rCns.GetJobDone(rCns.consensusGroup[i], SrCommitment) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// SignaturesCollected method checks if the signatures received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (rCns *RoundConsensus) SignaturesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		if rCns.GetJobDone(rCns.consensusGroup[i], SrBitmap) {
			if !rCns.GetJobDone(rCns.consensusGroup[i], SrSignature) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// ComputeSize method returns the number of messages received from the nodes belonging to the current jobDone group
// related to this subround
func (rCns *RoundConsensus) ComputeSize(subroundId chronology.SubroundId) int {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		if rCns.GetJobDone(rCns.consensusGroup[i], subroundId) {
			n++
		}
	}

	return n
}
