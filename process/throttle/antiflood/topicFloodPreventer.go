package antiflood

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
)

const topicMinMessages = 1

// topicFloodPreventer represents a cache of quotas per peer used in antiflooding mechanism
type topicFloodPreventer struct {
	mutOperation              *sync.RWMutex
	topicMaxMessages          map[string]uint32
	counterMap                map[string]map[string]uint32
	defaultMaxMessagesPerPeer uint32
}

// NewTopicFloodPreventer creates a new flood preventer based on topic
func NewTopicFloodPreventer(
	maxMessagesPerPeer uint32,
) (*topicFloodPreventer, error) {

	if maxMessagesPerPeer < topicMinMessages {
		return nil, fmt.Errorf("%w raised in NewTopicFloodPreventer, maxMessagesPerPeer: provided %d, minimum %d",
			process.ErrInvalidValue,
			maxMessagesPerPeer,
			topicMinMessages,
		)
	}

	return &topicFloodPreventer{
		mutOperation:              &sync.RWMutex{},
		topicMaxMessages:          make(map[string]uint32),
		counterMap:                make(map[string]map[string]uint32),
		defaultMaxMessagesPerPeer: maxMessagesPerPeer,
	}, nil
}

// Accumulate tries to increment the counter values held at "identifier" position for the given topic
// It returns true if it had succeeded incrementing (existing counter value is lower than provided maxMessagesPerPeer)
func (tfp *topicFloodPreventer) Accumulate(identifier string, topic string) bool {
	tfp.mutOperation.Lock()
	defer tfp.mutOperation.Unlock()

	_, ok := tfp.counterMap[topic]
	if !ok {
		tfp.counterMap[topic] = make(map[string]uint32)
	}

	_, ok = tfp.counterMap[topic][identifier]
	if !ok {
		tfp.counterMap[topic][identifier] = 1
		return true
	}

	// if this was already in the map, just increment it
	tfp.counterMap[topic][identifier]++

	if tfp.counterMap[topic][identifier] > tfp.maxMessagesForTopic(topic) {
		return false
	}

	return true
}

// SetMaxMessagesForTopic will update the maximum number of messages that can be received from a peer in a topic
func (tfp *topicFloodPreventer) SetMaxMessagesForTopic(topic string, numMessages uint32) {
	// TODO : call this
	tfp.mutOperation.Lock()
	tfp.topicMaxMessages[topic] = numMessages
	tfp.mutOperation.Unlock()
}

// ResetForTopic clears all map values for a given topic
func (tfp *topicFloodPreventer) ResetForTopic(topic string) {
	// TODO : call this
	tfp.mutOperation.Lock()
	defer tfp.mutOperation.Unlock()

	tfp.counterMap[topic] = make(map[string]uint32)
}

func (tfp *topicFloodPreventer) maxMessagesForTopic(topic string) uint32 {
	maxMessages, ok := tfp.topicMaxMessages[topic]
	if !ok {
		tfp.topicMaxMessages[topic] = tfp.defaultMaxMessagesPerPeer
		return tfp.defaultMaxMessagesPerPeer
	}

	return maxMessages
}

// IsInterfaceNil returns true if there is no value under the interface
func (tfp *topicFloodPreventer) IsInterfaceNil() bool {
	return tfp == nil
}
