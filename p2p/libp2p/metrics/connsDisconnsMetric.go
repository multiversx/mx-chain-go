package metrics

import (
	"sync/atomic"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// ConnsDisconnsMetric is a metric that counts connections and disconnections done by the host implementation
type ConnsDisconnsMetric struct {
	numConnections    uint32
	numDisconnections uint32
}

// NewConnsDisconnsMetric returns a new connsDisconnsMetric instance
func NewConnsDisconnsMetric() *ConnsDisconnsMetric {
	return &ConnsDisconnsMetric{
		numConnections:    0,
		numDisconnections: 0,
	}
}

// Listen is called when network starts listening on an addr
func (cdm *ConnsDisconnsMetric) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (cdm *ConnsDisconnsMetric) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened. It increments the numConnections counter
func (cdm *ConnsDisconnsMetric) Connected(network.Network, network.Conn) {
	atomic.AddUint32(&cdm.numConnections, 1)
}

// Disconnected is called when a connection closed it increments the numDisconnections counter
func (cdm *ConnsDisconnsMetric) Disconnected(network.Network, network.Conn) {
	atomic.AddUint32(&cdm.numDisconnections, 1)
}

// OpenedStream is called when a stream opened
func (cdm *ConnsDisconnsMetric) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when a stream closed
func (cdm *ConnsDisconnsMetric) ClosedStream(network.Network, network.Stream) {}

// ResetNumConnections resets the numConnections counter returning the previous value
func (cdm *ConnsDisconnsMetric) ResetNumConnections() uint32 {
	return atomic.SwapUint32(&cdm.numConnections, 0)
}

// ResetNumDisconnections resets the numDisconnections counter returning the previous value
func (cdm *ConnsDisconnsMetric) ResetNumDisconnections() uint32 {
	return atomic.SwapUint32(&cdm.numDisconnections, 0)
}
