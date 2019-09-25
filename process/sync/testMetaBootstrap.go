package sync

import "github.com/ElrondNetwork/elrond-go/data/block"

// TestMetaBootstrap extends MetaBootstrap and is used in integration tests as it exposes some funcs
// that are not supposed to be used in production code
// Exported funcs simplify the reproduction of edge cases
type TestMetaBootstrap struct {
	*MetaBootstrap
}

// ManualRollback calls the rollback on the current block from the blockchain structure
func (tmb *TestMetaBootstrap) ManualRollback() error {
	return tmb.rollback(tmb.blkc.GetCurrentBlockHeader().(*block.MetaBlock))
}

// SetProbableHighestNonce sets the probable highest nonce in the contained fork detector
func (tmb *TestMetaBootstrap) SetProbableHighestNonce(nonce uint64) {
	tmb.forkDetector.(*metaForkDetector).setProbableHighestNonce(nonce)
}
