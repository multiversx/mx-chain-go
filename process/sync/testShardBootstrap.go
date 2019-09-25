package sync

import "github.com/ElrondNetwork/elrond-go/data/block"

// TestShardBootstrap extends ShardBootstrap and is used in integration tests as it exposes some funcs
// that are not supposed to be used in production code
// Exported funcs simplify the reproduction of edge cases
type TestShardBootstrap struct {
	*ShardBootstrap
}

// ManualRollback calls the rollback on the current block from the blockchain structure
func (tsb *TestShardBootstrap) ManualRollback() error {
	return tsb.rollback(tsb.blkc.GetCurrentBlockHeader().(*block.Header))
}

// SetProbableHighestNonce sets the probable highest nonce in the contained fork detector
func (tsb *TestShardBootstrap) SetProbableHighestNonce(nonce uint64) {
	tsb.forkDetector.(*shardForkDetector).setProbableHighestNonce(nonce)
}
