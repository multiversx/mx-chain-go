package libp2p

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// connectionMonitorNotifiee is a wrapper over p2p.ConnectionMonitor that satisfies the Notifiee interface
// and is able to be notified by the current running host (connection status changes)
type connectionMonitorNotifiee struct {
	p2p.ConnectionMonitor
}

// Listen is called when network starts listening on an addr
func (cmn *connectionMonitorNotifiee) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (cmn *connectionMonitorNotifiee) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened
func (cmn *connectionMonitorNotifiee) Connected(netw network.Network, conn network.Conn) {
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
func (cmn *connectionMonitorNotifiee) Disconnected(_ network.Network, conn network.Conn) {
	err := cmn.HandleDisconnectedPeer(p2p.PeerID(conn.RemotePeer()))
	if err != nil {
		log.Trace("disconnection error",
			"pid", conn.RemotePeer().Pretty(),
			"error", err.Error(),
		)
	}
}

// OpenedStream is called when a stream opened
func (cmn *connectionMonitorNotifiee) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when a stream closed
func (cmn *connectionMonitorNotifiee) ClosedStream(network.Network, network.Stream) {}
