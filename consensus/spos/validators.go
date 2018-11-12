package spos

import (
	"sync"
)

// RoundValidation defines the data needed by spos to know the state of each node from the current validation group,
// regarding to the consensus agreement in each subround of the current round
type RoundValidation struct {
	validation map[Subround]bool
}

// NewRoundValidation creates a new RoundValidation object
func NewRoundValidation() *RoundValidation {
	rv := RoundValidation{}
	rv.validation = make(map[Subround]bool)
	return &rv
}

// ResetRoundValidation method resets the consensus agreement of each subround
func (rv *RoundValidation) ResetRoundValidation() {
	for k := range rv.validation {
		rv.validation[k] = false
	}
}

// Validation returns the consemsus agreement of the given subround
func (rv *RoundValidation) Validation(subround Subround) bool {
	return rv.validation[subround]
}

// SetValidation sets the consensus agreement of the given subround
func (rv *RoundValidation) SetValidation(subround Subround, value bool) {
	rv.validation[subround] = value
}

// Validators defines the data needed by spos to do the consensus in each round
type Validators struct {
	waitingList    []string
	eligibleList   []string
	consensusGroup []string
	self           string
	agreement      map[string]*RoundValidation
	mut            sync.RWMutex
}

// ConsensusGroup returns the consensus group ID's
func (vld *Validators) ConsensusGroup() []string {
	return vld.consensusGroup
}

// Self returns self ID
func (vld *Validators) Self() string {
	return vld.self
}

// SetSelf sets self ID
func (vld *Validators) SetSelf(self string) {
	vld.self = self
}

// NewValidators creates a new Validators object
func NewValidators(waitingList []string,
	eligibleList []string,
	consensusGroup []string,
	self string) *Validators {

	v := Validators{waitingList: waitingList,
		eligibleList:   eligibleList,
		consensusGroup: consensusGroup,
		self:           self}

	v.agreement = make(map[string]*RoundValidation)

	for i := 0; i < len(consensusGroup); i++ {
		v.agreement[v.consensusGroup[i]] = NewRoundValidation()
	}

	return &v
}

// Agreement returns the state of the action done, by the node represented by the key parameter,
// in subround given by the subround parameter
func (vld *Validators) Agreement(key string, subround Subround) bool {
	return vld.agreement[key].Validation(subround)
}

// SetAgreement set the state of the action done, by the node represented by the key parameter,
// in subround given by the subround parameter
func (vld *Validators) SetAgreement(key string, subround Subround, value bool) {
	vld.mut.Lock()
	vld.agreement[key].SetValidation(subround, value)
	vld.mut.Unlock()
}

// ResetAgreement method resets the state of each node from the current validation group, regarding to the
// consensus agreement
func (vld *Validators) ResetAgreement() {
	for i := 0; i < len(vld.consensusGroup); i++ {
		vld.agreement[vld.consensusGroup[i]].ResetRoundValidation()
	}
}

// IsNodeInBitmapGroup method checks if the node is part of the bitmap received from leader
func (vld *Validators) IsNodeInBitmapGroup(node string) bool {
	return vld.Agreement(node, SrBitmap)
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
		if vld.Agreement(vld.consensusGroup[i], SrBlock) {
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
		if vld.Agreement(vld.consensusGroup[i], SrCommitmentHash) {
			n++
		}
	}

	return n >= threshold
}

// IsBitmapInCommitmentHash method checks if the commitment hashes received from the nodes, belonging to the current
// validation group, are covering the bitmap received from the leader in the current round
func (vld *Validators) IsBitmapInCommitmentHash(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.Agreement(vld.consensusGroup[i], SrBitmap) {
			if !vld.Agreement(vld.consensusGroup[i], SrCommitmentHash) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// IsBitmapInCommitment method checks if the commitments received from the nodes, belonging to the current
// validation group, are covering the bitmap received from the leader in the current round
func (vld *Validators) IsBitmapInCommitment(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.Agreement(vld.consensusGroup[i], SrBitmap) {
			if !vld.Agreement(vld.consensusGroup[i], SrCommitment) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// IsBitmapInSignature method checks if the signatures received from the nodes, belonging to the current
// validation group, are covering the bitmap received from the leader in the current round
func (vld *Validators) IsBitmapInSignature(threshold int) bool {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.Agreement(vld.consensusGroup[i], SrBitmap) {
			if !vld.Agreement(vld.consensusGroup[i], SrSignature) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// ComputeSize method returns the number of messages received from the nodes belonging to the current validation group
// related to this subround
func (vld *Validators) ComputeSize(subround Subround) int {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.Agreement(vld.consensusGroup[i], subround) {
			n++
		}
	}

	return n
}
