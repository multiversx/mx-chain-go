package connectionMonitor

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// NilConnectionMonitor does not monitor the connections made by the host
type NilConnectionMonitor struct {
	thresholdMinConnectedPeers int
}

// Listen is called when network starts listening on an addr
func (n *NilConnectionMonitor) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (n *NilConnectionMonitor) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened
func (n *NilConnectionMonitor) Connected(network.Network, network.Conn) {}

// Disconnected is called when a connection closed
func (n *NilConnectionMonitor) Disconnected(network.Network, network.Conn) {}

// OpenedStream is called when a stream opened
func (n *NilConnectionMonitor) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when a stream closed
func (n *NilConnectionMonitor) ClosedStream(network.Network, network.Stream) {}

// IsConnectedToTheNetwork returns true if the number of connected peer is at least equal with thresholdMinConnectedPeers
func (n *NilConnectionMonitor) IsConnectedToTheNetwork(netw network.Network) bool {
	return len(netw.Conns()) >= n.thresholdMinConnectedPeers
}

// SetThresholdMinConnectedPeers sets the minimum connected peers number when the node is considered connected on the network
func (n *NilConnectionMonitor) SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, _ network.Network) {
	n.thresholdMinConnectedPeers = thresholdMinConnectedPeers
}

// ThresholdMinConnectedPeers returns the minimum connected peers number when the node is considered connected on the network
func (n *NilConnectionMonitor) ThresholdMinConnectedPeers() int {
	return n.thresholdMinConnectedPeers
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *NilConnectionMonitor) IsInterfaceNil() bool {
	return n == nil
}
