package libp2p

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/libp2p/go-libp2p-core/connmgr"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p/p2p/net/mock"
)

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

	lctx, err := NewLibp2pContext(ctx, NewConnectableHost(host.Host(h)))
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

// NewNetworkMessengerOnFreePort tries to create a new NetworkMessenger on a free port found in the system
// Should be used only in testing!
func NewNetworkMessengerOnFreePort(ctx context.Context,
	p2pPrivKey libp2pCrypto.PrivKey,
	conMgr connmgr.ConnManager,
	outgoingPLB p2p.ChannelLoadBalancer,
	peerDiscoverer p2p.PeerDiscoverer,
) (*networkMessenger, error) {
	return NewNetworkMessenger(
		ctx,
		0,
		p2pPrivKey,
		conMgr,
		outgoingPLB,
		peerDiscoverer,
		ListenLocalhostAddrWithIp4AndTcp,
	)
}
