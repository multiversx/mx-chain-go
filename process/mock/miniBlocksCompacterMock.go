package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type MiniBlocksCompacterMock struct {
	CompactCalled func(block.MiniBlockSlice, map[string]data.TransactionHandler) block.MiniBlockSlice
	ExpandCalled  func(block.MiniBlockSlice, map[string]data.TransactionHandler) (block.MiniBlockSlice, error)
}

func (mbcm *MiniBlocksCompacterMock) Compact(miniBlocks block.MiniBlockSlice, mapHashesAndTxs map[string]data.TransactionHandler) block.MiniBlockSlice {
	return mbcm.CompactCalled(miniBlocks, mapHashesAndTxs)
}

func (mbcm *MiniBlocksCompacterMock) Expand(miniBlocks block.MiniBlockSlice, mapHashesAndTxs map[string]data.TransactionHandler) (block.MiniBlockSlice, error) {
	return mbcm.ExpandCalled(miniBlocks, mapHashesAndTxs)
}

func (mbcm *MiniBlocksCompacterMock) IsInterfaceNil() bool {
	if mbcm == nil {
		return true
	}
	return false
}
