package txcache

type scoreComputer interface {
	computeScore(scoreParams senderScoreParams) uint32
}
