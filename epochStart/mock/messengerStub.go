package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessengerStub -
type MessengerStub struct {
	ConnectedPeersCalled           func() []core.PeerID
	RegisterMessageProcessorCalled func(topic string, identifier string, handler p2p.MessageProcessor) error
	UnjoinAllTopicsCalled          func() error
	IDCalled                       func() core.PeerID
}

// ConnectedPeersOnTopic -
func (m *MessengerStub) ConnectedPeersOnTopic(_ string) []core.PeerID {
	return []core.PeerID{"peer0"}
}

// SendToConnectedPeer -
func (m *MessengerStub) SendToConnectedPeer(_ string, _ []byte, _ core.PeerID) error {
	return nil
}

// IsInterfaceNil -
func (m *MessengerStub) IsInterfaceNil() bool {
	return m == nil
}

// HasTopic -
func (m *MessengerStub) HasTopic(_ string) bool {
	return false
}

// CreateTopic -
func (m *MessengerStub) CreateTopic(_ string, _ bool) error {
	return nil
}

// RegisterMessageProcessor -
func (m *MessengerStub) RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error {
	if m.RegisterMessageProcessorCalled != nil {
		return m.RegisterMessageProcessorCalled(topic, identifier, handler)
	}

	return nil
}

// UnregisterMessageProcessor -
func (m *MessengerStub) UnregisterMessageProcessor(_ string, _ string) error {
	return nil
}

// UnregisterAllMessageProcessors -
func (m *MessengerStub) UnregisterAllMessageProcessors() error {
	return nil
}

// UnjoinAllTopics -
func (m *MessengerStub) UnjoinAllTopics() error {
	if m.UnjoinAllTopicsCalled != nil {
		return m.UnjoinAllTopicsCalled()
	}

	return nil
}

// ConnectedPeers -
func (m *MessengerStub) ConnectedPeers() []core.PeerID {
	if m.ConnectedPeersCalled != nil {
		return m.ConnectedPeersCalled()
	}

	return []core.PeerID{"peer0", "peer1", "peer2", "peer3", "peer4", "peer5"}
}

// ID -
func (m *MessengerStub) ID() core.PeerID {
	if m.IDCalled != nil {
		return m.IDCalled()
	}

	return "peer ID"
}
