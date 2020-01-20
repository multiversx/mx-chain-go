package txcache

// GetKey return the key
func (listForSender *txListForSender) GetKey() string {
	return listForSender.sender
}

// ComputeScore computes the score of the sender, as an integer 0-100
func (listForSender *txListForSender) ComputeScore() uint32 {
	listForSender.lastComputedScore = uint32(0)
	return 0
}

// GetScoreChunk returns the score chunk the sender is currently in
func (listForSender *txListForSender) GetScoreChunk() *MapChunk {
	return listForSender.scoreChunk
}

// GetScoreChunk returns the score chunk the sender is currently in
func (listForSender *txListForSender) SetScoreChunk(scoreChunk *MapChunk) {
	listForSender.scoreChunk = scoreChunk
}
