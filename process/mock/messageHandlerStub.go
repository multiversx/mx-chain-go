package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// MessageHandlerStub -
type MessageHandlerStub struct {
	ConnectedPeersOnTopicCalled func(topic string) []core.PeerID
	SendToConnectedPeerCalled   func(topic string, buff []byte, peerID core.PeerID) error
}

// ConnectedPeersOnTopic -
func (mhs *MessageHandlerStub) ConnectedPeersOnTopic(topic string) []core.PeerID {
	return mhs.ConnectedPeersOnTopicCalled(topic)
}

// SendToConnectedPeer -
func (mhs *MessageHandlerStub) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	return mhs.SendToConnectedPeerCalled(topic, buff, peerID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mhs *MessageHandlerStub) IsInterfaceNil() bool {
	return mhs == nil
}
