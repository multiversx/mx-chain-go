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
	BroadcastOnChannelCalled         func(channel string, topic string, buff []byte)
	BroadcastCalled                  func(topic string, buff []byte)
	RegisterMessageProcessorCalled   func(topic string, identifier string, handler p2p.MessageProcessor) error
	BootstrapCalled                  func() error
	PeerAddressesCalled              func(pid core.PeerID) []string
	BroadcastOnChannelBlockingCalled func(channel string, topic string, buff []byte) error
	IsConnectedToTheNetworkCalled    func() bool
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

// BroadcastOnChannel -
func (ms *MessengerStub) BroadcastOnChannel(channel string, topic string, buff []byte) {
	ms.BroadcastOnChannelCalled(channel, topic, buff)
}

// Bootstrap -
func (ms *MessengerStub) Bootstrap(_ uint32) error {
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
	return ms.IsConnectedToTheNetworkCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
