package mock

type TopicAntiFloodStub struct {
	AccumulateCalled func(identifier string, topic string) bool
}

func (t *TopicAntiFloodStub) Accumulate(identifier string, topic string) bool {
	if t.AccumulateCalled != nil {
		return t.AccumulateCalled(identifier, topic)
	}

	return true
}

func (t *TopicAntiFloodStub) ResetForTopic(topic string) {
}

func (t *TopicAntiFloodStub) SetMaxMessagesForTopic(topic string, maxNum uint32) {
}

func (t *TopicAntiFloodStub) IsInterfaceNil() bool {
	return t == nil
}
