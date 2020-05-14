package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessengerStub -
type MessengerStub struct {
	ConnectedPeersCalled           func() []p2p.PeerID
	RegisterMessageProcessorCalled func(topic string, handler p2p.MessageProcessor) error
}

// ConnectedPeersOnTopic -
func (m *MessengerStub) ConnectedPeersOnTopic(topic string) []p2p.PeerID {
	return []p2p.PeerID{"peer0"}
}

// SendToConnectedPeer -
func (m *MessengerStub) SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	return nil
}

// IsInterfaceNil -
func (m *MessengerStub) IsInterfaceNil() bool {
	return m == nil
}

// HasTopic -
func (m *MessengerStub) HasTopic(name string) bool {
	return false
}

// CreateTopic -
func (m *MessengerStub) CreateTopic(name string, createChannelForTopic bool) error {
	return nil
}

// RegisterMessageProcessor -
func (m *MessengerStub) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	if m.RegisterMessageProcessorCalled != nil {
		return m.RegisterMessageProcessorCalled(topic, handler)
	}

	return nil
}

// UnregisterMessageProcessor -
func (m *MessengerStub) UnregisterMessageProcessor(topic string) error {
	return nil
}

// UnregisterAllMessageProcessors -
func (m *MessengerStub) UnregisterAllMessageProcessors() error {
	return nil
}

// ConnectedPeers -
func (m *MessengerStub) ConnectedPeers() []p2p.PeerID {
	if m.ConnectedPeersCalled != nil {
		return m.ConnectedPeersCalled()
	}

	return []p2p.PeerID{"peer0", "peer1", "peer2", "peer3", "peer4", "peer5"}
}
