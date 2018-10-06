package validators

type Validators struct {
	WaitingList    []string
	EligibleList   []string
	ConsensusGroup []string
	Self           string
}

func New(consensusGroup []string) Validators {
	v := Validators{}

	v.ConsensusGroup = make([]string, len(consensusGroup))

	for i := 0; i < len(consensusGroup); i++ {
		v.ConsensusGroup[i] = consensusGroup[i]
	}

	return v
}
