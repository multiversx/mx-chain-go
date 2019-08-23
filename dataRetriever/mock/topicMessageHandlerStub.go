package mock

type topicMessageHandlerStub struct {
	*TopicHandlerStub
	*MessageHandlerStub
}

func NewTopicMessageHandlerStub() *topicMessageHandlerStub {
	return &topicMessageHandlerStub{
		TopicHandlerStub:   &TopicHandlerStub{},
		MessageHandlerStub: &MessageHandlerStub{},
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *topicMessageHandlerStub) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}
