package libp2p

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

// ConnectionMonitor defines the behavior of a connection monitor
type ConnectionMonitor interface {
	network.Notifiee
	SetSharder(sharder p2p.CommonSharder) error
	IsConnectedToTheNetwork(netw network.Network) bool
	SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, netw network.Network)
	ThresholdMinConnectedPeers() int
}

// Sharder defines the eviction computing process of unwanted peers
type Sharder interface {
	ComputeEvictionList(pidList []peer.ID) []peer.ID
	Has(pid peer.ID, list []peer.ID) bool
	PeerShardResolver() p2p.PeerShardResolver
	IsInterfaceNil() bool
}

// PeerDiscovererWithSharder extends the PeerDiscoverer with the possibility to set the sharder
type PeerDiscovererWithSharder interface {
	p2p.PeerDiscoverer
	SetSharder(sharder Sharder) error
}
