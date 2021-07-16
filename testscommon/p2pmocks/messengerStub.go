package p2pmocks

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessengerStub -
type MessengerStub struct {
	ConnectedFullHistoryPeersOnTopicCalled func(topic string) []core.PeerID
	IDCalled                               func() core.PeerID
	CloseCalled                            func() error
	CreateTopicCalled                      func(name string, createChannelForTopic bool) error
	HasTopicCalled                         func(name string) bool
	HasTopicValidatorCalled                func(name string) bool
	BroadcastOnChannelCalled               func(channel string, topic string, buff []byte)
	BroadcastCalled                        func(topic string, buff []byte)
	RegisterMessageProcessorCalled         func(topic string, identifier string, handler p2p.MessageProcessor) error
	BootstrapCalled                        func() error
	PeerAddressesCalled                    func(pid core.PeerID) []string
	BroadcastOnChannelBlockingCalled       func(channel string, topic string, buff []byte) error
	IsConnectedToTheNetworkCalled          func() bool
	PeersCalled                            func() []core.PeerID
	AddressesCalled                        func() []string
	ConnectToPeerCalled                    func(address string) error
	IsConnectedCalled                      func(peerID core.PeerID) bool
	ConnectedPeersCalled                   func() []core.PeerID
	ConnectedAddressesCalled               func() []string
	ConnectedPeersOnTopicCalled            func(topic string) []core.PeerID
	UnregisterAllMessageProcessorsCalled   func() error
	UnregisterMessageProcessorCalled       func(topic string, identifier string) error
	SendToConnectedPeerCalled              func(topic string, buff []byte, peerID core.PeerID) error
	ThresholdMinConnectedPeersCalled       func() int
	SetThresholdMinConnectedPeersCalled    func(minConnectedPeers int) error
	SetPeerShardResolverCalled             func(peerShardResolver p2p.PeerShardResolver) error
	SetPeerDenialEvaluatorCalled           func(handler p2p.PeerDenialEvaluator) error
	GetConnectedPeersInfoCalled            func() *p2p.ConnectedPeersInfo
	UnjoinAllTopicsCalled                  func() error
	PortCalled                             func() int
}

// ConnectedFullHistoryPeersOnTopic -
func (ms *MessengerStub) ConnectedFullHistoryPeersOnTopic(topic string) []core.PeerID {
	if ms.ConnectedFullHistoryPeersOnTopicCalled != nil {
		return ms.ConnectedFullHistoryPeersOnTopicCalled(topic)
	}

	return make([]core.PeerID, 0)
}

// ID -
func (ms *MessengerStub) ID() core.PeerID {
	if ms.IDCalled != nil {
		return ms.IDCalled()
	}

	return ""
}

// RegisterMessageProcessor -
func (ms *MessengerStub) RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error {
	if ms.RegisterMessageProcessorCalled != nil {
		return ms.RegisterMessageProcessorCalled(topic, identifier, handler)
	}

	return nil
}

// Broadcast -
func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	if ms.BroadcastCalled != nil {
		ms.BroadcastCalled(topic, buff)
	}
}

// Close -
func (ms *MessengerStub) Close() error {
	if ms.CloseCalled != nil {
		return ms.CloseCalled()
	}

	return nil
}

// CreateTopic -
func (ms *MessengerStub) CreateTopic(name string, createChannelForTopic bool) error {
	if ms.CreateTopicCalled != nil {
		return ms.CreateTopicCalled(name, createChannelForTopic)
	}

	return nil
}

// HasTopic -
func (ms *MessengerStub) HasTopic(name string) bool {
	if ms.HasTopicCalled != nil {
		return ms.HasTopicCalled(name)
	}

	return false
}

// HasTopicValidator -
func (ms *MessengerStub) HasTopicValidator(name string) bool {
	if ms.HasTopicValidatorCalled != nil {
		return ms.HasTopicValidatorCalled(name)
	}

	return false
}

// BroadcastOnChannel -
func (ms *MessengerStub) BroadcastOnChannel(channel string, topic string, buff []byte) {
	if ms.BroadcastOnChannelCalled != nil {
		ms.BroadcastOnChannelCalled(channel, topic, buff)
	}
}

// Bootstrap -
func (ms *MessengerStub) Bootstrap() error {
	if ms.BootstrapCalled != nil {
		return ms.BootstrapCalled()
	}

	return nil
}

