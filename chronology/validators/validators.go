package validators

type Validators struct {
	WaitingList    []string
	EligibleList   []string
	ConsensusGroup []string
	Self           string
	ValidationMap  map[string]RoundStateValidation
}

type RoundStateValidation struct {
	Block         bool
	ComitmentHash bool
	Bitmap        bool
	Comitment     bool
	Signature     bool
}

func New(consensusGroup []string, self string) Validators {
	var v Validators

	v.ConsensusGroup = make([]string, len(consensusGroup))

	for i := 0; i < len(consensusGroup); i++ {
		v.ConsensusGroup[i] = consensusGroup[i]
	}

	v.Self = self

	v.ValidationMap = make(map[string]RoundStateValidation)
	v.ResetValidationMap()

	return v
}

func (v *Validators) ResetValidationMap() {
	for i := 0; i < len(v.ConsensusGroup); i++ {
		v.ValidationMap[v.ConsensusGroup[i]] = RoundStateValidation{false, false, false, false, false}
	}
}
