package libp2p

import (
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
	err := checkParameters(args)
	if err != nil {
		return nil, err
	}
	if mockNet == nil {
		return nil, p2p.ErrNilMockNet
	}

	h, err := mockNet.GenPeer()
	if err != nil {
		return nil, err
	}

	mes, err := createMessenger(args, h, false)
	if err != nil {
		return nil, err
	}

	return mes, err
}
