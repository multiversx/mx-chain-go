package connectionMonitor

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// NoConnectionMonitor does not monitor the connections mad by the host
type NoConnectionMonitor struct {
	thresholdMinConnectedPeers int
}

// Listen is called when network starts listening on an addr
func (n *NoConnectionMonitor) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (n *NoConnectionMonitor) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened
func (n *NoConnectionMonitor) Connected(network.Network, network.Conn) {}

// Disconnected is called when a connection closed
func (n *NoConnectionMonitor) Disconnected(network.Network, network.Conn) {}

// OpenedStream is called when a stream opened
func (n *NoConnectionMonitor) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when a stream closed
func (n *NoConnectionMonitor) ClosedStream(network.Network, network.Stream) {}

// IsConnectedToTheNetwork returns true if the number of connected peer is at least equal with thresholdMinConnectedPeers
func (n *NoConnectionMonitor) IsConnectedToTheNetwork(netw network.Network) bool {
	return len(netw.Conns()) >= n.thresholdMinConnectedPeers
}

// SetThresholdMinConnectedPeers sets the minimum connected peers number when the node is considered connected on the network
func (n *NoConnectionMonitor) SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, _ network.Network) {
	n.thresholdMinConnectedPeers = thresholdMinConnectedPeers
}

// ThresholdMinConnectedPeers returns the minimum connected peers number when the node is considered connected on the network
func (n *NoConnectionMonitor) ThresholdMinConnectedPeers() int {
	return n.thresholdMinConnectedPeers
}

// SetSharder sets the sharder that is able to sort the peers by their distance. Should not be used inside this implementation
func (n *NoConnectionMonitor) SetSharder(_ p2p.CommonSharder) error {
	return fmt.Errorf("%w while calling NilConnectionMonitor.SetSharder", p2p.ErrIncompatibleMethodCalled)
}
