package libp2p

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

// ThresholdMinimumConnectedPeers if the number of connected peers drop under this value, for each disconnecting
// peer, a trigger to reconnect to initial peers is done
var ThresholdMinimumConnectedPeers = 3

// DurationBetweenReconnectAttempts is used as to not call reconnecter.ReconnectToNetwork() to often
// when there are a lot of peers disconnecting and reconnection to initial nodes succeed
var DurationBetweenReconnectAttempts = time.Duration(time.Second * 5)

type libp2pConnectionMonitor struct {
	chDoReconnect chan struct{}
	reconnecter   p2p.Reconnecter
}

func newLibp2pConnectionMonitor(reconnecter p2p.Reconnecter) *libp2pConnectionMonitor {
	cm := &libp2pConnectionMonitor{
		reconnecter:   reconnecter,
		chDoReconnect: make(chan struct{}, 0),
	}

	if reconnecter != nil {
		go cm.doReconnection()
	}

	return cm
}

// Listen is called when network starts listening on an addr
func (lcm *libp2pConnectionMonitor) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (lcm *libp2pConnectionMonitor) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened
func (lcm *libp2pConnectionMonitor) Connected(network.Network, network.Conn) {}

// Disconnected is called when a connection closed
func (lcm *libp2pConnectionMonitor) Disconnected(netw network.Network, conn network.Conn) {
	if len(netw.Conns()) < ThresholdMinimumConnectedPeers {
		select {
		case lcm.chDoReconnect <- struct{}{}:
		default:
		}
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
