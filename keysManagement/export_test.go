package keysManagement

func (pInfo *peerInfo) getRoundsWithoutReceivedMessages() int {
	pInfo.mutChangeableData.RLock()
	defer pInfo.mutChangeableData.RUnlock()

	return pInfo.roundsWithoutReceivedMessages
}
