package requesterscontainer

// NumCrossShardPeers -
func (brcf *baseRequestersContainerFactory) NumCrossShardPeers() int {
	return brcf.numCrossShardPeers
}

// NumTotalPeers -
func (brcf *baseRequestersContainerFactory) NumTotalPeers() int {
	return brcf.numTotalPeers
}

// NumFullHistoryPeers -
func (brcf *baseRequestersContainerFactory) NumFullHistoryPeers() int {
	return brcf.numFullHistoryPeers
}
