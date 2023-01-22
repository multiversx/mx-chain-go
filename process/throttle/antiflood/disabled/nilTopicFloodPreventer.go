package disabled

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.TopicFloodPreventer = (*nilTopicFloodPreventer)(nil)

// nilTopicFloodPreventer is a nil (disabled) implementation of a flood preventer that will not count nor keep track
// of the received messages on a topic
type nilTopicFloodPreventer struct {
}

// NewNilTopicFloodPreventer returns a new instance of nilTopicFloodPreventer
func NewNilTopicFloodPreventer() *nilTopicFloodPreventer {
	return &nilTopicFloodPreventer{}
}

// IncreaseLoad will always return nil
func (ntfp *nilTopicFloodPreventer) IncreaseLoad(_ core.PeerID, _ string, _ uint32) error {
	return nil
}

// ResetForTopic does nothing
func (ntfp *nilTopicFloodPreventer) ResetForTopic(_ string) {
}

// ResetForNotRegisteredTopics does nothing
func (ntfp *nilTopicFloodPreventer) ResetForNotRegisteredTopics() {
}

// SetMaxMessagesForTopic does nothing
func (ntfp *nilTopicFloodPreventer) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (ntfp *nilTopicFloodPreventer) IsInterfaceNil() bool {
	return ntfp == nil
}
