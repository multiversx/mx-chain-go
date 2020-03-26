package floodPreventers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/process"
)

const topicMinMessages = 1

// WildcardCharacter is the character string used to specify that the topic refers to a
const WildcardCharacter = "*"

var log = logger.GetOrCreate("process/throttle/antiflood")

// topicFloodPreventer represents a flood preventer based on limitations of messages per given topics
type topicFloodPreventer struct {
	mutTopicMaxMessages       *sync.RWMutex
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
		mutTopicMaxMessages:       &sync.RWMutex{},
		topicMaxMessages:          make(map[string]uint32),
		counterMap:                make(map[string]map[string]uint32),
		defaultMaxMessagesPerPeer: maxMessagesPerPeer,
	}, nil
}

// Accumulate tries to increment the counter values held at "identifier" position for the given topic
// It returns true if it had succeeded incrementing (existing counter value is lower than provided maxMessagesPerPeer)
func (tfp *topicFloodPreventer) Accumulate(identifier string, topic string) bool {
	tfp.mutTopicMaxMessages.Lock()
	defer tfp.mutTopicMaxMessages.Unlock()

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

	limitExceeded := tfp.counterMap[topic][identifier] > tfp.maxMessagesForTopic(topic)
	return !limitExceeded
}

// SetMaxMessagesForTopic will update the maximum number of messages that can be received from a peer in a topic
func (tfp *topicFloodPreventer) SetMaxMessagesForTopic(topic string, numMessages uint32) {
	log.Debug("SetMaxMessagesForTopic", "topic", topic, "num messages", numMessages)
	tfp.mutTopicMaxMessages.Lock()
	tfp.topicMaxMessages[topic] = numMessages
	tfp.mutTopicMaxMessages.Unlock()
}

// ResetForTopic clears all map values for a given topic
func (tfp *topicFloodPreventer) ResetForTopic(topic string) {
	tfp.mutTopicMaxMessages.Lock()
	defer tfp.mutTopicMaxMessages.Unlock()

	if strings.Contains(topic, WildcardCharacter) {
		tfp.resetTopicWithWildCard(topic)
	}
	tfp.counterMap[topic] = make(map[string]uint32)
}

func (tfp *topicFloodPreventer) resetTopicWithWildCard(topic string) {
	topicWithoutWildcard := strings.Replace(topic, WildcardCharacter, "", 1)
	for topicKey := range tfp.counterMap {
		if strings.Contains(topicKey, topicWithoutWildcard) {
			tfp.counterMap[topicKey] = make(map[string]uint32)
		}
	}
}

func (tfp *topicFloodPreventer) maxMessagesForTopic(topic string) uint32 {
	maxMessages, ok := tfp.topicMaxMessages[topic]
	if !ok {
		maxMessages = tfp.maxMessagesForTopicWildcard(topic)
		tfp.topicMaxMessages[topic] = maxMessages
	}

	return maxMessages
}

func (tfp *topicFloodPreventer) maxMessagesForTopicWildcard(topic string) uint32 {
	for t, maxMessages := range tfp.topicMaxMessages {
		if !strings.Contains(t, WildcardCharacter) {
			continue
		}

		topicWithoutWildcard := strings.Replace(t, WildcardCharacter, "", 1)
		if strings.Contains(topic, topicWithoutWildcard) {
			return maxMessages
		}
	}

	return tfp.defaultMaxMessagesPerPeer
}

// IsInterfaceNil returns true if there is no value under the interface
func (tfp *topicFloodPreventer) IsInterfaceNil() bool {
	return tfp == nil
}
