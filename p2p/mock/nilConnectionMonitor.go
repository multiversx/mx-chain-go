package mock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

type NilConnectionMonitor struct {
	threshold int
}

func (n *NilConnectionMonitor) Listen(network.Network, multiaddr.Multiaddr) {}

func (n *NilConnectionMonitor) ListenClose(network.Network, multiaddr.Multiaddr) {}

func (n *NilConnectionMonitor) Connected(network.Network, network.Conn) {}

func (n *NilConnectionMonitor) Disconnected(network.Network, network.Conn) {}

func (n *NilConnectionMonitor) OpenedStream(network.Network, network.Stream) {}

func (n *NilConnectionMonitor) ClosedStream(network.Network, network.Stream) {}

func (n *NilConnectionMonitor) IsConnectedToTheNetwork(_ network.Network) bool {
	return true
}

func (n *NilConnectionMonitor) SetThresholdMinConnectedPeers(threshold int, _ network.Network) {
	n.threshold = threshold
}

func (n *NilConnectionMonitor) ThresholdMinConnectedPeers() int {
	return n.threshold
}

func (n *NilConnectionMonitor) SetSharder(_ interface{}) error {
	return fmt.Errorf("%w while calling NilConnectionMonitor.SetSharder", p2p.ErrIncompatibleMethodCalled)
}
