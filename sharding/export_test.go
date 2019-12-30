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

func (ns *NodesSetup) ProcessConfig() error {
	return ns.processConfig()
}

func (ns *NodesSetup) ProcessShardAssignment() {
	ns.processShardAssignment()
}

func (ns *NodesSetup) ProcessMetaChainAssigment() {
	ns.processMetaChainAssigment()
}

func (ns *NodesSetup) CreateInitialNodesInfo() {
	ns.createInitialNodesInfo()
}

func CommunicationIdentifierBetweenShards(shardId1 uint32, shardId2 uint32) string {
	return communicationIdentifierBetweenShards(shardId1, shardId2)
}

func (ihgs *indexHashedNodesCoordinator) EligibleList() []Validator {
	return ihgs.nodesMap[ihgs.shardId]
}

func (ihgs *indexHashedNodesCoordinatorWithRater) ExpandEligibleList(shardId uint32) []Validator {
	return ihgs.expandEligibleList(shardId)
}
