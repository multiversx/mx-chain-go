package stateChanges

// GetStateChanges -
func (c *collector) GetStateChanges() []StateChangesForTx {
	scs, _ := c.getStateChangesForTxs()
	return scs
}

// GetStateChanges -
func (c *collector) GetDataAnalysisStateChanges() []dataAnalysisStateChangesForTx {
	scs, _ := c.getDataAnalysisStateChangesForTxs()
	return scs
}
