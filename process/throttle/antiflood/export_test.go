package antiflood

func (qfp *quotaFloodPreventer) SetGlobalQuotaValues(maxMessages uint32, size uint64) {
	qfp.mutOperation.Lock()
	qfp.globalQuota.numReceivedMessages = maxMessages
	qfp.globalQuota.sizeReceivedMessages = size
	qfp.mutOperation.Unlock()
}

func (tfp *topicFloodPreventer) CountForTopicAndIdentifier(topic string, identifier string) uint32 {
	tfp.mutOperation.RLock()
	defer tfp.mutOperation.RUnlock()

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
