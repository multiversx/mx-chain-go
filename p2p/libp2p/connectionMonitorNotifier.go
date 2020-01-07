package libp2p

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// connectionMonitorNotifier is a wrapper over p2p.ConnectionMonitor that satisfies the Notifiee interface
// and is able to be notified by the current running host (connection status changes)
type connectionMonitorNotifier struct {
	p2p.ConnectionMonitor
}

// Listen is called when network starts listening on an addr
func (cmn *connectionMonitorNotifier) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (cmn *connectionMonitorNotifier) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened
func (cmn *connectionMonitorNotifier) Connected(netw network.Network, conn network.Conn) {
	err := cmn.HandleConnectedPeer(p2p.PeerID(conn.RemotePeer()))
	if err != nil {
		log.Trace("connection error",
			"pid", conn.RemotePeer().Pretty(),
			"error", err.Error(),
		)

		err = netw.ClosePeer(conn.RemotePeer())
		if err != nil {
			log.Trace("connection close error",
				"pid", conn.RemotePeer().Pretty(),
				"error", err.Error(),
			)
		}
	}
}

// Disconnected is called when a connection closed
func (cmn *connectionMonitorNotifier) Disconnected(_ network.Network, conn network.Conn) {
	cmn.HandleDisconnectedPeer(p2p.PeerID(conn.RemotePeer()))
}

// OpenedStream is called when a stream opened
func (cmn *connectionMonitorNotifier) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when a stream closed
func (cmn *connectionMonitorNotifier) ClosedStream(network.Network, network.Stream) {}
