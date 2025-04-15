package stateAccesses

import data "github.com/multiversx/mx-chain-core-go/data/stateChange"

// GetStateChanges -
func (c *collector) GetStateChanges() map[string]*data.StateAccesses {
	return getStateAccessesForTxs(c.stateAccesses)
}
