package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// MiniBlocksCompacterMock -
type MiniBlocksCompacterMock struct {
	CompactCalled func(block.MiniBlockSlice, map[string]data.TransactionHandler) block.MiniBlockSlice
	ExpandCalled  func(block.MiniBlockSlice, map[string]data.TransactionHandler) (block.MiniBlockSlice, error)
}

// Compact -
func (mbcm *MiniBlocksCompacterMock) Compact(miniBlocks block.MiniBlockSlice, mapHashesAndTxs map[string]data.TransactionHandler) block.MiniBlockSlice {
	return mbcm.CompactCalled(miniBlocks, mapHashesAndTxs)
}

// Expand -
func (mbcm *MiniBlocksCompacterMock) Expand(miniBlocks block.MiniBlockSlice, mapHashesAndTxs map[string]data.TransactionHandler) (block.MiniBlockSlice, error) {
	return mbcm.ExpandCalled(miniBlocks, mapHashesAndTxs)
}

// IsInterfaceNil -
func (mbcm *MiniBlocksCompacterMock) IsInterfaceNil() bool {
	return mbcm == nil
}
