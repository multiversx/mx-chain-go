package spos

// RoundValidation will be comment in 2030
type RoundValidation struct {
	Block         bool
	ComitmentHash bool
	Bitmap        bool
	Comitment     bool
	Signature     bool
}

func NewRoundValidation(block bool, comitmentHash bool, bitmap bool, comitment bool, signature bool) *RoundValidation {
	rv := RoundValidation{Block: block, ComitmentHash: comitmentHash, Bitmap: bitmap, Comitment: comitment, Signature: signature}
	return &rv
}

type Validators struct {
	WaitingList    []string
	EligibleList   []string
	ConsensusGroup []string
	Self           string
	ValidationMap  map[string]RoundValidation
}

func NewValidators(consensusGroup []string, self string) *Validators {
	var v Validators

	v.ConsensusGroup = make([]string, len(consensusGroup))

	for i := 0; i < len(consensusGroup); i++ {
		v.ConsensusGroup[i] = consensusGroup[i]
	}

	v.Self = self

	v.ValidationMap = make(map[string]RoundValidation)
	v.ResetValidationMap()

	return &v
}

func (vld *Validators) ResetValidationMap() {
	for i := 0; i < len(vld.ConsensusGroup); i++ {
		vld.ValidationMap[vld.ConsensusGroup[i]] = RoundValidation{false, false, false, false, false}
	}
}

func (vld *Validators) IsNodeInBitmapGroup(node string) bool {
	return vld.ValidationMap[node].Bitmap
}

func (vld *Validators) IsNodeInValidationGroup(node string) bool {
	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ConsensusGroup[i] == node {
			return true
		}
	}

	return false
}

func (vld *Validators) IsBlockReceived(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].Block {
			n++
		}
	}

	return n >= threshold, n
}

func (vld *Validators) IsComitmentHashReceived(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].ComitmentHash {
			n++
		}
	}

	return n >= threshold, n
}

func (vld *Validators) IsComitmentHashInBitmap(threshold int) (bool, int) {
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

func (vld *Validators) IsComitmentInSignature(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].Comitment {
			if !vld.ValidationMap[vld.ConsensusGroup[i]].Signature {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

func (vld *Validators) GetComitmentHashesCount() int {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].ComitmentHash {
			n++
		}
	}

	return n
}

func (vld *Validators) GetComitmentsCount() int {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].Comitment {
			n++
		}
	}

	return n
}

func (vld *Validators) GetSignaturesCount() int {
	n := 0

	for i := 0; i < len(vld.ConsensusGroup); i++ {
		if vld.ValidationMap[vld.ConsensusGroup[i]].Signature {
			n++
		}
	}

	return n
}
