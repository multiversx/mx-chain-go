package sharding

// test exports for genesis private process functions

func (g *Genesis) ProcessConfig() error {
	return g.processConfig()
}

func (g *Genesis) ProcessShardAssignment() {
	g.processShardAssignment()
}

func (g *Genesis) CreateInitialNodesPubKeys() {
	g.createInitialNodesPubKeys()
}
