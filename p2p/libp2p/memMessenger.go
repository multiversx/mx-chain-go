package libp2p

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/net/mock"
)

// NewMemoryMessenger creates a new sandbox testable instance of libP2P messenger
// It should not open ports on current machine
// Should be used only in testing!
func NewMemoryMessenger(
	ctx context.Context,
	mockNet mocknet.Mocknet,
	peerDiscoveryType p2p.PeerDiscoveryType) (*networkMessenger, error) {

	if ctx == nil {
		return nil, p2p.ErrNilContext
	}

	if mockNet == nil {
		return nil, p2p.ErrNilMockNet
	}

	h, err := mockNet.GenPeer()
	if err != nil {
		return nil, err
	}

	mes, err := createMessenger(
		ctx,
		h,
		false,
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		peerDiscoveryType,
	)
	if err != nil {
		return nil, err
	}

	mes.preconnectPeerHandler = func(pInfo peerstore.PeerInfo) {
		_ = mockNet.LinkAll()
	}

	return mes, err
}
