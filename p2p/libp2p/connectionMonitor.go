package libp2p

import (
	"fmt"
	"math"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

type connectionMonitor struct {
	chDoReconnect              chan struct{}
	reconnecter                p2p.Reconnecter
	libp2pContext              *Libp2pContext
	netw                       network.Network
	thresholdMinConnectedPeers int
	thresholdDiscoveryResume   int // if the number of connections drops under this value, the discovery is restarted
	thresholdDiscoveryPause    int // if the number of connections is over this value, the discovery is stopped
	thresholdConnTrim          int // if the number of connections is over this value, we start trimming
}

func newConnectionMonitor(
	reconnecter p2p.Reconnecter,
	libp2pContext *Libp2pContext,
	thresholdMinConnectedPeers int,
	targetConnCount int,
) (*connectionMonitor, error) {

	if thresholdMinConnectedPeers <= 0 {
		return nil, fmt.Errorf("%w, thresholdMinConnectedPeers expected to be higher than 0",
			p2p.ErrInvalidValue)
	}
	if targetConnCount <= 0 {
		return nil, fmt.Errorf("%w, targetConnCount expected to be higher than 0", p2p.ErrInvalidValue)
	}
	if check.IfNil(libp2pContext) {
		return nil, p2p.ErrNilContextProvider
	}

	cm := &connectionMonitor{
		reconnecter:                reconnecter,
		chDoReconnect:              make(chan struct{}, 0),
		libp2pContext:              libp2pContext,
		netw:                       libp2pContext.connHost.Network(),
		thresholdMinConnectedPeers: thresholdMinConnectedPeers,
		thresholdDiscoveryResume:   0,
		thresholdDiscoveryPause:    math.MaxInt32,
		thresholdConnTrim:          math.MaxInt32,
	}

	if targetConnCount > 0 {
		cm.thresholdDiscoveryResume = targetConnCount * 4 / 5
		cm.thresholdDiscoveryPause = targetConnCount
		cm.thresholdConnTrim = targetConnCount * 6 / 5
	}

	return cm, nil
}

// HandleConnectedPeer is called whenever a new peer is connected to the current host
func (cm *connectionMonitor) HandleConnectedPeer(pid p2p.PeerID) error {
	blacklistHandler := cm.libp2pContext.PeerBlacklist()
	if blacklistHandler.Has(string(pid)) {
		return fmt.Errorf("%w, pid: %s", p2p.ErrPeerBlacklisted, pid.Pretty())
	}

	if len(cm.netw.Conns()) > cm.thresholdDiscoveryPause && cm.reconnecter != nil {
		cm.reconnecter.Pause()
	}
	if len(cm.netw.Conns()) > cm.thresholdConnTrim {
		sorted := networksharding.Get().SortList(cm.netw.Peers(), cm.netw.LocalPeer())
		for i := cm.thresholdDiscoveryPause; i < len(sorted); i++ {
			cm.closePeer(sorted[i])
		}
		cm.doReconn()
	}

	return nil
}

// HandleDisconnectedPeer is called whenever a new peer is disconnected from the current host
func (cm *connectionMonitor) HandleDisconnectedPeer(_ p2p.PeerID) error {
	cm.doReconnectionIfNeeded()

	if len(cm.netw.Conns()) < cm.thresholdDiscoveryResume && cm.reconnecter != nil {
		cm.reconnecter.Resume()
		cm.doReconn()
	}

	return nil
}

// DoReconnectionBlocking will try to reconnect to the initial addresses (seeders)
func (cm *connectionMonitor) DoReconnectionBlocking() {
	select {
	case <-cm.chDoReconnect:
		if cm.reconnecter != nil {
			cm.reconnecter.ReconnectToNetwork()
		}
	}
}

// CheckConnectionsBlocking will sweep all available connections checking if the connection has or not been blacklisted
func (cm *connectionMonitor) CheckConnectionsBlocking() {
	peers := cm.netw.Peers()
	blacklistHandler := cm.libp2pContext.PeerBlacklist()
	for _, pid := range peers {
		if blacklistHandler.Has(string(pid)) {
			cm.closePeer(pid)
		}
	}
}

func (cm *connectionMonitor) closePeer(pid peer.ID) {
	err := cm.netw.ClosePeer(pid)
	if err != nil {
		log.Trace("error closing peer connection HandleConnectedPeer",
			"pid", pid.Pretty(),
			"error", err.Error(),
		)
	}
}

// Request a reconnect to initial list
func (cm *connectionMonitor) doReconn() {
	select {
	case cm.chDoReconnect <- struct{}{}:
	default:
	}
}

func (cm *connectionMonitor) doReconnectionIfNeeded() {
	if !cm.isConnectedToTheNetwork() {
		cm.doReconn()
	}
}

func (cm *connectionMonitor) isConnectedToTheNetwork() bool {
	return len(cm.netw.Conns()) >= cm.thresholdMinConnectedPeers
}

// IsInterfaceNil returns true if there is no value under the interface
func (cm *connectionMonitor) IsInterfaceNil() bool {
	return cm == nil
}
