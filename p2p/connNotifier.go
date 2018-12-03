package p2p

import (
	"fmt"
	"sync"
	"time"

	//"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
)

// durRefreshConnections represents the duration used to pause between refreshing connections to known peers
const durRefreshConnections = 1000 * time.Millisecond

//the value of 2 was chosen as a mean to iterate the known peers at least 1 time but not 2 times
const maxFullCycles = 2

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
	stopChan     chan bool
	stoppedChan  chan bool
	mutIsRunning sync.RWMutex
	isRunning    bool

	maxAllowedPeers int

	// GetKnownPeers is a pointer to a function that will return all known peers found by a Messenger
	GetKnownPeers func(sender *ConnNotifier) []peer.ID
	// ConnectToPeer is a pointer to a function that has to make the Messenger object to connect ta peerID
	ConnectToPeer func(sender *ConnNotifier, pid peer.ID) error
	// GetConnections is a pointer to a function that returns a snapshot of all known connections held by a Messenger
	GetConnections func(sender *ConnNotifier) []net.Conn
	// IsConnected is a pointer to a function that returns tre if current messenger is connected to peer with ID = pid
	IsConnected func(sender *ConnNotifier, pid peer.ID) bool

	indexKnownPeers int
}

// NewConnNotifier will create a new object
func NewConnNotifier(maxAllowedPeers int) *ConnNotifier {
	cn := ConnNotifier{
		maxAllowedPeers: maxAllowedPeers,
		mutIsRunning:    sync.RWMutex{},
		stopChan:        make(chan bool, 0),
		stoppedChan:     make(chan bool, 0),
	}

	return &cn
}

// TaskResolveConnections resolves the connections to other peers. It should not be called too often as the
// connections are not done instantly. Even if the connection is made in a short time, there is a delay
// until the connected peer might close down the connections because it reached the maximum limit.
// This function handles the array connections that mdns service provides.
// This function always tries to find a new connection by closing the oldest one.
// It tries to create a new outbound connection by iterating over known peers for at least one cycle but not 2 or more.
func (cn *ConnNotifier) TaskResolveConnections() ResultType {
	if cn.maxAllowedPeers < 1 {
		//won't try to connect to other peers
		return WontConnect
	}

	conns := cn.getConnections()
	knownPeers := cn.getKnownPeers()
	inConns, _ := cn.computeInboundOutboundConns(conns)

	//test whether we only have inbound connection (security issue)
	if inConns >= cn.maxAllowedPeers {
		_ = conns[0].Close()
		//TODO log error

		return OnlyInboundConnections
	}

	//try to connect to other peers
	if len(conns) < cn.maxAllowedPeers && len(knownPeers) > 0 {
		return cn.iterateThroughPeersAndTryToConnect(knownPeers)
	}

	return NothingDone
}

func (cn *ConnNotifier) getConnections() []net.Conn {
	if cn.GetConnections != nil {
		return cn.GetConnections(cn)
	} else {
		return make([]net.Conn, 0)
	}
}

func (cn *ConnNotifier) getKnownPeers() []peer.ID {
	if cn.GetKnownPeers != nil {
		return cn.GetKnownPeers(cn)
	} else {
		return make([]peer.ID, 0)
	}
}

func (cn *ConnNotifier) computeInboundOutboundConns(conns []net.Conn) (inConns, outConns int) {
	//get how many inbound and outbound connection we have
	for i := 0; i < len(conns); i++ {
		if conns[i].Stat().Direction == net.DirInbound {
			inConns++
		}

		if conns[i].Stat().Direction == net.DirOutbound {
			outConns++
		}
	}

	return
}

func (cn *ConnNotifier) iterateThroughPeersAndTryToConnect(knownPeers []peer.ID) ResultType {
	fullCycles := 0

	for fullCycles < maxFullCycles {
		if cn.indexKnownPeers >= len(knownPeers) {
			//index out of bound, do 0 (restart the list)
			cn.indexKnownPeers = 0
			fullCycles++
		}

		//get the known peerID
		peerID := knownPeers[cn.indexKnownPeers]
		cn.indexKnownPeers++

		//func pointers are associated
		if cn.ConnectToPeer != nil && cn.IsConnected != nil {
			isConnected := cn.IsConnected(cn, peerID)

			if !isConnected {
				err := cn.ConnectToPeer(cn, peerID)

				if err == nil {
					return SuccessfullyConnected
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
	if cn.GetConnections == nil {
		_ = conn.Close()
		//TODO log error
		return
	}

	conns := cn.GetConnections(cn)

	//refuse other connections if max connection has been reached
	if cn.maxAllowedPeers < len(conns) {
		_ = conn.Close()
		//TODO log error
		return
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

// Starts the ConnNotifier main process
func (cn *ConnNotifier) Start() {
	cn.mutIsRunning.Lock()
	defer cn.mutIsRunning.Unlock()

	if cn.isRunning {
		return
	}

	cn.isRunning = true
	go cn.maintainPeers()
}

// Stops the ConnNotifier main process
func (cn *ConnNotifier) Stop() {
	cn.mutIsRunning.Lock()
	defer cn.mutIsRunning.Unlock()

	if !cn.isRunning {
		return
	}

	cn.isRunning = false

	//send stop notification
	cn.stopChan <- true
	//await to finalise	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	<-cn.stoppedChan
}

// maintainPeers is a routine that periodically calls TaskResolveConnections to resolve peer connections
func (cn *ConnNotifier) maintainPeers() {
	for {
		select {
		case <-cn.stopChan:
			//TODO logger
			fmt.Println("ConnNotifier object has stopped!")
			cn.stoppedChan <- true
			return
		case <-time.After(durRefreshConnections):
		}

		cn.TaskResolveConnections()
	}
}
