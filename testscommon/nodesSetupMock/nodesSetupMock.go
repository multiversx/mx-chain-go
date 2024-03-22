package nodesSetupMock

// NodesSetupMock -
type NodesSetupMock struct {
	NumberOfShardsField        uint32
	HysteresisField            float32
	MinNumberOfMetaNodesField  uint32
	MinNumberOfShardNodesField uint32
}

// NumberOfShards -
func (n *NodesSetupMock) NumberOfShards() uint32 {
	return n.NumberOfShardsField
}

// GetHysteresis -
func (n *NodesSetupMock) GetHysteresis() float32 {
	return n.HysteresisField
}

// MinNumberOfMetaNodes -
func (n *NodesSetupMock) MinNumberOfMetaNodes() uint32 {
	return n.MinNumberOfMetaNodesField
}

// MinNumberOfShardNodes -
func (n *NodesSetupMock) MinNumberOfShardNodes() uint32 {
	return n.MinNumberOfShardNodesField
}

// MinNumberOfNodes -
func (n *NodesSetupMock) MinNumberOfNodes() uint32 {
	return n.NumberOfShardsField*n.MinNumberOfShardNodesField + n.MinNumberOfMetaNodesField
}

// MinNumberOfNodesWithHysteresis -
func (n *NodesSetupMock) MinNumberOfNodesWithHysteresis() uint32 {
	hystNodesMeta := getHysteresisNodes(n.MinNumberOfMetaNodesField, n.HysteresisField)
	hystNodesShard := getHysteresisNodes(n.MinNumberOfShardNodesField, n.HysteresisField)
	minNumberOfNodes := n.MinNumberOfNodes()

	return minNumberOfNodes + hystNodesMeta + n.NumberOfShardsField*hystNodesShard
}

func getHysteresisNodes(minNumNodes uint32, hysteresis float32) uint32 {
	return uint32(float32(minNumNodes) * hysteresis)
}
