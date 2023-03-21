package config

// NodesSetupHandler provides nodes setup information
type NodesSetupHandler interface {
	MinNumberOfNodesWithHysteresis() uint32
	MinNumberOfNodes() uint32
	MinNumberOfShardNodes() uint32
	MinNumberOfMetaNodes() uint32
	GetHysteresis() float32
	NumberOfShards() uint32
}
