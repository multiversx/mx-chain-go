package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type MessageHandlerStub struct {
	ConnectedPeersOnTopicCalled func(topic string) []p2p.PeerID
	SendToConnectedPeerCalled   func(topic string, buff []byte, peerID p2p.PeerID) error
}

func (mhs *MessageHandlerStub) ConnectedPeersOnTopic(topic string) []p2p.PeerID {
	return mhs.ConnectedPeersOnTopicCalled(topic)
}

func (mhs *MessageHandlerStub) SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	return mhs.SendToConnectedPeerCalled(topic, buff, peerID)
}
