package resolverscontainer

// NumCrossShardPeers -
func (brcf *baseResolversContainerFactory) NumCrossShardPeers() int {
	return brcf.numCrossShardPeers
}

// NumTotalPeers -
func (brcf *baseResolversContainerFactory) NumTotalPeers() int {
	return brcf.numTotalPeers
}

// NumFullHistoryPeers -
func (brcf *baseResolversContainerFactory) NumFullHistoryPeers() int {
	return brcf.numFullHistoryPeers
}
