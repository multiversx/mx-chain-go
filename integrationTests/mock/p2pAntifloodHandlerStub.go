package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

// P2PAntifloodHandlerStub -
type P2PAntifloodHandlerStub struct {
	CanProcessMessageCalled func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error
}

// CanProcessMessage -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	return p2pahs.CanProcessMessageCalled(message, fromConnectedPeer)
}

// IsInterfaceNil -
func (p2pahs *P2PAntifloodHandlerStub) IsInterfaceNil() bool {
	return p2pahs == nil
}
