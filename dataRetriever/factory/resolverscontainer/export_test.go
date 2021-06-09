package resolverscontainer

// NumCrossShardPeers -
func (brcf *baseResolversContainerFactory) NumCrossShardPeers() int {
	return brcf.numCrossShardPeers
}

// NumIntraShardPeers -
func (brcf *baseResolversContainerFactory) NumIntraShardPeers() int {
	return brcf.numIntraShardPeers
}

// NumFullHistoryPeers -
func (brcf *baseResolversContainerFactory) NumFullHistoryPeers() int {
	return brcf.numFullHistoryPeers
}
