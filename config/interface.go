package config

// NodesSetupHandler provides nodes setup information
type NodesSetupHandler interface {
	MinNumberOfNodesWithHysteresis() uint32
	NumberOfShards() uint32
}
