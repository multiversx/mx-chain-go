package p2pmocks

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
)

// MessengerStub -
type MessengerStub struct {
	ConnectedFullHistoryPeersOnTopicCalled  func(topic string) []core.PeerID
	IDCalled                                func() core.PeerID
	CloseCalled                             func() error
	CreateTopicCalled                       func(name string, createChannelForTopic bool) error
	HasTopicCalled                          func(name string) bool
	HasTopicValidatorCalled                 func(name string) bool
	BroadcastOnChannelCalled                func(channel string, topic string, buff []byte)
	BroadcastCalled                         func(topic string, buff []byte)
	RegisterMessageProcessorCalled          func(topic string, identifier string, handler p2p.MessageProcessor) error
	BootstrapCalled                         func() error
	PeerAddressesCalled                     func(pid core.PeerID) []string
	IsConnectedToTheNetworkCalled           func() bool
	PeersCalled                             func() []core.PeerID
	AddressesCalled                         func() []string
	ConnectToPeerCalled                     func(address string) error
	IsConnectedCalled                       func(peerID core.PeerID) bool
	ConnectedPeersCalled                    func() []core.PeerID
	ConnectedAddressesCalled                func() []string
	ConnectedPeersOnTopicCalled             func(topic string) []core.PeerID
	UnregisterAllMessageProcessorsCalled    func() error
	UnregisterMessageProcessorCalled        func(topic string, identifier string) error
	SendToConnectedPeerCalled               func(topic string, buff []byte, peerID core.PeerID) error
	ThresholdMinConnectedPeersCalled        func() int
	SetThresholdMinConnectedPeersCalled     func(minConnectedPeers int) error
	SetPeerShardResolverCalled              func(peerShardResolver p2p.PeerShardResolver) error
	SetPeerDenialEvaluatorCalled            func(handler p2p.PeerDenialEvaluator) error
	GetConnectedPeersInfoCalled             func() *p2p.ConnectedPeersInfo
	UnJoinAllTopicsCalled                   func() error
	PortCalled                              func() int
	WaitForConnectionsCalled                func(maxWaitingTime time.Duration, minNumOfPeers uint32)
	SignCalled                              func(payload []byte) ([]byte, error)
	VerifyCalled                            func(payload []byte, pid core.PeerID, signature []byte) error
	AddPeerTopicNotifierCalled              func(notifier p2p.PeerTopicNotifier) error
	BroadcastUsingPrivateKeyCalled          func(topic string, buff []byte, pid core.PeerID, skBytes []byte)
	BroadcastOnChannelUsingPrivateKeyCalled func(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte)
	SignUsingPrivateKeyCalled               func(skBytes []byte, payload []byte) ([]byte, error)
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

	return "peer ID"
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

// UnJoinAllTopics -
func (ms *MessengerStub) UnJoinAllTopics() error {
	if ms.UnJoinAllTopicsCalled != nil {
		return ms.UnJoinAllTopicsCalled()
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

// WaitForConnections -
func (ms *MessengerStub) WaitForConnections(maxWaitingTime time.Duration, minNumOfPeers uint32) {
	if ms.WaitForConnectionsCalled != nil {
		ms.WaitForConnectionsCalled(maxWaitingTime, minNumOfPeers)
	}
}

// Sign -
func (ms *MessengerStub) Sign(payload []byte) ([]byte, error) {
	if ms.SignCalled != nil {
		return ms.SignCalled(payload)
	}

	return make([]byte, 0), nil
}

// Verify -
func (ms *MessengerStub) Verify(payload []byte, pid core.PeerID, signature []byte) error {
	if ms.VerifyCalled != nil {
		return ms.VerifyCalled(payload, pid, signature)
	}

	return nil
}

// AddPeerTopicNotifier -
func (ms *MessengerStub) AddPeerTopicNotifier(notifier p2p.PeerTopicNotifier) error {
	if ms.AddPeerTopicNotifierCalled != nil {
		return ms.AddPeerTopicNotifierCalled(notifier)
	}

	return nil
}

// BroadcastUsingPrivateKey -
func (ms *MessengerStub) BroadcastUsingPrivateKey(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	if ms.BroadcastUsingPrivateKeyCalled != nil {
		ms.BroadcastUsingPrivateKeyCalled(topic, buff, pid, skBytes)
	}
}

// BroadcastOnChannelUsingPrivateKey -
func (ms *MessengerStub) BroadcastOnChannelUsingPrivateKey(channel string, topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	if ms.BroadcastOnChannelUsingPrivateKeyCalled != nil {
		ms.BroadcastOnChannelUsingPrivateKeyCalled(channel, topic, buff, pid, skBytes)
	}
}

// SignUsingPrivateKey -
func (ms *MessengerStub) SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error) {
	if ms.SignUsingPrivateKeyCalled != nil {
		return ms.SignUsingPrivateKeyCalled(skBytes, payload)
	}

	return make([]byte, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
