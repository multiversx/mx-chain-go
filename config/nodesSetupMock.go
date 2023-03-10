package config

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
	hystNodesMeta := getMinNumNodesWithHysteresis(n.MinNumberOfMetaNodesField, n.HysteresisField)
	hystNodesShard := getMinNumNodesWithHysteresis(n.MinNumberOfShardNodesField, n.HysteresisField)
	minNumberOfNodes := n.MinNumberOfNodes()

	return minNumberOfNodes + hystNodesMeta + n.NumberOfShardsField*hystNodesShard
}
