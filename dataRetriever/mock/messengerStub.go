package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessengerStub -
type MessengerStub struct {
	ConnectedPeersOnTopicCalled       func(topic string) []core.PeerID
	CloseCalled                       func() error
	IDCalled                          func() core.PeerID
	PeersCalled                       func() []core.PeerID
	AddressesCalled                   func() []string
	ConnectToPeerCalled               func(address string) error
	TrimConnectionsCalled             func()
	IsConnectedCalled                 func(peerID core.PeerID) bool
	ConnectedPeersCalled              func() []core.PeerID
	CreateTopicCalled                 func(name string, createChannelForTopic bool) error
	HasTopicCalled                    func(name string) bool
	HasTopicValidatorCalled           func(name string) bool
	BroadcastOnChannelCalled          func(channel string, topic string, buff []byte)
	BroadcastCalled                   func(topic string, buff []byte)
	RegisterMessageProcessorCalled    func(topic string, identifier string, handler p2p.MessageProcessor) error
	UnregisterMessageProcessorCalled  func(topic string, identifier string) error
	SendToConnectedPeerCalled         func(topic string, buff []byte, peerID core.PeerID) error
	OutgoingChannelLoadBalancerCalled func() p2p.ChannelLoadBalancer
	BootstrapCalled                   func() error
}

// ConnectedPeersOnTopic -
func (ms *MessengerStub) ConnectedPeersOnTopic(topic string) []core.PeerID {
	if ms.ConnectedPeersOnTopicCalled != nil {
		return ms.ConnectedPeersOnTopicCalled(topic)
	}

	return make([]core.PeerID, 0)
}

// RegisterMessageProcessor -
func (ms *MessengerStub) RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error {
	return ms.RegisterMessageProcessorCalled(topic, identifier, handler)
}

// UnregisterMessageProcessor -
func (ms *MessengerStub) UnregisterMessageProcessor(topic string, identifier string) error {
	return ms.UnregisterMessageProcessorCalled(topic, identifier)
}

// Broadcast -
func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	ms.BroadcastCalled(topic, buff)
}

// OutgoingChannelLoadBalancer -
func (ms *MessengerStub) OutgoingChannelLoadBalancer() p2p.ChannelLoadBalancer {
	return ms.OutgoingChannelLoadBalancerCalled()
}

// Close -
func (ms *MessengerStub) Close() error {
	return ms.CloseCalled()
}

// ID -
func (ms *MessengerStub) ID() core.PeerID {
	if ms.IDCalled != nil {
		return ms.IDCalled()
	}

	return ""
}

// Peers -
func (ms *MessengerStub) Peers() []core.PeerID {
	return ms.PeersCalled()
}

// Addresses -
func (ms *MessengerStub) Addresses() []string {
	return ms.AddressesCalled()
}

// ConnectToPeer -
func (ms *MessengerStub) ConnectToPeer(address string) error {
	return ms.ConnectToPeerCalled(address)
}

// TrimConnections -
func (ms *MessengerStub) TrimConnections() {
	ms.TrimConnectionsCalled()
}

// IsConnected -
func (ms *MessengerStub) IsConnected(peerID core.PeerID) bool {
	return ms.IsConnectedCalled(peerID)
}

// ConnectedPeers -
func (ms *MessengerStub) ConnectedPeers() []core.PeerID {
	return ms.ConnectedPeersCalled()
}

// CreateTopic -
func (ms *MessengerStub) CreateTopic(name string, createChannelForTopic bool) error {
	return ms.CreateTopicCalled(name, createChannelForTopic)
}

// HasTopic -
func (ms *MessengerStub) HasTopic(name string) bool {
	return ms.HasTopicCalled(name)
}

// HasTopicValidator -
func (ms *MessengerStub) HasTopicValidator(name string) bool {
	return ms.HasTopicValidatorCalled(name)
}

// BroadcastOnChannel -
func (ms *MessengerStub) BroadcastOnChannel(channel string, topic string, buff []byte) {
	ms.BroadcastOnChannelCalled(channel, topic, buff)
}

// SendToConnectedPeer -
func (ms *MessengerStub) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	return ms.SendToConnectedPeerCalled(topic, buff, peerID)
}

// Bootstrap -
func (ms *MessengerStub) Bootstrap() error {
	return ms.BootstrapCalled()
}

// UnregisterAllMessageProcessors -
func (ms *MessengerStub) UnregisterAllMessageProcessors() error {
	return nil
}

// UnjoinAllTopics -
func (ms *MessengerStub) UnjoinAllTopics() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
