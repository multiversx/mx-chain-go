package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// P2PAntifloodHandlerStub -
type P2PAntifloodHandlerStub struct {
	CanProcessMessageCalled         func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopicCalled func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
}

// ResetForTopic -
func (p2pahs *P2PAntifloodHandlerStub) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic -
func (p2pahs *P2PAntifloodHandlerStub) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// CanProcessMessage -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if p2pahs.CanProcessMessageCalled == nil {
		return nil
	}

	return p2pahs.CanProcessMessageCalled(message, fromConnectedPeer)
}

// CanProcessMessagesOnTopic -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
	if p2pahs.CanProcessMessagesOnTopicCalled == nil {
		return nil
	}

	return p2pahs.CanProcessMessagesOnTopicCalled(peer, topic, numMessages, totalSize, sequence)
}

// IsInterfaceNil -
func (p2pahs *P2PAntifloodHandlerStub) IsInterfaceNil() bool {
	return p2pahs == nil
}
