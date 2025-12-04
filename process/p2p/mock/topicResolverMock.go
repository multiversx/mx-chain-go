package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
)

// TopicResolverMock -
type TopicResolverMock struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) ([]byte, error)
	RequestTopicCalled           func() string
	SetDebugHandlerCalled        func(handler dataRetriever.DebugHandler) error
	CloseCalled                  func() error
}

// SetDebugHandler -
func (mock *TopicResolverMock) SetDebugHandler(handler dataRetriever.DebugHandler) error {
	if mock.SetDebugHandlerCalled != nil {
		return mock.SetDebugHandlerCalled(handler)
	}

	return nil
}

// ProcessReceivedMessage -
func (mock *TopicResolverMock) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID, source p2p.MessageHandler) ([]byte, error) {
	if mock.ProcessReceivedMessageCalled != nil {
		return mock.ProcessReceivedMessageCalled(message, fromConnectedPeer, source)
	}

	return make([]byte, 0), nil
}

// RequestTopic -
func (mock *TopicResolverMock) RequestTopic() string {
	if mock.RequestTopicCalled != nil {
		return mock.RequestTopicCalled()
	}

	return ""
}

// Close -
func (mock *TopicResolverMock) Close() error {
	if mock.CloseCalled != nil {
		return mock.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (mock *TopicResolverMock) IsInterfaceNil() bool {
	return mock == nil
}
