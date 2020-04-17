package mock

// TopicAntiFloodStub -
type TopicAntiFloodStub struct {
	IncreaseLoadCalled func(identifier string, topic string, numMessages uint32) error
}

// IncreaseLoad -
func (t *TopicAntiFloodStub) IncreaseLoad(identifier string, topic string, numMessages uint32) error {
	if t.IncreaseLoadCalled != nil {
		return t.IncreaseLoadCalled(identifier, topic, numMessages)
	}

	return nil
}

// ResetForTopic -
func (t *TopicAntiFloodStub) ResetForTopic(_ string) {
}

// ResetForNotRegisteredTopics -
func (t *TopicAntiFloodStub) ResetForNotRegisteredTopics() {
}

// SetMaxMessagesForTopic -
func (t *TopicAntiFloodStub) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// IsInterfaceNil -
func (t *TopicAntiFloodStub) IsInterfaceNil() bool {
	return t == nil
}
