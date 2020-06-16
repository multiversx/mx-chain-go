package floodPreventers

import "github.com/ElrondNetwork/elrond-go/core"

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

func (qfp *quotaFloodPreventer) IncreaseLoadOnPid(pid core.PeerID, size uint64, percentReserved float32) error {
	qfp.mutOperation.Lock()
	defer qfp.mutOperation.Unlock()

	return qfp.increaseLoad(pid, size, percentReserved)
}
