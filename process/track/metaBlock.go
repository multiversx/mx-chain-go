package track

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

// metaBlock implements NotarisedBlocksTracker interface which tracks notarised blocks
type metaBlock struct {
}

// NewMetaBlock creates a new metaBlock object
func NewMetaBlock() (*metaBlock, error) {
	mb := metaBlock{}
	return &mb, nil
}

// UnnotarisedBlocks gets all the blocks which are not notarised yet
func (mb *metaBlock) UnnotarisedBlocks() []data.HeaderHandler {
	return make([]data.HeaderHandler, 0)
}

// RemoveNotarisedBlocks removes all the blocks which already have been notarised
func (mb *metaBlock) RemoveNotarisedBlocks(headerHandler data.HeaderHandler) {
}

// AddBlock adds new block to be tracked
func (mb *metaBlock) AddBlock(headerHandler data.HeaderHandler) {
}

// SetBlockBroadcastRound sets the round in which the block with the given nonce has been broadcast
func (mb *metaBlock) SetBlockBroadcastRound(nonce uint64, round int32) {
}

// BlockBroadcastRound gets the round in which the block with given nonce has been broadcast
func (mb *metaBlock) BlockBroadcastRound(nonce uint64) int32 {
	return 0
}
