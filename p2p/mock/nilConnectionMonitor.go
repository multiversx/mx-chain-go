package mock

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// NilConnectionMonitor -
type NilConnectionMonitor struct {
	threshold int
}

// Listen -
func (n *NilConnectionMonitor) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose -
func (n *NilConnectionMonitor) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected -
func (n *NilConnectionMonitor) Connected(network.Network, network.Conn) {}

// Disconnected -
func (n *NilConnectionMonitor) Disconnected(network.Network, network.Conn) {}

// OpenedStream -
func (n *NilConnectionMonitor) OpenedStream(network.Network, network.Stream) {}

// ClosedStream -
func (n *NilConnectionMonitor) ClosedStream(network.Network, network.Stream) {}

// IsConnectedToTheNetwork -
func (n *NilConnectionMonitor) IsConnectedToTheNetwork(_ network.Network) bool {
	return true
}

// SetThresholdMinConnectedPeers -
func (n *NilConnectionMonitor) SetThresholdMinConnectedPeers(threshold int, _ network.Network) {
	n.threshold = threshold
}

// ThresholdMinConnectedPeers -
func (n *NilConnectionMonitor) ThresholdMinConnectedPeers() int {
	return n.threshold
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *NilConnectionMonitor) IsInterfaceNil() bool {
	return n == nil
}
