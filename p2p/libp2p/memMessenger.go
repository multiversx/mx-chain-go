package libp2p

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	crypto "github.com/libp2p/go-libp2p-crypto"
	ifconnmgr "github.com/libp2p/go-libp2p-interface-connmgr"
	"github.com/libp2p/go-libp2p/p2p/net/mock"
)

const maxPort = 65535

var portMutex *sync.Mutex

func init() {
	portMutex = &sync.Mutex{}
}

// NewMemoryMessenger creates a new sandbox testable instance of libP2P messenger
// It should not open ports on current machine
// Should be used only in testing!
func NewMemoryMessenger(
	ctx context.Context,
	mockNet mocknet.Mocknet,
	peerDiscoverer p2p.PeerDiscoverer) (*networkMessenger, error) {

	if ctx == nil {
		return nil, p2p.ErrNilContext
	}

	if mockNet == nil {
		return nil, p2p.ErrNilMockNet
	}

	if peerDiscoverer == nil {
		return nil, p2p.ErrNilPeerDiscoverer
	}

	h, err := mockNet.GenPeer()
	if err != nil {
		return nil, err
	}

	lctx, err := NewLibp2pContext(ctx, NewConnectableHost(h))
	if err != nil {
		log.LogIfError(h.Close())
		return nil, err
	}

	mes, err := createMessenger(
		lctx,
		false,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		peerDiscoverer,
	)
	if err != nil {
		return nil, err
	}

	return mes, err
}

// NewNetworkMessengerWithPortSweep tries to create a new NetworkMessenger on the given port
// If the port is opened, will try the next one until it succeeds or fails for another reason other than
// occupied port
// Should be used only in testing!
func NewNetworkMessengerWithPortSweep(ctx context.Context,
	port int,
	p2pPrivKey crypto.PrivKey,
	conMgr ifconnmgr.ConnManager,
	outgoingPLB p2p.ChannelLoadBalancer,
	peerDiscoverer p2p.PeerDiscoverer,
) (*networkMessenger, int, error) {
	portMutex.Lock()
	defer portMutex.Unlock()

	for ; port < maxPort; port++ {
		if isTcpPortFree(port) {
			mes, err := NewNetworkMessenger(
				ctx,
				port,
				p2pPrivKey,
				conMgr,
				outgoingPLB,
				peerDiscoverer,
			)

			return mes, port, err
		}
	}

	return nil, port, p2p.ErrNoUsablePortsOnMachine
}

func isTcpPortFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if ln != nil {
		_ = ln.Close()
	}

	return err == nil
}
