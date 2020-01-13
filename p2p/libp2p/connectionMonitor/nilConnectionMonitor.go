package connectionMonitor

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

type NilConnectionMonitor struct {
	thresholdMinConnectedPeers int
}

func (n *NilConnectionMonitor) Listen(network.Network, multiaddr.Multiaddr) {}

func (n *NilConnectionMonitor) ListenClose(network.Network, multiaddr.Multiaddr) {}

func (n *NilConnectionMonitor) Connected(network.Network, network.Conn) {}

func (n *NilConnectionMonitor) Disconnected(network.Network, network.Conn) {}

func (n *NilConnectionMonitor) OpenedStream(network.Network, network.Stream) {}

func (n *NilConnectionMonitor) ClosedStream(network.Network, network.Stream) {}

func (n *NilConnectionMonitor) IsConnectedToTheNetwork(netw network.Network) bool {
	return len(netw.Conns()) >= n.thresholdMinConnectedPeers
}

func (n *NilConnectionMonitor) SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, _ network.Network) {
	n.thresholdMinConnectedPeers = thresholdMinConnectedPeers
}

func (n *NilConnectionMonitor) ThresholdMinConnectedPeers() int {
	return n.thresholdMinConnectedPeers
}

func (n *NilConnectionMonitor) SetSharder(_ interface{}) error {
	return fmt.Errorf("%w while calling NilConnectionMonitor.SetSharder", p2p.ErrIncompatibleMethodCalled)
}
