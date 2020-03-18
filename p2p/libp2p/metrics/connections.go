package metrics

import (
	"sync/atomic"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// Connections is a metric that counts connections and disconnections done by the host implementation
type Connections struct {
	numConnections    uint32
	numDisconnections uint32
}

// NewConnections returns a new connsDisconnsMetric instance
func NewConnections() *Connections {
	return &Connections{
		numConnections:    0,
		numDisconnections: 0,
	}
}

// Listen is called when network starts listening on an addr
func (conns *Connections) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (conns *Connections) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened. It increments the numConnections counter
func (conns *Connections) Connected(network.Network, network.Conn) {
	atomic.AddUint32(&conns.numConnections, 1)
}

// Disconnected is called when a connection closed it increments the numDisconnections counter
func (conns *Connections) Disconnected(network.Network, network.Conn) {
	atomic.AddUint32(&conns.numDisconnections, 1)
}

// OpenedStream is called when a stream opened
func (conns *Connections) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when a stream closed
func (conns *Connections) ClosedStream(network.Network, network.Stream) {}

// ResetNumConnections resets the numConnections counter returning the previous value
func (conns *Connections) ResetNumConnections() uint32 {
	return atomic.SwapUint32(&conns.numConnections, 0)
}

// ResetNumDisconnections resets the numDisconnections counter returning the previous value
func (conns *Connections) ResetNumDisconnections() uint32 {
	return atomic.SwapUint32(&conns.numDisconnections, 0)
}
