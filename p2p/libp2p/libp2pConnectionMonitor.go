package libp2p

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

const (
	watchdogTimeout = 5 * time.Minute
)

// DurationBetweenReconnectAttempts is used as to not call reconnecter.ReconnectToNetwork() too often
// when there are a lot of peers disconnecting and reconnection to initial nodes succeeds
var DurationBetweenReconnectAttempts = time.Second * 5

type libp2pConnectionMonitor struct {
	chDoReconnect              chan struct{}
	reconnecter                p2p.Reconnecter
	thresholdMinConnectedPeers int
	sharder                    Sharder
}

func newLibp2pConnectionMonitor(reconnecter p2p.Reconnecter, thresholdMinConnectedPeers int) (*libp2pConnectionMonitor, error) {
	if thresholdMinConnectedPeers < 0 {
		return nil, p2p.ErrInvalidValue
	}

	cm := &libp2pConnectionMonitor{
		reconnecter:                reconnecter,
		chDoReconnect:              make(chan struct{}, 0),
		thresholdMinConnectedPeers: thresholdMinConnectedPeers,
	}

	if reconnecter != nil {
		go cm.doReconnection()
		_ = reconnecter.StartWatchdog(watchdogTimeout)
	}

	return cm, nil
}

// Listen is called when network starts listening on an addr
func (lcm *libp2pConnectionMonitor) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (lcm *libp2pConnectionMonitor) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Request a reconnect to initial list
func (lcm *libp2pConnectionMonitor) doReconn() {
	select {
	case lcm.chDoReconnect <- struct{}{}:
	default:
	}
}

func (lcm *libp2pConnectionMonitor) kickWatchdog() {
	if lcm.reconnecter != nil {
		_ = lcm.reconnecter.KickWatchdog()
	}
}

// Connected is called when a connection opened
func (lcm *libp2pConnectionMonitor) Connected(netw network.Network, conn network.Conn) {
	if lcm.sharder != nil {
		evicted := lcm.sharder.ComputeEvictList(conn.RemotePeer(), netw.Peers())
		for _, pid := range evicted {
			_ = netw.ClosePeer(pid)
		}
	}
}

// Disconnected is called when a connection closed
func (lcm *libp2pConnectionMonitor) Disconnected(netw network.Network, _ network.Conn) {
	lcm.doReconnectionIfNeeded(netw)
}

func (lcm *libp2pConnectionMonitor) doReconnectionIfNeeded(netw network.Network) {
	if !lcm.isConnectedToTheNetwork(netw) {
		lcm.doReconn()
	}
}

// OpenedStream is called when a stream opened
func (lcm *libp2pConnectionMonitor) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when a stream closed
func (lcm *libp2pConnectionMonitor) ClosedStream(network.Network, network.Stream) {}

func (lcm *libp2pConnectionMonitor) doReconnection() {
	for {
		select {
		case <-lcm.chDoReconnect:
			<-lcm.reconnecter.ReconnectToNetwork()
		}

		time.Sleep(DurationBetweenReconnectAttempts)
	}
}

func (lcm *libp2pConnectionMonitor) isConnectedToTheNetwork(netw network.Network) bool {
	return len(netw.Conns()) >= lcm.thresholdMinConnectedPeers
}

// SetSharder sets the sharder that is able to sort the peers by their distance
func (lcm *libp2pConnectionMonitor) SetSharder(sharder Sharder) error {
	if check.IfNil(sharder) {
		return p2p.ErrNilSharder
	}

	lcm.sharder = sharder

	return nil
}