// PeerAddresses -
func (ms *MessengerStub) PeerAddresses(pid core.PeerID) []string {
	if ms.PeerAddressesCalled != nil {
		return ms.PeerAddressesCalled(pid)
	}

	return make([]string, 0)
}

// BroadcastOnChannelBlocking -
func (ms *MessengerStub) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	if ms.BroadcastOnChannelBlockingCalled != nil {
		return ms.BroadcastOnChannelBlockingCalled(channel, topic, buff)
	}

	return nil
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
	if ms.AddressesCalled != nil {
		return ms.AddressesCalled()
	}

	return make([]string, 0)
}

// ConnectToPeer -
func (ms *MessengerStub) ConnectToPeer(address string) error {
	if ms.ConnectToPeerCalled != nil {
		return ms.ConnectToPeerCalled(address)
	}

	return nil
}

// IsConnected -
func (ms *MessengerStub) IsConnected(peerID core.PeerID) bool {
	if ms.IsConnectedCalled != nil {
		return ms.IsConnectedCalled(peerID)
	}

	return false
}

// ConnectedPeers -
func (ms *MessengerStub) ConnectedPeers() []core.PeerID {
	if ms.ConnectedPeersCalled != nil {
		return ms.ConnectedPeersCalled()
	}

	return make([]core.PeerID, 0)
}

// ConnectedAddresses -
func (ms *MessengerStub) ConnectedAddresses() []string {
	if ms.ConnectedAddressesCalled != nil {
		return ms.ConnectedAddressesCalled()
	}

	return make([]string, 0)
}

// ConnectedPeersOnTopic -
func (ms *MessengerStub) ConnectedPeersOnTopic(topic string) []core.PeerID {
	if ms.ConnectedPeersOnTopicCalled != nil {
		return ms.ConnectedPeersOnTopicCalled(topic)
	}

	return make([]core.PeerID, 0)
}

// UnregisterAllMessageProcessors -
func (ms *MessengerStub) UnregisterAllMessageProcessors() error {
	if ms.UnregisterAllMessageProcessorsCalled != nil {
		return ms.UnregisterAllMessageProcessorsCalled()
	}

	return nil
}

// UnregisterMessageProcessor -
func (ms *MessengerStub) UnregisterMessageProcessor(topic string, identifier string) error {
	if ms.UnregisterMessageProcessorCalled != nil {
		return ms.UnregisterMessageProcessorCalled(topic, identifier)
	}

	return nil
}

// SendToConnectedPeer -
func (ms *MessengerStub) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	if ms.SendToConnectedPeerCalled != nil {
		return ms.SendToConnectedPeerCalled(topic, buff, peerID)
	}

	return nil
}

// ThresholdMinConnectedPeers -
func (ms *MessengerStub) ThresholdMinConnectedPeers() int {
	if ms.ThresholdMinConnectedPeersCalled != nil {
		return ms.ThresholdMinConnectedPeersCalled()
	}

	return 0
}

// SetThresholdMinConnectedPeers -
func (ms *MessengerStub) SetThresholdMinConnectedPeers(minConnectedPeers int) error {
	if ms.SetThresholdMinConnectedPeersCalled != nil {
		return ms.SetThresholdMinConnectedPeersCalled(minConnectedPeers)
	}

	return nil
}

// SetPeerShardResolver -
func (ms *MessengerStub) SetPeerShardResolver(peerShardResolver p2p.PeerShardResolver) error {
	if ms.SetPeerShardResolverCalled != nil {
		return ms.SetPeerShardResolverCalled(peerShardResolver)
	}

	return nil
}

// SetPeerDenialEvaluator -
func (ms *MessengerStub) SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error {
	if ms.SetPeerDenialEvaluatorCalled != nil {
		return ms.SetPeerDenialEvaluatorCalled(handler)
	}

	return nil
}

// GetConnectedPeersInfo -
func (ms *MessengerStub) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	if ms.GetConnectedPeersInfoCalled != nil {
		return ms.GetConnectedPeersInfoCalled()
	}

	return nil
}

// UnjoinAllTopics -
func (ms *MessengerStub) UnjoinAllTopics() error {
	if ms.UnjoinAllTopicsCalled != nil {
		return ms.UnjoinAllTopicsCalled()
	}

	return nil
}

// Port -
func (ms *MessengerStub) Port() int {
	if ms.PortCalled != nil {
		return ms.PortCalled()
	}

	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
