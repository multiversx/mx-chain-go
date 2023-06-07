package config

import "github.com/multiversx/mx-chain-communication-go/p2p/config"

// P2PConfig will hold all the P2P settings
type P2PConfig = config.P2PConfig

// NodeConfig will hold basic p2p settings
type NodeConfig = config.NodeConfig

// KadDhtPeerDiscoveryConfig will hold the kad-dht discovery config settings
type KadDhtPeerDiscoveryConfig = config.KadDhtPeerDiscoveryConfig

// ShardingConfig will hold the network sharding config settings
type ShardingConfig = config.ShardingConfig

// AdditionalConnectionsConfig will hold the additional connections that will be open when certain conditions are met
// All these values should be added to the maximum target peer count value
type AdditionalConnectionsConfig = config.AdditionalConnectionsConfig
