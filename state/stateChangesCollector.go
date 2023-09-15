package state

// StateChangeDTO is used to collect state changes
type StateChangeDTO struct {
	MainTrieKey     []byte
	MainTrieVal     []byte
	DataTrieChanges []DataTrieChange
}

// DataTrieChange represents a change in the data trie
type DataTrieChange struct {
	Key []byte
	Val []byte
}

type stateChangesCollector struct {
	stateChanges []StateChangeDTO
}

// NewStateChangesCollector creates a new StateChangesCollector
func NewStateChangesCollector() *stateChangesCollector {
	return &stateChangesCollector{
		stateChanges: []StateChangeDTO{},
	}
}

// AddStateChange adds a new state change to the collector
func (scc *stateChangesCollector) AddStateChange(stateChange StateChangeDTO) {
	scc.stateChanges = append(scc.stateChanges, stateChange)
}

// GetStateChanges returns the accumulated state changes
func (scc *stateChangesCollector) GetStateChanges() []StateChangeDTO {
	return scc.stateChanges
}

// Reset resets the state changes collector
func (scc *stateChangesCollector) Reset() {
	scc.stateChanges = []StateChangeDTO{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (scc *stateChangesCollector) IsInterfaceNil() bool {
	return scc == nil
}
