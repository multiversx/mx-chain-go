package mock

// TopicAntiFloodStub -
type TopicAntiFloodStub struct {
	IncreaseLoadCalled           func(identifier string, topic string, numMessages uint32) error
	ResetForTopicCalled          func(topic string)
	SetMaxMessagesForTopicCalled func(topic string, num uint32)
}

// IncreaseLoad -
func (t *TopicAntiFloodStub) IncreaseLoad(identifier string, topic string, numMessages uint32) error {
	if t.IncreaseLoadCalled != nil {
		return t.IncreaseLoadCalled(identifier, topic, numMessages)
	}

	return nil
}

// ResetForTopic -
func (t *TopicAntiFloodStub) ResetForTopic(topic string) {
	if t.ResetForTopicCalled != nil {
		t.ResetForTopicCalled(topic)
	}
}

// ResetForNotRegisteredTopics -
func (t *TopicAntiFloodStub) ResetForNotRegisteredTopics() {
}

// SetMaxMessagesForTopic -
func (t *TopicAntiFloodStub) SetMaxMessagesForTopic(topic string, num uint32) {
	if t.SetMaxMessagesForTopicCalled != nil {
		t.SetMaxMessagesForTopicCalled(topic, num)
	}
}

// IsInterfaceNil -
func (t *TopicAntiFloodStub) IsInterfaceNil() bool {
	return t == nil
}
