package sharding

func (msc *multiShardCoordinator) CalculateMasks() (uint32, uint32) {
	return msc.calculateMasks()
}

func (msc *multiShardCoordinator) Masks() (uint32, uint32) {
	return msc.maskHigh, msc.maskLow
}

func (g *Genesis) ProcessConfig() error {
	return g.processConfig()
}

func (n *Nodes) ProcessConfig() error {
	return n.processConfig()
}

func (n *Nodes) ProcessShardAssignment() {
	n.processShardAssignment()
}

func (n *Nodes) ProcessMetaChainAssigment() {
	n.processMetaChainAssigment()
}

func (n *Nodes) CreateInitialNodesPubKeys() {
	n.createInitialNodesPubKeys()
}

func CommunicationIdentifierBetweenShards(shardId1 uint32, shardId2 uint32) string {
	return communicationIdentifierBetweenShards(shardId1, shardId2)
}
