package spos

import (
	"sync"
)

// RoundValidation defines the data needed by spos to know the state of each node from the current validation group,
// regarding to the consensus agreement in each subround of the current round
type RoundValidation struct {
	block          bool
	commitmentHash bool
	bitmap         bool
	commitment     bool
	signature      bool
}

// NewRoundValidation creates a new RoundValidation object
func NewRoundValidation(block bool,
	commitmentHash bool,
	bitmap bool,
	commitment bool,
	signature bool) *RoundValidation {

	rv := RoundValidation{block: block,
		commitmentHash: commitmentHash,
		bitmap:         bitmap,
		commitment:     commitment,
		signature:      signature}

	return &rv
}

// ResetRoundValidation method resets the consensus agreement of each subround
func (rv *RoundValidation) ResetRoundValidation() {
	rv.block = false
	rv.commitmentHash = false
	rv.bitmap = false
	rv.commitment = false
	rv.signature = false
}

// Validators defines the data needed by spos to do the consensus in each round
type Validators struct {
	waitingList    []string
	eligibleList   []string
	consensusGroup []string
	self           string
	validationMap  map[string]*RoundValidation
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

	v.validationMap = make(map[string]*RoundValidation)

	for i := 0; i < len(consensusGroup); i++ {

		v.validationMap[v.consensusGroup[i]] = NewRoundValidation(false,
			false,
			false,
			false,
			false)
	}

	return &v
}

// ValidationMap returns the state of the action done, by the node represented by the key parameter,
// in subround given by the subround parameter
func (vld *Validators) ValidationMap(key string, subround Subround) bool {
	vld.mut.RLock()
	defer vld.mut.RUnlock()

	switch subround {
	case SrBlock:
		return vld.validationMap[key].block
	case SrCommitmentHash:
		return vld.validationMap[key].commitmentHash
	case SrBitmap:
		return vld.validationMap[key].bitmap
	case SrCommitment:
		return vld.validationMap[key].commitment
	case SrSignature:
		return vld.validationMap[key].signature
	}

	return false
}

// SetValidationMap set the state of the action done, by the node represented by the key parameter,
// in subround given by the subround parameter
func (vld *Validators) SetValidationMap(key string, value bool, subround Subround) {
	vld.mut.Lock()
	defer vld.mut.Unlock()

	switch subround {
	case SrBlock:
		vld.validationMap[key].block = value
	case SrCommitmentHash:
		vld.validationMap[key].commitmentHash = value
	case SrBitmap:
		vld.validationMap[key].bitmap = value
	case SrCommitment:
		vld.validationMap[key].commitment = value
	case SrSignature:
		vld.validationMap[key].signature = value
	}
}

// ResetValidationMap method resets the state of each node from the current validation group, regarding to the
// consensus agreement
func (vld *Validators) ResetValidationMap() {
	for i := 0; i < len(vld.consensusGroup); i++ {
		vld.validationMap[vld.consensusGroup[i]].ResetRoundValidation()
	}
}

// IsNodeInBitmapGroup method checks if the node is part of the bitmap received from leader
func (vld *Validators) IsNodeInBitmapGroup(node string) bool {
	return vld.ValidationMap(node, SrBitmap)
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
		if vld.ValidationMap(vld.consensusGroup[i], SrBlock) {
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
		if vld.ValidationMap(vld.consensusGroup[i], SrCommitmentHash) {
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
		if vld.ValidationMap(vld.consensusGroup[i], SrBitmap) {
			if !vld.ValidationMap(vld.consensusGroup[i], SrCommitmentHash) {
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
		if vld.ValidationMap(vld.consensusGroup[i], SrBitmap) {
			if !vld.ValidationMap(vld.consensusGroup[i], SrCommitment) {
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
		if vld.ValidationMap(vld.consensusGroup[i], SrBitmap) {
			if !vld.ValidationMap(vld.consensusGroup[i], SrSignature) {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// GetBlocksCount method returns the number of blocks received (usually should be only one and it should be received
// from the leader)
func (vld *Validators) GetBlocksCount() int {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.ValidationMap(vld.consensusGroup[i], SrBlock) {
			n++
		}
	}

	return n
}

// GetCommitmentHashesCount method returns the number of commitment hashes received from the nodes belonging to the
// current validation group
func (vld *Validators) GetCommitmentHashesCount() int {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.ValidationMap(vld.consensusGroup[i], SrCommitmentHash) {
			n++
		}
	}

	return n
}

// GetBitmapsCount method returns the number of Validators which are included in the bitmap received from the leader
func (vld *Validators) GetBitmapsCount() int {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.ValidationMap(vld.consensusGroup[i], SrBitmap) {
			n++
		}
	}

	return n
}

// GetCommitmentsCount method returns the number of commitments received from the nodes belonging to the current
// validation group
func (vld *Validators) GetCommitmentsCount() int {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.ValidationMap(vld.consensusGroup[i], SrCommitment) {
			n++
		}
	}

	return n
}

// GetSignaturesCount method returns the number of signatures received from the nodes belonging to the current
// validation group
func (vld *Validators) GetSignaturesCount() int {
	n := 0

	for i := 0; i < len(vld.consensusGroup); i++ {
		if vld.ValidationMap(vld.consensusGroup[i], SrSignature) {
			n++
		}
	}

	return n
}
