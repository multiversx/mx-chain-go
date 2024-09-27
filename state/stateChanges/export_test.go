package stateChanges

// GetStateChanges -
func (scc *stateChangesCollector) GetStateChanges() []StateChangesForTx {
	scs, _ := scc.getStateChangesForTxs()
	return scs
}

// GetStateChanges -
func (dsc *dataAnalysisCollector) GetStateChanges() []dataAnalysisStateChangesForTx {
	scs, _ := dsc.getDataAnalysisStateChangesForTxs()
	return scs
}
