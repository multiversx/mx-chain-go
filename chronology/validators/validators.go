package validators

type Validators struct {
	WaitingList    []string
	EligibleList   []string
	ConsensusGroup []string
	Self           string
}

func (v *Validators) SetConsensusGroup(consensusGroup []string) {
	v.ConsensusGroup = make([]string, len(consensusGroup))

	for i := 0; i < len(consensusGroup); i++ {
		v.ConsensusGroup[i] = consensusGroup[i]
	}
}

func (v *Validators) GetConsensusGroup() []string {
	return v.ConsensusGroup
}

func (v *Validators) SetSelf(self string) {
	v.Self = self
}

func (v *Validators) GetSelf() string {
	return v.Self
}
