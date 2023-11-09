package floodPreventers

import "github.com/multiversx/mx-chain-core-go/core"

func (tfp *topicFloodPreventer) CountForTopicAndIdentifier(topic string, pid core.PeerID) uint32 {
	tfp.mutTopicMaxMessages.RLock()
	defer tfp.mutTopicMaxMessages.RUnlock()

	mapForTopic, ok := tfp.counterMap[topic]
	if !ok {
		return 0
	}
	countForId, ok := mapForTopic[pid]
	if !ok {
		return 0
	}
	return countForId
}

func (tfp *topicFloodPreventer) MaxMessagesForTopic(topic string) uint32 {
	return tfp.maxMessagesForTopic(topic)
}

func (tfp *topicFloodPreventer) TopicMaxMessages() map[string]uint32 {
	copiedMaxMessages := make(map[string]uint32)
	tfp.mutTopicMaxMessages.RLock()
	for topic, val := range tfp.topicMaxMessages {
		copiedMaxMessages[topic] = val
	}
	tfp.mutTopicMaxMessages.RUnlock()

	return copiedMaxMessages
}
