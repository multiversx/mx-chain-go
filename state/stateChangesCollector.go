package state

// DataTrieChange represents a change in the data trie
type DataTrieChange struct {
	Key []byte
	Val []byte
}

// StateChangeDTO is used to collect state changes
type StateChangeDTO struct {
	MainTrieKey     []byte
	MainTrieVal     []byte
	DataTrieChanges []DataTrieChange
}

// StateChangesForTx is used to collect state changes for a transaction hash
type StateChangesForTx struct {
	TxHash       []byte
	StateChanges []StateChangeDTO
}

type stateChangesCollector struct {
	stateChanges      []StateChangeDTO
	stateChangesForTx []StateChangesForTx
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
func (scc *stateChangesCollector) GetStateChanges() []StateChangesForTx {
	if len(scc.stateChanges) > 0 {
		scc.AddTxHashToCollectedStateChanges([]byte{})
	}

	return scc.stateChangesForTx
}

// Reset resets the state changes collector
func (scc *stateChangesCollector) Reset() {
	scc.stateChanges = make([]StateChangeDTO, 0)
	scc.stateChangesForTx = make([]StateChangesForTx, 0)
}

func (scc *stateChangesCollector) AddTxHashToCollectedStateChanges(txHash []byte) {
	stateChangesForTx := StateChangesForTx{
		TxHash:       txHash,
		StateChanges: scc.stateChanges,
	}

	scc.stateChanges = make([]StateChangeDTO, 0)
	scc.stateChangesForTx = append(scc.stateChangesForTx, stateChangesForTx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scc *stateChangesCollector) IsInterfaceNil() bool {
	return scc == nil
}
