package floodPreventers

func (qfp *quotaFloodPreventer) SetGlobalQuotaValues(maxMessages uint32, size uint64) {
	qfp.mutOperation.Lock()
	qfp.globalQuota.numReceivedMessages = maxMessages
	qfp.globalQuota.sizeReceivedMessages = size
	qfp.mutOperation.Unlock()
}

func (tfp *topicFloodPreventer) CountForTopicAndIdentifier(topic string, identifier string) uint32 {
	tfp.mutTopicMaxMessages.RLock()
	defer tfp.mutTopicMaxMessages.RUnlock()

	mapForTopic, ok := tfp.counterMap[topic]
	if !ok {
		return 0
	}
	countForId, ok := mapForTopic[identifier]
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
