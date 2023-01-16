package mock

import (
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// MiniBlocksProviderStub -
type MiniBlocksProviderStub struct {
	GetMiniBlocksCalled           func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
	GetMiniBlocksFromPoolCalled   func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
	GetMiniBlocksFromStorerCalled func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
}

// GetMiniBlocks -
func (mbps *MiniBlocksProviderStub) GetMiniBlocks(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
	if mbps.GetMiniBlocksCalled != nil {
		return mbps.GetMiniBlocksCalled(hashes)
	}
	return nil, nil
}

// GetMiniBlocksFromPool -
func (mbps *MiniBlocksProviderStub) GetMiniBlocksFromPool(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
	if mbps.GetMiniBlocksFromPoolCalled != nil {
		return mbps.GetMiniBlocksFromPoolCalled(hashes)
	}
	return nil, nil
}

// GetMiniBlocksFromStorer -
func (mbps *MiniBlocksProviderStub) GetMiniBlocksFromStorer(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
	if mbps.GetMiniBlocksFromStorerCalled != nil {
		return mbps.GetMiniBlocksFromStorerCalled(hashes)
	}
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mbps *MiniBlocksProviderStub) IsInterfaceNil() bool {
	return mbps == nil
}
