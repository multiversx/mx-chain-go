package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessengerStub -
type MessengerStub struct {
	IDCalled                         func() core.PeerID
	CloseCalled                      func() error
	CreateTopicCalled                func(name string, createChannelForTopic bool) error
	HasTopicCalled                   func(name string) bool
	HasTopicValidatorCalled          func(name string) bool
	BroadcastOnChannelCalled         func(channel string, topic string, buff []byte)
	BroadcastCalled                  func(topic string, buff []byte)
	RegisterMessageProcessorCalled   func(topic string, handler p2p.MessageProcessor) error
	BootstrapCalled                  func() error
	PeerAddressesCalled              func(pid core.PeerID) []string
	BroadcastOnChannelBlockingCalled func(channel string, topic string, buff []byte) error
	IsConnectedToTheNetworkCalled    func() bool
	PeersCalled                      func() []core.PeerID
}

// ID -
func (ms *MessengerStub) ID() core.PeerID {
	if ms.IDCalled != nil {
		return ms.IDCalled()
	}

	return ""
}

// RegisterMessageProcessor -
func (ms *MessengerStub) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	if ms.RegisterMessageProcessorCalled != nil {
		return ms.RegisterMessageProcessorCalled(topic, handler)
	}
	return nil
}

// Broadcast -
func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	ms.BroadcastCalled(topic, buff)
}

// Close -
func (ms *MessengerStub) Close() error {
	return ms.CloseCalled()
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

// Bootstrap -
func (ms *MessengerStub) Bootstrap() error {
	return ms.BootstrapCalled()
}

// PeerAddresses -
func (ms *MessengerStub) PeerAddresses(pid core.PeerID) []string {
	return ms.PeerAddressesCalled(pid)
}

// BroadcastOnChannelBlocking -
func (ms *MessengerStub) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	return ms.BroadcastOnChannelBlockingCalled(channel, topic, buff)
}

// IsConnectedToTheNetwork -
func (ms *MessengerStub) IsConnectedToTheNetwork() bool {
	if ms.IsConnectedToTheNetworkCalled != nil {
		return ms.IsConnectedToTheNetworkCalled()
	}

	return true
}

// Peers -
func (ms *MessengerStub) Peers() []core.PeerID {
	if ms.PeersCalled != nil {
		return ms.PeersCalled()
	}

	return make([]core.PeerID, 0)
}

// Addresses -
func (ms *MessengerStub) Addresses() []string {
	return make([]string, 0)
}

// ConnectToPeer -
func (ms *MessengerStub) ConnectToPeer(_ string) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}

// IsConnected
func (ms *MessengerStub) IsConnected(_ core.PeerID) bool {
	return false
}

// ConnectedPeers -
func (ms *MessengerStub) ConnectedPeers() []core.PeerID {
	return nil
}

// ConnectedAddresses -
func (ms *MessengerStub) ConnectedAddresses() []string {
	return nil
}

// ConnectedPeersOnTopic -
func (ms *MessengerStub) ConnectedPeersOnTopic(_ string) []core.PeerID {
	return nil
}

// UnregisterAllMessageProcessors -
func (ms *MessengerStub) UnregisterAllMessageProcessors() error {
	return nil
}

// UnregisterMessageProcessor -
func (ms *MessengerStub) UnregisterMessageProcessor(_ string) error {
	return nil
}

// SendToConnectedPeer -
func (ms *MessengerStub) SendToConnectedPeer(_ string, _ []byte, _ core.PeerID) error {
	return nil
}

// ThresholdMinConnectedPeers -
func (ms *MessengerStub) ThresholdMinConnectedPeers() int {
	return 0
}

// SetThresholdMinConnectedPeers -
func (ms *MessengerStub) SetThresholdMinConnectedPeers(_ int) error {
	return nil
}

// SetPeerShardResolver -
func (ms *MessengerStub) SetPeerShardResolver(_ p2p.PeerShardResolver) error {
	return nil
}

// SetPeerDenialEvaluator -
func (ms *MessengerStub) SetPeerDenialEvaluator(_ p2p.PeerDenialEvaluator) error {
	return nil
}

// GetConnectedPeersInfo -
func (ms *MessengerStub) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	return nil
}

// UnjoinAllTopics -
func (ms *MessengerStub) UnjoinAllTopics() error {
	return nil
}
