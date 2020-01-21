package connectionMonitor

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	ns "github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

const (
	watchdogTimeout = 5 * time.Minute
)

// DurationBetweenReconnectAttempts is used as to not call reconnecter.ReconnectToNetwork() too often
// when there are a lot of peers disconnecting and reconnection to initial nodes succeeds
var DurationBetweenReconnectAttempts = time.Second * 5

// Libp2pConnectionMonitor is the first variant of the connection monitor implementation
type libp2pConnectionMonitor struct {
	chDoReconnect              chan struct{}
	reconnecter                p2p.ReconnecterWithPauseResumeAndWatchdog
	thresholdMinConnectedPeers int
	thresholdDiscoveryResume   int // if the number of connections drops under this value, the discovery is restarted
	thresholdDiscoveryPause    int // if the number of connections is over this value, the discovery is stopped
	thresholdConnTrim          int // if the number of connections is over this value, we start trimming
	mutSharder                 sync.RWMutex
	sharder                    ns.Sharder
}

// NewLibp2pConnectionMonitor creates a new connection monitor (version 1 that calls pause and resume on the discovery process)
func NewLibp2pConnectionMonitor(
	reconnecter p2p.ReconnecterWithPauseResumeAndWatchdog,
	thresholdMinConnectedPeers int,
	targetConnCount int,
) (*libp2pConnectionMonitor, error) {

	if thresholdMinConnectedPeers < 0 {
		return nil, p2p.ErrInvalidValue
	}

	cm := &libp2pConnectionMonitor{
		reconnecter:                reconnecter,
		chDoReconnect:              make(chan struct{}, 0),
		thresholdMinConnectedPeers: thresholdMinConnectedPeers,
		thresholdDiscoveryResume:   0,
		thresholdDiscoveryPause:    math.MaxInt32,
		thresholdConnTrim:          math.MaxInt32,
		sharder:                    &ns.NoSharder{},
	}

	if targetConnCount > 0 {
		cm.thresholdDiscoveryResume = targetConnCount * 4 / 5
		cm.thresholdDiscoveryPause = targetConnCount
		cm.thresholdConnTrim = targetConnCount * 6 / 5
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
func (lcm *libp2pConnectionMonitor) Connected(netw network.Network, _ network.Conn) {
	if len(netw.Conns()) > lcm.thresholdConnTrim {
		lcm.mutSharder.RLock()
		sorted, isBalanced := lcm.sharder.SortList(netw.Peers(), netw.LocalPeer())
		lcm.mutSharder.RUnlock()
		for i := lcm.thresholdDiscoveryPause; i < len(sorted); i++ {
			_ = netw.ClosePeer(sorted[i])
		}
		lcm.doReconn()
		if isBalanced && lcm.reconnecter != nil {
			lcm.reconnecter.Pause()
		}
	}
	lcm.kickWatchdog()
}

// Disconnected is called when a connection closed
func (lcm *libp2pConnectionMonitor) Disconnected(netw network.Network, _ network.Conn) {
	lcm.doReconnectionIfNeeded(netw)

	if len(netw.Conns()) < lcm.thresholdDiscoveryResume && lcm.reconnecter != nil {
		lcm.reconnecter.Resume()
		lcm.doReconn()
	}
	lcm.kickWatchdog()
}

func (lcm *libp2pConnectionMonitor) doReconnectionIfNeeded(netw network.Network) {
	if !lcm.IsConnectedToTheNetwork(netw) {
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

// IsConnectedToTheNetwork returns true if the number of connected peer is at least equal with thresholdMinConnectedPeers
func (lcm *libp2pConnectionMonitor) IsConnectedToTheNetwork(netw network.Network) bool {
	return len(netw.Conns()) >= lcm.thresholdMinConnectedPeers
}

// SetThresholdMinConnectedPeers sets the minimum connected peers number when the node is considered connected on the network
//TODO(iulian) refactor this in a future PR (not require to inject the netw pointer)
func (lcm *libp2pConnectionMonitor) SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, netw network.Network) {
	if netw == nil {
		return
	}
	lcm.thresholdMinConnectedPeers = thresholdMinConnectedPeers
	lcm.doReconnectionIfNeeded(netw)
}

// ThresholdMinConnectedPeers returns the minimum connected peers number when the node is considered connected on the network
func (lcm *libp2pConnectionMonitor) ThresholdMinConnectedPeers() int {
	return lcm.thresholdMinConnectedPeers
}

// SetSharder sets the sharder that is able to sort the peers by their distance
//TODO(iulian) change this from interface{} to Sharder interface when all implementations will be uniformized
// (maybe merge with libp2pConnectionMonitor2)
func (lcm *libp2pConnectionMonitor) SetSharder(sharder interface{}) error {
	sharderIntf, ok := sharder.(ns.Sharder)
	if !ok {
		return fmt.Errorf("%w when applying sharder: expected interface networksharding.Sharder", p2p.ErrWrongTypeAssertion)
	}
	if check.IfNil(sharderIntf) {
		return p2p.ErrNilSharder
	}

	lcm.mutSharder.Lock()
	lcm.sharder = sharderIntf
	lcm.mutSharder.Unlock()

	return nil
}
