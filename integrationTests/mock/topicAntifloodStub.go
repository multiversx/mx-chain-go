package mock

import "github.com/multiversx/mx-chain-core-go/core"

// TopicAntiFloodStub -
type TopicAntiFloodStub struct {
	IncreaseLoadCalled func(pid core.PeerID, topic string, numMessages uint32) error
}

// IncreaseLoad -
func (t *TopicAntiFloodStub) IncreaseLoad(pid core.PeerID, topic string, numMessages uint32) error {
	if t.IncreaseLoadCalled != nil {
		return t.IncreaseLoadCalled(pid, topic, numMessages)
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
