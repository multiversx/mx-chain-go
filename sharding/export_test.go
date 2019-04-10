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

func (g *Genesis) ProcessShardAssignment() {
	g.processShardAssignment()
}

func (g *Genesis) CreateInitialNodesPubKeys() {
	g.createInitialNodesPubKeys()
}

func CommunicationIdentifierBetweenShards(shardId1 uint32, shardId2 uint32) string {
	return communicationIdentifierBetweenShards(shardId1, shardId2)
}
