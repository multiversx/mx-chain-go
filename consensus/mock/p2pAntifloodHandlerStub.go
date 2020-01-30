package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type P2PAntifloodHandlerStub struct {
	CanProcessMessageCalled        func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error
	CanProcessMessageOnTopicCalled func(peer p2p.PeerID, topic string) error
}

func (p2pahs *P2PAntifloodHandlerStub) ResetForTopic(topic string) {
}

func (p2pahs *P2PAntifloodHandlerStub) SetMaxMessagesForTopic(topic string, maxNum uint32) {
}

func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	if p2pahs.CanProcessMessageCalled != nil {
		return p2pahs.CanProcessMessageCalled(message, fromConnectedPeer)
	}

	return nil
}

func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessageOnTopic(peer p2p.PeerID, topic string) error {
	if p2pahs.CanProcessMessageOnTopicCalled != nil {
		return p2pahs.CanProcessMessageOnTopicCalled(peer, topic)
	}

	return nil
}

func (p2pahs *P2PAntifloodHandlerStub) IsInterfaceNil() bool {
	return p2pahs == nil
}
