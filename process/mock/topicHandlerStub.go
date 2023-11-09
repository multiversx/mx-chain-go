package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
)

// TopicHandlerStub -
type TopicHandlerStub struct {
	HasTopicCalled                 func(name string) bool
	CreateTopicCalled              func(name string, createChannelForTopic bool) error
	RegisterMessageProcessorCalled func(topic string, identifier string, handler p2p.MessageProcessor) error
	IDCalled                       func() core.PeerID
}

// HasTopic -
func (stub *TopicHandlerStub) HasTopic(name string) bool {
	if stub.HasTopicCalled != nil {
		return stub.HasTopicCalled(name)
	}
	return false
}

// CreateTopic -
func (stub *TopicHandlerStub) CreateTopic(name string, createChannelForTopic bool) error {
	if stub.CreateTopicCalled != nil {
		return stub.CreateTopicCalled(name, createChannelForTopic)
	}
	return nil
}

// RegisterMessageProcessor -
func (stub *TopicHandlerStub) RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error {
	if stub.RegisterMessageProcessorCalled != nil {
		return stub.RegisterMessageProcessorCalled(topic, identifier, handler)
	}
	return nil
}

// ID -
func (stub *TopicHandlerStub) ID() core.PeerID {
	if stub.IDCalled != nil {
		return stub.IDCalled()
	}

	return "peer ID"
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *TopicHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
