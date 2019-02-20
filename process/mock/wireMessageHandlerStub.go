package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type WireMessageHandlerStub struct {
	ConnectedPeersCalled      func() []p2p.PeerID
	SendToConnectedPeerCalled func(topic string, buff []byte, peerID p2p.PeerID) error
}

func (wmhs *WireMessageHandlerStub) ConnectedPeers() []p2p.PeerID {
	return wmhs.ConnectedPeersCalled()
}

func (wmhs *WireMessageHandlerStub) SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	return wmhs.SendToConnectedPeerCalled(topic, buff, peerID)
}
