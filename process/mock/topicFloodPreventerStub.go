package mock

// TopicAntiFloodStub -
type TopicAntiFloodStub struct {
	AccumulateCalled func(identifier string, topic string, numMessages uint32) error
}

// Accumulate -
func (t *TopicAntiFloodStub) Accumulate(identifier string, topic string, numMessages uint32) error {
	if t.AccumulateCalled != nil {
		return t.AccumulateCalled(identifier, topic, numMessages)
	}

	return nil
}

// ResetForTopic -
func (t *TopicAntiFloodStub) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic -
func (t *TopicAntiFloodStub) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// IsInterfaceNil -
func (t *TopicAntiFloodStub) IsInterfaceNil() bool {
	return t == nil
}
