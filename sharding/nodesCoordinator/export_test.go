package nodesCoordinator

// AddDummyEpoch adds the epoch in the cached map
func (ihnc *indexHashedNodesCoordinator) AddDummyEpoch(epoch uint32) {
	ihnc.mutNodesConfig.Lock()
	defer ihnc.mutNodesConfig.Unlock()

	ihnc.nodesConfig[epoch] = &epochNodesConfig{}
}
