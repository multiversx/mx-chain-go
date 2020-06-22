package txcache

type scoreComputer interface {
	computeScore(scoreParams senderScoreParams) uint32
}

// ForEachTransaction is an iterator callback
type ForEachTransaction func(txHash []byte, value *WrappedTransaction)
