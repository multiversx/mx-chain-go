package config

import "github.com/multiversx/mx-chain-communication-go/p2p/config"

// P2PConfig will hold all the P2P settings
type P2PConfig = config.P2PConfig

// P2PTransportConfig will hold the P2P transports config
type P2PTransportConfig = config.TransportConfig

// P2PTCPTransport will hold the P2P TCP transport config
type P2PTCPTransport = config.TCPProtocolConfig

// NodeConfig will hold basic p2p settings
type NodeConfig = config.NodeConfig

// KadDhtPeerDiscoveryConfig will hold the kad-dht discovery config settings
type KadDhtPeerDiscoveryConfig = config.KadDhtPeerDiscoveryConfig

// ShardingConfig will hold the network sharding config settings
type ShardingConfig = config.ShardingConfig
