package sync

// TestShardBootstrap extends ShardBootstrap and is used in integration tests as it exposes some funcs
// that are not supposed to be used in production code
// Exported funcs simplify the reproduction of edge cases
type TestShardBootstrap struct {
	*ShardBootstrap
}

// RollBack decides to call (or not) the rollBackOneBlock on the current block from the blockchain structure
func (tsb *TestShardBootstrap) RollBack(revertUsingForkNonce bool) error {
	return tsb.rollBack(revertUsingForkNonce)
}

// SetProbableHighestNonce sets the probable highest nonce in the contained fork detector
func (tsb *TestShardBootstrap) SetProbableHighestNonce(nonce uint64) {
	forkDetector, ok := tsb.forkDetector.(*shardForkDetector)
	if !ok {
		log.Error("inner forkdetector impl is not of type shardForkDetector")
		return
	}

	forkDetector.setProbableHighestNonce(nonce)
}
