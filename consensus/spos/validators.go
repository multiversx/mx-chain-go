package spos

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// RoundValidation defines the data needed by spos to know the state of each node from the current validation group,
// regarding to the consensus agreement in each subround of the current round
type RoundValidation struct {
	validation map[chronology.SubroundId]bool
	mut        sync.RWMutex
}

// NewRoundValidation creates a new RoundValidation object
func NewRoundValidation() *RoundValidation {
	rv := RoundValidation{}
	rv.validation = make(map[chronology.SubroundId]bool)
	return &rv
}

// ResetRoundValidation method resets the consensus agreement of each subround
func (rv *RoundValidation) ResetRoundValidation() {
	for k := range rv.validation {
		rv.mut.Lock()
		rv.validation[k] = false
		rv.mut.Unlock()
	}
}

// Validation returns the consensus agreement of the given subroundId
func (rv *RoundValidation) Validation(subroundId chronology.SubroundId) bool {
	rv.mut.RLock()
	retcode := rv.validation[subroundId]
	rv.mut.RUnlock()
	return retcode
}

// SetValidation sets the consensus agreement of the given subroundId
func (rv *RoundValidation) SetValidation(subroundId chronology.SubroundId, value bool) {
	rv.mut.Lock()
	rv.validation[subroundId] = value
	rv.mut.Unlock()
}

// Validators defines the data needed by spos to do the consensus in each round
type Validators struct {
	waitingList    []string
	eligibleList   []string
	consensusGroup []string
	selfId         string
	agreement      map[string]*RoundValidation
	mut            sync.RWMutex
}

// ConsensusGroup returns the consensus group ID's
func (vld *Validators) ConsensusGroup() []string {
	return vld.consensusGroup
}

// SelfId returns selfId ID
func (vld *Validators) SelfId() string {
	return vld.selfId
}

// SetSelfId sets selfId ID
func (vld *Validators) SetSelfId(selfId string) {
	vld.selfId = selfId
}

// NewValidators creates a new Validators object
func NewValidators(
	waitingList []string,
	eligibleList []string,
	consensusGroup []string,
	selfId string,
) *Validators {

	v := Validators{
		waitingList:    waitingList,
		eligibleList:   eligibleList,
		consensusGroup: consensusGroup,
		selfId:         selfId,
	}

	v.agreement = make(map[string]*RoundValidation)

	for i := 0; i < len(consensusGroup); i++ {
		v.agreement[v.consensusGroup[i]] = NewRoundValidation()
	}

	return &v
}

// GetValidation returns the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (vld *Validators) GetValidation(key string, subroundId chronology.SubroundId) bool {
	vld.mut.RLock()
	retcode := vld.agreement[key].Validation(subroundId)
	vld.mut.RUnlock()
	return retcode
}

// SetValidation set the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (vld *Validators) SetValidation(key string, subroundId chronology.SubroundId, value bool) {
	vld.mut.Lock()
	vld.agreement[key].SetValidation(subroundId, value)
	vld.mut.Unlock()
}

// ResetValidation method resets the state of each node from the current validation group, regarding to the
// consensus agreement
func (vld *Validators) ResetValidation() {
	for i := 0; i < len(vld.consensusGroup); i++ {
		vld.mut.Lock()
		vld.agreement[vld.consensusGroup[i]].ResetRoundValidation()
		vld.mut.Unlock()
	}
}

// IsNodeInBitmapGroup method checks if the node is part of the bitmap received from leader
func (vld *Validators) IsNodeInBitmapGroup(node string) bool {
	return vld.GetValidation(node, SrBitmap)
}

// IsNodeInValidationGroup method checks if the node is part of the validation group of the current round
func (vld *Validators) IsNodeInValidationGroup(node string) bool {
	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.consensusGroup[i] == node {
			return true
		}
	}

	return false
}

// IsBlockReceived method checks if the block was received from the leader in the curent round
func (vld *Validators) IsBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetValidation(vld.consensusGroup[i], SrBlock) {
			n++
		}
	}

	return n >= threshold
}

// IsCommitmentHashReceived method checks if the commitment hashes from the nodes, belonging to the current validation
// group, was received in current round
func (vld *Validators) IsCommitmentHashReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetValidation(vld.consensusGroup[i], SrCommitmentHash) {
			n++
		}
	}

	return n >= threshold
}

// CommitmentHashesCollected method checks if the commitment hashes received from the nodes, belonging to the current
// validation group, are covering the bitmap received from the leader in the current round
func (vld *Validators) CommitmentHashesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetValidation(vld.consensusGroup[i], SrBitmap) {
			if !vld.GetValidation(vld.consensusGroup[i], SrCommitmentHash) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// CommitmentsCollected method checks if the commitments received from the nodes, belonging to the current
// validation group, are covering the bitmap received from the leader in the current round
func (vld *Validators) CommitmentsCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetValidation(vld.consensusGroup[i], SrBitmap) {
			if !vld.GetValidation(vld.consensusGroup[i], SrCommitment) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// SignaturesCollected method checks if the signatures received from the nodes, belonging to the current
// validation group, are covering the bitmap received from the leader in the current round
func (vld *Validators) SignaturesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetValidation(vld.consensusGroup[i], SrBitmap) {
			if !vld.GetValidation(vld.consensusGroup[i], SrSignature) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// ComputeSize method returns the number of messages received from the nodes belonging to the current validation group
// related to this subround
func (vld *Validators) ComputeSize(subroundId chronology.SubroundId) int {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.GetValidation(vld.consensusGroup[i], subroundId) {
			n++
		}
	}

	return n
}
