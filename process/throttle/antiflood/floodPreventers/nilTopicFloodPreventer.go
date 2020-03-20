package floodPreventers

// nilTopicFloodPreventer is a nil (disabled) implementation of a flood preventer that will not count nor keep track
// of the received messages on a topic
type nilTopicFloodPreventer struct {
}

// NewNilTopicFloodPreventer returns a new instance of nilTopicFloodPreventer
func NewNilTopicFloodPreventer() *nilTopicFloodPreventer {
	return &nilTopicFloodPreventer{}
}

// Accumulate will always return true
func (ntfp *nilTopicFloodPreventer) Accumulate(_ string, _ string) bool {
	return true
}

// ResetForTopic does nothing
func (ntfp *nilTopicFloodPreventer) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic does nothing
func (ntfp *nilTopicFloodPreventer) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (ntfp *nilTopicFloodPreventer) IsInterfaceNil() bool {
	return ntfp == nil
}
