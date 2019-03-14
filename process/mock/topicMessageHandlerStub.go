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
