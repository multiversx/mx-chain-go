package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type MessageHandlerStub struct {
	ConnectedPeersCalled      func() []p2p.PeerID
	SendToConnectedPeerCalled func(topic string, buff []byte, peerID p2p.PeerID) error
}

func (mhs *MessageHandlerStub) ConnectedPeers() []p2p.PeerID {
	return mhs.ConnectedPeersCalled()
}

func (mhs *MessageHandlerStub) SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	return mhs.SendToConnectedPeerCalled(topic, buff, peerID)
}
