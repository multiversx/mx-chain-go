package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

// P2PAntifloodHandlerStub -
type P2PAntifloodHandlerStub struct {
	CanProcessMessageCalled        func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error
	CanProcessMessageOnTopicCalled func(peer p2p.PeerID, topic string) error
}

// ResetForTopic -
func (p2pahs *P2PAntifloodHandlerStub) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic -
func (p2pahs *P2PAntifloodHandlerStub) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// CanProcessMessage -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	if p2pahs.CanProcessMessageCalled == nil {
		return nil
	}

	return p2pahs.CanProcessMessageCalled(message, fromConnectedPeer)
}

// CanProcessMessageOnTopic -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessageOnTopic(peer p2p.PeerID, topic string) error {
	if p2pahs.CanProcessMessageOnTopicCalled == nil {
		return nil
	}

	return p2pahs.CanProcessMessageOnTopicCalled(peer, topic)
}

// IsInterfaceNil -
func (p2pahs *P2PAntifloodHandlerStub) IsInterfaceNil() bool {
	return p2pahs == nil
}
