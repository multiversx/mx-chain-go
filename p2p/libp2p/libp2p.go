package libp2p

import (
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
)

// ArgsNetworkMessenger defines the options used to create a p2p wrapper
type ArgsNetworkMessenger = libp2p.ArgsNetworkMessenger

// NewNetworkMessenger creates a libP2P messenger by opening a port on the current machine
func NewNetworkMessenger(args ArgsNetworkMessenger) (p2p.Messenger, error) {
	return libp2p.NewNetworkMessenger(args)
}

// NewMockMessenger creates a new sandbox testable instance of libP2P messenger
// It should not open ports on current machine
// Should be used only in testing!
func NewMockMessenger(
	args ArgsNetworkMessenger,
	mockNet mocknet.Mocknet,
) (p2p.Messenger, error) {
	return libp2p.NewMockMessenger(args, mockNet)
}

// LocalSyncTimer uses the local system to provide the current time
type LocalSyncTimer = libp2p.LocalSyncTimer
