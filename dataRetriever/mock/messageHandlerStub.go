package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// MessageHandlerStub -
type MessageHandlerStub struct {
	ConnectedPeersOnTopicCalled func(topic string) []core.PeerID
	SendToConnectedPeerCalled   func(topic string, buff []byte, peerID core.PeerID) error
	IDCalled                    func() core.PeerID
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

// IsInterfaceNil returns true if there is no value under the interface
func (mhs *MessageHandlerStub) IsInterfaceNil() bool {
	return mhs == nil
}
