package state

type StateChangeDTO struct {
	MainTrieKey     []byte
	MainTrieVal     []byte
	DataTrieChanges []DataTrieChange
}

type DataTrieChange struct {
	Key []byte
	Val []byte
}

type stateChangesCollector struct {
	stateChanges []StateChangeDTO
}

func NewStateChangesCollector() *stateChangesCollector {
	return &stateChangesCollector{
		stateChanges: []StateChangeDTO{},
	}
}

func (scc *stateChangesCollector) AddStateChange(stateChange StateChangeDTO) {
	scc.stateChanges = append(scc.stateChanges, stateChange)
}

func (scc *stateChangesCollector) GetStateChanges() []StateChangeDTO {
	return scc.stateChanges
}

func (scc *stateChangesCollector) Reset() {
	scc.stateChanges = []StateChangeDTO{}
}

func (scc *stateChangesCollector) IsInterfaceNil() bool {
	return scc == nil
}
