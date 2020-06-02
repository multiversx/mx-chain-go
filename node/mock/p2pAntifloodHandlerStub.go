package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

// P2PAntifloodHandlerStub -
type P2PAntifloodHandlerStub struct {
	CanProcessMessageCalled         func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error
	CanProcessMessagesOnTopicCalled func(peer p2p.PeerID, topic string, numMessages uint32) error
	ApplyConsensusSizeCalled        func(size int)
}

// ResetForTopic -
func (p2pahs *P2PAntifloodHandlerStub) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic -
func (p2pahs *P2PAntifloodHandlerStub) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// ApplyConsensusSize -
func (p2pahs *P2PAntifloodHandlerStub) ApplyConsensusSize(size int) {
	if p2pahs.ApplyConsensusSizeCalled != nil {
		p2pahs.ApplyConsensusSizeCalled(size)
	}
}

// CanProcessMessage -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	if p2pahs.CanProcessMessageCalled == nil {
		return nil
	}

	return p2pahs.CanProcessMessageCalled(message, fromConnectedPeer)
}

// CanProcessMessagesOnTopic -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessagesOnTopic(peer p2p.PeerID, topic string, numMessages uint32) error {
	if p2pahs.CanProcessMessagesOnTopicCalled == nil {
		return nil
	}

	return p2pahs.CanProcessMessagesOnTopicCalled(peer, topic, numMessages)
}

// IsInterfaceNil -
func (p2pahs *P2PAntifloodHandlerStub) IsInterfaceNil() bool {
	return p2pahs == nil
}
