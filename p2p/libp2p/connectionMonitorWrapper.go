package libp2p

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
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
	network             network.Network
	mutPeerBlackList    sync.RWMutex
	peerDenialEvaluator p2p.PeerDenialEvaluator
}

func newConnectionMonitorWrapper(
	network network.Network,
	connMonitor ConnectionMonitor,
	peerDenialEvaluator p2p.PeerDenialEvaluator,
) *connectionMonitorWrapper {
	return &connectionMonitorWrapper{
		ConnectionMonitor:   connMonitor,
		network:             network,
		peerDenialEvaluator: peerDenialEvaluator,
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
	peerBlackList := cmw.peerDenialEvaluator
	cmw.mutPeerBlackList.RUnlock()

	pid := conn.RemotePeer()
	if peerBlackList.IsDenied(core.PeerID(pid)) {
		log.Trace("dropping connection to blacklisted peer",
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
	peerDenialEvaluator := cmw.peerDenialEvaluator
	cmw.mutPeerBlackList.RUnlock()

	for _, pid := range peers {
		if peerDenialEvaluator.IsDenied(core.PeerID(pid)) {
			log.Trace("dropping connection to blacklisted peer",
				"pid", pid.Pretty(),
			)
			_ = cmw.network.ClosePeer(pid)
		}
	}
}

// SetPeerDenialEvaluator sets the handler that is able to tell if a peer can connect to self or not (is or not blacklisted)
func (cmw *connectionMonitorWrapper) SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error {
	if check.IfNil(handler) {
		return p2p.ErrNilPeerDenialEvaluator
	}

	cmw.mutPeerBlackList.Lock()
	cmw.peerDenialEvaluator = handler
	cmw.mutPeerBlackList.Unlock()

	return nil
}

// PeerDenialEvaluator gets the peer denial evauator
func (cmw *connectionMonitorWrapper) PeerDenialEvaluator() p2p.PeerDenialEvaluator {
	cmw.mutPeerBlackList.RLock()
	defer cmw.mutPeerBlackList.RUnlock()

	return cmw.peerDenialEvaluator
}

// IsInterfaceNil returns true if there is no value under the interface
func (cmw *connectionMonitorWrapper) IsInterfaceNil() bool {
	return cmw == nil
}
