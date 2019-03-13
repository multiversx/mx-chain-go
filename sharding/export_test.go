package sharding

// test exports for genesis private process functions

func (g *Genesis) ProcessConfig() {
	g.processConfig()
}

func (g *Genesis) ProcessShardAssigment() {
	g.processShardAssigment()
}

func (g *Genesis) CreateInitialNodesPubKeys() {
	g.createInitialNodesPubKeys()
}

func (g *Genesis) CreateInitialNodesAddress() {
	g.createInitialNodesAddress()
}
