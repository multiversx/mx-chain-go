package p2p

import (
	_ "fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"

	//"sync"
	"time"
)

type ConnNotifier struct {
	execution.RoutineWrapper

	Mes Messenger

	//stopConnKeeper chan bool

	//mutStat sync.RWMutex
	//stat    GoRoutineStat

	MaxPeersAllowed int

	//OnDoSimpleTask             func(sender *ConnNotifier)
	OnGetKnownPeers            func(sender *ConnNotifier) []peer.ID
	OnNeedToConnectToOtherPeer func(sender *ConnNotifier, pid peer.ID) error

	indexKnownPeers int
}

func NewConnNotifier(m Messenger) *ConnNotifier {
	if m == nil {
		panic("Nil messenger!")
	}

	cn := ConnNotifier{Mes: m}
	cn.RoutineWrapper = *execution.NewRoutineWrapper()

	return &cn
}

func TaskMonitorConnections(cn *ConnNotifier) (testOutput int) {
	if cn.MaxPeersAllowed < 1 {
		//won't try to connect to other peers
		return 1
	}

	defer func() {
		time.Sleep(time.Millisecond * 100)
	}()

	conns := cn.Mes.Conns()

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
	if inConns > cn.MaxPeersAllowed-1 {
		conns[0].Close()

		return 2
	}

	fullCycles := 0

	//try to connect to other peers
	if len(conns) < cn.MaxPeersAllowed && len(knownPeers) > 0 {
		for fullCycles < 2 {
			if cn.indexKnownPeers >= len(knownPeers) {
				//index out of bound, do 0 (restart the list)
				cn.indexKnownPeers = 0
				fullCycles++
			}

			//get the known peerID
			peerID := knownPeers[cn.indexKnownPeers]
			cn.indexKnownPeers++

			if cn.Mes.Connectedness(peerID) == net.NotConnected {
				if cn.OnNeedToConnectToOtherPeer != nil {
					err := cn.OnNeedToConnectToOtherPeer(cn, peerID)
					if err == nil {
						//connection succeed
						return 0
					}
				}
			}
		}
	}

	return 3
}

// called when network starts listening on an addr
func (cn *ConnNotifier) Listen(netw net.Network, ma multiaddr.Multiaddr) {}

// called when network starts listening on an addr
func (cn *ConnNotifier) ListenClose(netw net.Network, ma multiaddr.Multiaddr) {}

// called when a connection opened
func (cn *ConnNotifier) Connected(netw net.Network, conn net.Conn) {
	//fmt.Printf("Connected %s: %v\n", netw.LocalPeer().Pretty(), conn.RemotePeer())

	//refuse other connections
	if cn.MaxPeersAllowed < len(cn.Mes.Conns()) {
		conn.Close()
	}
}

// called when a connection closed
func (cn *ConnNotifier) Disconnected(netw net.Network, conn net.Conn) {
	//fmt.Printf("Disconnected %s: %v\n", netw.LocalPeer().Pretty(), conn.RemotePeer())
}

// called when a stream opened
func (cn *ConnNotifier) OpenedStream(netw net.Network, stream net.Stream) {}

func (cn *ConnNotifier) ClosedStream(netw net.Network, stream net.Stream) {}
