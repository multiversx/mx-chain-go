package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// MiniBlocksGetterStub -
type MiniBlocksGetterStub struct {
	GetMiniBlocksCalled         func(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
	GetMiniBlocksFromPoolCalled func(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
}

// GetMiniBlocks -
func (mbgs *MiniBlocksGetterStub) GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	if mbgs.GetMiniBlocksCalled != nil {
		return mbgs.GetMiniBlocksCalled(hashes)
	}
	return nil, nil
}

// GetMiniBlocksFromPool -
func (mbgs *MiniBlocksGetterStub) GetMiniBlocksFromPool(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	if mbgs.GetMiniBlocksFromPoolCalled != nil {
		return mbgs.GetMiniBlocksFromPoolCalled(hashes)
	}
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mbgs *MiniBlocksGetterStub) IsInterfaceNil() bool {
	return mbgs == nil
}
