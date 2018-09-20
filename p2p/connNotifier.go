package p2p

import (
	_ "fmt"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
	"sync"
	"time"
)

type ConnNotifier struct {
	node *Node

	stopConnKeeper chan bool

	mutStat sync.RWMutex
	stat    GoRoutineStat

	MaxPeersAllowed int

	OnDoSimpleTask             func(sender *ConnNotifier)
	OnGetKnownPeers            func(sender *ConnNotifier) []peer.ID
	OnNeedToConnectToOtherPeer func(sender *ConnNotifier, pid peer.ID) error

	indexKnownPeers int
}

func NewConnNotifier(n *Node) *ConnNotifier {
	if n == nil {
		panic("Nil node!")
	}

	return &ConnNotifier{node: n, stopConnKeeper: make(chan bool, 1), stat: CLOSED}
}

func (cn *ConnNotifier) Stat() GoRoutineStat {
	cn.mutStat.RLock()
	defer cn.mutStat.RUnlock()

	return cn.stat
}

func (cn *ConnNotifier) Start() {
	cn.mutStat.Lock()
	if cn.stat != CLOSED {
		cn.mutStat.Unlock()
		return
	}

	cn.stat = STARTED
	cn.mutStat.Unlock()

	go cn.connKeeper()
}

func (cn *ConnNotifier) Stop() {
	cn.mutStat.Lock()
	defer cn.mutStat.Unlock()

	if cn.stat == STARTED {
		cn.stat = CLOSING

		cn.stopConnKeeper <- true
	}

}

func (cn *ConnNotifier) connKeeper() {
	defer func() {
		cn.mutStat.Lock()
		cn.stat = CLOSED
		cn.mutStat.Unlock()
	}()

	for {

		select {
		default:
			if cn.OnDoSimpleTask != nil {
				cn.OnDoSimpleTask(cn)
			}
		case <-cn.stopConnKeeper:
			return
		}
	}
}

func TaskMonitorConnections(cn *ConnNotifier) (testOutput int) {
	if cn.MaxPeersAllowed < 1 {
		//won't try to connect to other peers
		return 1
	}

	defer func() {
		time.Sleep(time.Millisecond * 100)
	}()

	conns := cn.node.P2pNode.Network().Conns()

	knownPeers := []peer.ID{}

	if cn.OnGetKnownPeers != nil {
		knownPeers = cn.OnGetKnownPeers(cn)
	}

	inConns := 0
	outConns := 0

	for _, conn := range conns {
		if conn.Stat().Direction == net.DirInbound {
			inConns++
		}

		if conn.Stat().Direction == net.DirOutbound {
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

			if cn.node.P2pNode.Network().Connectedness(peerID) == net.NotConnected {
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
	if cn.MaxPeersAllowed < len(cn.node.P2pNode.Network().Conns()) {
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

//fmt.Printf("Listen %s: %v\n", netw.LocalPeer().Pretty(), ma)
//fmt.Printf("Listen close %s: %v\n", netw.LocalPeer().Pretty(), ma)
//fmt.Printf("Disconnected %s: %v\n", netw.LocalPeer().Pretty(), conn.RemotePeer())
//fmt.Printf("Opened stream %s: %v\n", netw.LocalPeer().Pretty(), stream.Conn().RemotePeer())
//fmt.Printf("Closed stream %s: %v\n", netw.LocalPeer().Pretty(), stream.Conn().RemotePeer())
