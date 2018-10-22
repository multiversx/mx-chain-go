package spos

// RoundValidation defines the data needed by spos to know the state of each node from the current validation group,
// regarding to the consensus agreement in each subround of the current round
type RoundValidation struct {
	Block         bool
	ComitmentHash bool
	Bitmap        bool
	Comitment     bool
	Signature     bool
}

// NewRoundValidation creates a new RoundValidation object
func NewRoundValidation(block bool, comitmentHash bool, bitmap bool, comitment bool, signature bool) *RoundValidation {
	rv := RoundValidation{Block: block, ComitmentHash: comitmentHash, Bitmap: bitmap, Comitment: comitment, Signature: signature}
	return &rv
}

// ResetRoundValidation method resets the consensus agreement of each subround
func (rv *RoundValidation) ResetRoundValidation() {
	rv.Block = false
	rv.ComitmentHash = false
	rv.Bitmap = false
	rv.Comitment = false
	rv.Signature = false
}

// Validators defines the data needed by spos to do the consensus in each round
type Validators struct {
	WaitingList    []string
	EligibleList   []string
	ConsensusGroup []string
	Self           string
	ValidationMap  map[string]*RoundValidation
}

// NewValidators creates a new Validators object
func NewValidators(waitingList []string, eligibleList []string, consensusGroup []string, self string) *Validators {
	v := Validators{WaitingList: waitingList, EligibleList: eligibleList, ConsensusGroup: consensusGroup, Self: self}

	v.ValidationMap = make(map[string]*RoundValidation)

	for i := 0; i < len(consensusGroup); i++ {
		v.ValidationMap[v.ConsensusGroup[i]] = NewRoundValidation(false, false, false, false, false)
	}

	return &v
}

// ResetValidationMap method resets the state of each node from the current validation group, regarding to the consensus agreement
func (vld *Validators) ResetValidationMap() {
	for i := 0; i < len(vld.ConsensusGroup); i++ {
		vld.ValidationMap[vld.ConsensusGroup[i]].ResetRoundValidation()
	}
}

// IsNodeInBitmapGroup method checks if the node is part of the bitmap received from leader
func (vld *Validators) IsNodeInBitmapGroup(node string) bool {
	return vld.ValidationMap[node].Bitmap
}

// IsNodeInValidationGroup method checks if the node is part of the validation group of the current round
func (vld *Validators) IsNodeInValidationGroup(node string) bool {
	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ConsensusGroup[i] == node {
			return true
		}
	}

	return false
}

// IsBlockReceived method checks if the block was received from the leader in the curent round
func (vld *Validators) IsBlockReceived(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].Block {
			n++
		}
	}

	return n >= threshold, n
}

// IsComitmentHashReceived method checks if the comitment hashes from the nodes, belonging to the current validation group, was received in current round
func (vld *Validators) IsComitmentHashReceived(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].ComitmentHash {
			n++
		}
	}

	return n >= threshold, n
}

// IsBitmapInComitmentHash method checks if the comitment hashes received from the nodes, belonging to the current validation group, are covering the bitmap
// received from the leader in the current round
func (vld *Validators) IsBitmapInComitmentHash(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].Bitmap {
			if !vld.ValidationMap[vld.ConsensusGroup[i]].ComitmentHash {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

// IsBitmapInComitment method checks if the comitments received from the nodes, belonging to the current validation group, are covering the bitmap
// received from the leader in the current round
func (vld *Validators) IsBitmapInComitment(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].Bitmap {
			if !vld.ValidationMap[vld.ConsensusGroup[i]].Comitment {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

// IsBitmapInSignature method checks if the signatures received from the nodes, belonging to the current validation group, are covering the bitmap
// received from the leader in the current round
func (vld *Validators) IsBitmapInSignature(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].Bitmap {
			if !vld.ValidationMap[vld.ConsensusGroup[i]].Signature {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

// GetComitmentHashesCount method returns the number of comitment hashes received from the nodes belonging to the current validation group
func (vld *Validators) GetComitmentHashesCount() int {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].ComitmentHash {
			n++
		}
	}

	return n
}

// GetComitmentsCount method returns the number of comitments received from the nodes belonging to the current validation group
func (vld *Validators) GetComitmentsCount() int {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].Comitment {
			n++
		}
	}

	return n
}

// GetSignaturesCount method returns the number of signatures received from the nodes belonging to the current validation group
func (vld *Validators) GetSignaturesCount() int {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].Signature {
			n++
		}
	}

	return n
}
