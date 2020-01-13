package connectionMonitor

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

type libp2pConnectionMonitor2 struct {
	chDoReconnect              chan struct{}
	reconnecter                p2p.Reconnecter
	thresholdMinConnectedPeers int
	sharder                    Sharder
}

// NewLibp2pConnectionMonitor2 creates a new connection monitor (version 2 that is more streamlined and does not care
//about pausing and resuming the discovery process)
func NewLibp2pConnectionMonitor2(reconnecter p2p.Reconnecter, thresholdMinConnectedPeers int) (*libp2pConnectionMonitor2, error) {
	if thresholdMinConnectedPeers < 0 {
		return nil, p2p.ErrInvalidValue
	}

	cm := &libp2pConnectionMonitor2{
		reconnecter:                reconnecter,
		chDoReconnect:              make(chan struct{}),
		thresholdMinConnectedPeers: thresholdMinConnectedPeers,
	}

	if reconnecter != nil {
		go cm.doReconnection()
	}

	return cm, nil
}

// Listen is called when network starts listening on an addr
func (lcm *libp2pConnectionMonitor2) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (lcm *libp2pConnectionMonitor2) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Request a reconnect to initial list
func (lcm *libp2pConnectionMonitor2) doReconn() {
	select {
	case lcm.chDoReconnect <- struct{}{}:
	default:
	}
}

// Connected is called when a connection opened
func (lcm *libp2pConnectionMonitor2) Connected(netw network.Network, conn network.Conn) {
	if lcm.sharder != nil {
		evicted := lcm.sharder.ComputeEvictList(conn.RemotePeer(), netw.Peers())
		for _, pid := range evicted {
			_ = netw.ClosePeer(pid)
		}
	}
}

// Disconnected is called when a connection closed
func (lcm *libp2pConnectionMonitor2) Disconnected(netw network.Network, _ network.Conn) {
	lcm.doReconnectionIfNeeded(netw)
}

func (lcm *libp2pConnectionMonitor2) doReconnectionIfNeeded(netw network.Network) {
	if !lcm.IsConnectedToTheNetwork(netw) {
		lcm.doReconn()
	}
}

// OpenedStream is called when a stream opened
func (lcm *libp2pConnectionMonitor2) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when a stream closed
func (lcm *libp2pConnectionMonitor2) ClosedStream(network.Network, network.Stream) {}

func (lcm *libp2pConnectionMonitor2) doReconnection() {
	for {
		<-lcm.chDoReconnect
		<-lcm.reconnecter.ReconnectToNetwork()

		time.Sleep(DurationBetweenReconnectAttempts)
	}
}

// IsConnectedToTheNetwork returns true if the number of connected peer is at least equal with thresholdMinConnectedPeers
func (lcm *libp2pConnectionMonitor2) IsConnectedToTheNetwork(netw network.Network) bool {
	return len(netw.Peers()) >= lcm.thresholdMinConnectedPeers
}

// SetThresholdMinConnectedPeers sets the minimum connected peers number when the node is considered connected on the network
//TODO(iulian) refactor this in a future PR (not require to inject the netw pointer)
func (lcm *libp2pConnectionMonitor2) SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, netw network.Network) {
	if netw == nil {
		return
	}
	lcm.thresholdMinConnectedPeers = thresholdMinConnectedPeers
	lcm.doReconnectionIfNeeded(netw)
}

// ThresholdMinConnectedPeers returns the minimum connected peers number when the node is considered connected on the network
func (lcm *libp2pConnectionMonitor2) ThresholdMinConnectedPeers() int {
	return lcm.thresholdMinConnectedPeers
}

// SetSharder sets the sharder that is able to sort the peers by their distance
// TODO(iulian) change this from interface{} to Sharder interface when all implementations will be uniformized
func (lcm *libp2pConnectionMonitor2) SetSharder(sharder interface{}) error {
	sharderIntf, ok := sharder.(Sharder)
	if !ok {
		return fmt.Errorf("%w when applying sharder: expected interface libp2p.Sharder", p2p.ErrWrongTypeAssertion)
	}
	if check.IfNil(sharderIntf) {
		return p2p.ErrNilSharder
	}

	lcm.sharder = sharderIntf

	return nil
}
