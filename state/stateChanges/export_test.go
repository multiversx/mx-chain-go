package stateChanges

// GetStateChanges -
func (scc *stateChangesCollector) GetStateChanges() []StateChangesForTx {
	scs, _ := scc.GetStateChangesForTxs()
	return scs
}
