package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// MessageHandlerStub -
type MessageHandlerStub struct {
	ConnectedPeersOnTopicCalled func(topic string) []core.PeerID
	SendToConnectedPeerCalled   func(topic string, buff []byte, peerID core.PeerID) error
	IDCalled                    func() core.PeerID
	ConnectedPeersCalled        func() []core.PeerID
	IsConnectedCalled           func(peerID core.PeerID) bool
}

// ConnectedPeersOnTopic -
func (mhs *MessageHandlerStub) ConnectedPeersOnTopic(topic string) []core.PeerID {
	return mhs.ConnectedPeersOnTopicCalled(topic)
}

// SendToConnectedPeer -
func (mhs *MessageHandlerStub) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	return mhs.SendToConnectedPeerCalled(topic, buff, peerID)
}

// ID -
func (mhs *MessageHandlerStub) ID() core.PeerID {
	if mhs.IDCalled != nil {
		return mhs.IDCalled()
	}

	return ""
}

// ConnectedPeers -
func (mhs *MessageHandlerStub) ConnectedPeers() []core.PeerID {
	if mhs.ConnectedPeersCalled != nil {
		return mhs.ConnectedPeersCalled()
	}

	return make([]core.PeerID, 0)
}

// IsConnected -
func (mhs *MessageHandlerStub) IsConnected(peerID core.PeerID) bool {
	if mhs.IsConnectedCalled != nil {
		return mhs.IsConnectedCalled(peerID)
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (mhs *MessageHandlerStub) IsInterfaceNil() bool {
	return mhs == nil
}
