package p2p

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

// durRefreshConnections represents the duration used to pause between refreshing connections to known peers
const durRefreshConnections = 1000 * time.Millisecond

// ResultType will signal the result
type ResultType int

const (
	// WontConnect will not try to connect to other peers
	WontConnect ResultType = iota
	// OnlyInboundConnections means that there are only inbound connections
	OnlyInboundConnections
	// SuccessfullyConnected signals that has successfully connected to a peer
	SuccessfullyConnected
	// NothingDone nothing has been done
	NothingDone
)

// ConnNotifier is used to manage the connections to other peers
type ConnNotifier struct {
	execution.RoutineWrapper

	Msgr Messenger

	MaxAllowedPeers int

	GetKnownPeers func(sender *ConnNotifier) []peer.ID
	ConnectToPeer func(sender *ConnNotifier, pid peer.ID) error

	indexKnownPeers int
}

// NewConnNotifier will create a new object and link it to the messenger provided as parameter
func NewConnNotifier(m Messenger) *ConnNotifier {
	if m == nil {
		panic("Nil messenger!")
	}

	cn := ConnNotifier{Msgr: m}
	cn.RoutineWrapper = *execution.NewRoutineWrapper()
	//there is a delay between calls so the connecting and disconnecting is not done
	//very often (take into account that doing a connection is a lengthy process)
	cn.RoutineWrapper.DurCalls = durRefreshConnections

	return &cn
}

// TaskResolveConnections resolves the connections to other peers. It should not be called too often as the
// connections are not done instantly. Even if the connection is made in a short time, there is a delay
// until the connected peer might close down the connections because it reached the maximum limit.
// This function handles the array connections that mdns service provides.
// This function always tries to find a new connection by closing the oldest one.
// It tries to create a new outbound connection by iterating over known peers for at least one cycle but not 2 or more.
func TaskResolveConnections(cn *ConnNotifier) ResultType {
	if cn.MaxAllowedPeers < 1 {
		//won't try to connect to other peers
		return WontConnect
	}

	conns := cn.Msgr.Conns()

	knownPeers := make([]peer.ID, 0)

	if cn.GetKnownPeers != nil {
		knownPeers = cn.GetKnownPeers(cn)
	}

	inConns := 0
	outConns := 0

	//get how many inbound and outbound connection we have
	for i := 0; i < len(conns); i++ {
		if conns[i].Stat().Direction == net.DirInbound {
			inConns++
		}

		if conns[i].Stat().Direction == net.DirOutbound {
			outConns++
		}
	}

	//test whether we only have inbound connection (security issue)
	if inConns > cn.MaxAllowedPeers-1 {
		conns[0].Close()

		return OnlyInboundConnections
	}

	fullCycles := 0

	//try to connect to other peers
	if len(conns) < cn.MaxAllowedPeers && len(knownPeers) > 0 {
		//the value of 2 was chosen as a mean to iterate the known peers at least 1 time but not 2 times
		for fullCycles < 2 {
			if cn.indexKnownPeers >= len(knownPeers) {
				//index out of bound, do 0 (restart the list)
				cn.indexKnownPeers = 0
				fullCycles++
			}

			//get the known peerID
			peerID := knownPeers[cn.indexKnownPeers]
			cn.indexKnownPeers++

			if cn.Msgr.Connectedness(peerID) == net.NotConnected {
				if cn.ConnectToPeer != nil {
					err := cn.ConnectToPeer(cn, peerID)
					if err == nil {
						//connection succeed
						return SuccessfullyConnected
					}
				}
			}
		}
	}

	return NothingDone
}

// Listen is called when network starts listening on an addr
func (cn *ConnNotifier) Listen(netw net.Network, ma multiaddr.Multiaddr) {
	//Nothing to be done
}

// ListenClose is called when network starts listening on an addr
func (cn *ConnNotifier) ListenClose(netw net.Network, ma multiaddr.Multiaddr) {
	//Nothing to be done
}

// Connected is called when a connection opened
func (cn *ConnNotifier) Connected(netw net.Network, conn net.Conn) {
	//refuse other connections if max connection has been reached
	if cn.MaxAllowedPeers < len(cn.Msgr.Conns()) {
		conn.Close()
	}
}

// Disconnected is called when a connection closed
func (cn *ConnNotifier) Disconnected(netw net.Network, conn net.Conn) {
	//Nothing to be done
}

// OpenedStream is called when a stream opened
func (cn *ConnNotifier) OpenedStream(netw net.Network, stream net.Stream) {
	//Nothing to be done
}

// ClosedStream is called when a stream was closed
func (cn *ConnNotifier) ClosedStream(netw net.Network, stream net.Stream) {
	//Nothing to be done
}
