package factory

type disabledTopicFloodPreventer struct {
}

// IncreaseLoad won't do anything and will return nil
func (d *disabledTopicFloodPreventer) IncreaseLoad(_ string, _ string, _ uint32) error {
	return nil
}

// ResetForTopic won't do anything
func (d *disabledTopicFloodPreventer) ResetForTopic(_ string) {
}

// ResetForNotRegisteredTopics won't do anything
func (d *disabledTopicFloodPreventer) ResetForNotRegisteredTopics() {
}

// SetMaxMessagesForTopic won't do anything
func (d *disabledTopicFloodPreventer) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledTopicFloodPreventer) IsInterfaceNil() bool {
	return d == nil
}
