package libp2p

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/p2p"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
)

// NewMockMessenger creates a new sandbox testable instance of libP2P messenger
// It should not open ports on current machine
// Should be used only in testing!
func NewMockMessenger(
	args ArgsNetworkMessenger,
	mockNet mocknet.Mocknet,
) (*networkMessenger, error) {
	if mockNet == nil {
		return nil, p2p.ErrNilMockNet
	}

	h, err := mockNet.GenPeer()
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	p2pNode := &networkMessenger{
		p2pHost:    NewConnectableHost(h),
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}

	err = addComponentsToNode(args, p2pNode, withoutMessageSigning)
	if err != nil {
		return nil, err
	}

	return p2pNode, err
}
