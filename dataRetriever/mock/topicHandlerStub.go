package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type TopicHandlerStub struct {
	HasTopicCalled                 func(name string) bool
	CreateTopicCalled              func(name string, createChannelForTopic bool) error
	RegisterMessageProcessorCalled func(topic string, handler p2p.MessageProcessor) error
}

func (ths *TopicHandlerStub) HasTopic(name string) bool {
	return ths.HasTopicCalled(name)
}

func (ths *TopicHandlerStub) CreateTopic(name string, createChannelForTopic bool) error {
	return ths.CreateTopicCalled(name, createChannelForTopic)
}

func (ths *TopicHandlerStub) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	return ths.RegisterMessageProcessorCalled(topic, handler)
}
