package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type WireTopicHandlerStub struct {
	HasTopicCalled                 func(name string) bool
	CreateTopicCalled              func(name string, createChannelForTopic bool) error
	RegisterMessageProcessorCalled func(topic string, handler p2p.MessageProcessor) error
}

func (wths *WireTopicHandlerStub) HasTopic(name string) bool {
	return wths.HasTopicCalled(name)
}

func (wths *WireTopicHandlerStub) CreateTopic(name string, createChannelForTopic bool) error {
	return wths.CreateTopicCalled(name, createChannelForTopic)
}

func (wths *WireTopicHandlerStub) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	return wths.RegisterMessageProcessorCalled(topic, handler)
}
