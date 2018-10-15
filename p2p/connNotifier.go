package p2p

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"

	"time"
)

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

	OnGetKnownPeers func(sender *ConnNotifier) []peer.ID
	OnNeedToConn    func(sender *ConnNotifier, pid peer.ID) error

	indexKnownPeers int
}

// NewConnNotifier will create a new object and link-it to the messenger provided as parameter
func NewConnNotifier(m Messenger) *ConnNotifier {
	if m == nil {
		panic("Nil messenger!")
	}

	cn := ConnNotifier{Msgr: m}
	cn.RoutineWrapper = *execution.NewRoutineWrapper()
	//there is a 100 ms delay between calls so the connecting and disconnecting is not done
	//very often (take into account that doing a connection is a lengthy process)
	cn.RoutineWrapper.DurCalls = 100 * time.Millisecond

	return &cn
}

// TaskMonitorConnections monitors the connections. It should not be called to often as the
// connections are not done instantly. Even if the connection is made in a short time, there is a delay
// until the connected peer might close down the connections because it reached the maximum limit.
func TaskMonitorConnections(cn *ConnNotifier) ResultType {
	if cn.MaxAllowedPeers < 1 {
		//won't try to connect to other peers
		return WontConnect
	}

	conns := cn.Msgr.Conns()

	knownPeers := make([]peer.ID, 0)

	if cn.OnGetKnownPeers != nil {
		knownPeers = cn.OnGetKnownPeers(cn)
	}

	inConns := 0
	outConns := 0

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
				if cn.OnNeedToConn != nil {
					err := cn.OnNeedToConn(cn, peerID)
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
func (cn *ConnNotifier) Listen(netw net.Network, ma multiaddr.Multiaddr) {}

// ListenClose is called when network starts listening on an addr
func (cn *ConnNotifier) ListenClose(netw net.Network, ma multiaddr.Multiaddr) {}

// Connected is called when a connection opened
func (cn *ConnNotifier) Connected(netw net.Network, conn net.Conn) {
	//refuse other connections
	if cn.MaxAllowedPeers < len(cn.Msgr.Conns()) {
		conn.Close()
	}
}

// Disconnected is called when a connection closed
func (cn *ConnNotifier) Disconnected(netw net.Network, conn net.Conn) {}

// OpenedStream is called when a stream opened
func (cn *ConnNotifier) OpenedStream(netw net.Network, stream net.Stream) {}

// ClosedStream is called when a stream was closed
func (cn *ConnNotifier) ClosedStream(netw net.Network, stream net.Stream) {}
