package config

// P2PConfig will hold all the P2P settings
type P2PConfig struct {
	Node                NodeConfig
	KadDhtPeerDiscovery KadDhtPeerDiscoveryConfig
	Sharding            ShardingConfig
}

// NodeConfig will hold basic p2p settings
type NodeConfig struct {
	Port 						string
	Seed 						string
	MaximumExpectedPeerCount 	uint64
}

// KadDhtPeerDiscoveryConfig will hold the kad-dht discovery config settings
type KadDhtPeerDiscoveryConfig struct {
	Enabled                          bool
	RefreshIntervalInSec             uint32
	RandezVous                       string
	InitialPeerList                  []string
	BucketSize                       uint32
	RoutingTableRefreshIntervalInSec uint32
}

// ShardingConfig will hold the network sharding config settings
type ShardingConfig struct {
	TargetPeerCount         int
	MaxIntraShardValidators uint32
	MaxCrossShardValidators uint32
	MaxIntraShardObservers  uint32
	MaxCrossShardObservers  uint32
	Type                    string
}
