package queue

// SetLastAddedNonce will set the last added nonce
func (bq *blocksQueue) SetLastAddedNonce(lastAddedNonce uint64) {
	bq.mutex.Lock()
	defer bq.mutex.Unlock()

	bq.lastAddedNonce = lastAddedNonce
}
