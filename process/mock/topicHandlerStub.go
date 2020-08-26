package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// TopicHandlerStub -
type TopicHandlerStub struct {
	HasTopicCalled                 func(name string) bool
	CreateTopicCalled              func(name string, createChannelForTopic bool) error
	RegisterMessageProcessorCalled func(topic string, handler p2p.MessageProcessor) error
	IDCalled                       func() core.PeerID
}

// HasTopic -
func (ths *TopicHandlerStub) HasTopic(name string) bool {
	return ths.HasTopicCalled(name)
}

// CreateTopic -
func (ths *TopicHandlerStub) CreateTopic(name string, createChannelForTopic bool) error {
	return ths.CreateTopicCalled(name, createChannelForTopic)
}

// RegisterMessageProcessor -
func (ths *TopicHandlerStub) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	return ths.RegisterMessageProcessorCalled(topic, handler)
}

// ID -
func (ths *TopicHandlerStub) ID() core.PeerID {
	if ths.IDCalled != nil {
		return ths.IDCalled()
	}

	return "peer ID"
}

// IsInterfaceNil returns true if there is no value under the interface
func (ths *TopicHandlerStub) IsInterfaceNil() bool {
	return ths == nil
}
