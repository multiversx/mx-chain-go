package libp2p

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

var _ ConnectionMonitor = (*connectionMonitorWrapper)(nil)

// connectionMonitorWrapper is a wrapper over p2p.ConnectionMonitor that satisfies the Notifiee interface
// and is able to be notified by the current running host (connection status changes)
// it handles black list peers
type connectionMonitorWrapper struct {
	ConnectionMonitor
	network          network.Network
	mutPeerBlackList sync.RWMutex
	peerBlackList    p2p.BlacklistHandler
}

func newConnectionMonitorWrapper(
	network network.Network,
	connMonitor ConnectionMonitor,
	blackList p2p.BlacklistHandler,
) *connectionMonitorWrapper {
	return &connectionMonitorWrapper{
		ConnectionMonitor: connMonitor,
		network:           network,
		peerBlackList:     blackList,
	}
}

// Listen is called when network starts listening on an addr
func (cmw *connectionMonitorWrapper) Listen(netw network.Network, ma multiaddr.Multiaddr) {
	cmw.ConnectionMonitor.Listen(netw, ma)
}

// ListenClose is called when network stops listening on an addr
func (cmw *connectionMonitorWrapper) ListenClose(netw network.Network, ma multiaddr.Multiaddr) {
	cmw.ConnectionMonitor.ListenClose(netw, ma)
}

// Connected is called when a connection opened
func (cmw *connectionMonitorWrapper) Connected(netw network.Network, conn network.Conn) {
	cmw.mutPeerBlackList.RLock()
	peerBlackList := cmw.peerBlackList
	cmw.mutPeerBlackList.RUnlock()

	pid := conn.RemotePeer()
	if peerBlackList.Has(pid.Pretty()) {
		log.Debug("dropping connection to black listed peer",
			"pid", pid.Pretty(),
		)
		_ = conn.Close()

		return
	}

	cmw.ConnectionMonitor.Connected(netw, conn)
}

// Disconnected is called when a connection closed
func (cmw *connectionMonitorWrapper) Disconnected(netw network.Network, conn network.Conn) {
	cmw.ConnectionMonitor.Disconnected(netw, conn)
}

// OpenedStream is called when a stream opened
func (cmw *connectionMonitorWrapper) OpenedStream(netw network.Network, stream network.Stream) {
	cmw.ConnectionMonitor.OpenedStream(netw, stream)
}

// ClosedStream is called when a stream closed
func (cmw *connectionMonitorWrapper) ClosedStream(netw network.Network, stream network.Stream) {
	cmw.ConnectionMonitor.ClosedStream(netw, stream)
}

// CheckConnectionsBlocking does a peer sweep, calling Close on those peers that are black listed
func (cmw *connectionMonitorWrapper) CheckConnectionsBlocking() {
	peers := cmw.network.Peers()
	cmw.mutPeerBlackList.RLock()
	blacklistHandler := cmw.peerBlackList
	cmw.mutPeerBlackList.RUnlock()

	for _, pid := range peers {
		if blacklistHandler.Has(pid.Pretty()) {
			log.Debug("dropping connection to black listed peer",
				"pid", pid.Pretty(),
			)
			_ = cmw.network.ClosePeer(pid)
		}
	}
}

// SetBlackListHandler sets the black list handler
func (cmw *connectionMonitorWrapper) SetBlackListHandler(handler p2p.BlacklistHandler) error {
	if check.IfNil(handler) {
		return p2p.ErrNilPeerBlacklistHandler
	}

	cmw.mutPeerBlackList.Lock()
	cmw.peerBlackList = handler
	cmw.mutPeerBlackList.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (cmw *connectionMonitorWrapper) IsInterfaceNil() bool {
	return cmw == nil
}
