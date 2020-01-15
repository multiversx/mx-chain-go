package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type P2PAntifloodHandlerStub struct {
	CanProcessMessageCalled        func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error
	CanProcessMessageOnTopicCalled func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID, topic string) error
}

func (p2pahs *P2PAntifloodHandlerStub) ResetForTopic(topic string) {
}

func (p2pahs *P2PAntifloodHandlerStub) SetMaxMessagesForTopic(topic string, maxNum uint32) {
}

func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	return p2pahs.CanProcessMessageCalled(message, fromConnectedPeer)
}

func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessageOnTopic(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID, topic string) error {
	return p2pahs.CanProcessMessageOnTopicCalled(message, fromConnectedPeer, topic)
}

func (p2pahs *P2PAntifloodHandlerStub) IsInterfaceNil() bool {
	return p2pahs == nil
}
