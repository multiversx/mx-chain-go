package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// BootstraperMock mocks the implementation for a Bootstraper
type BootstraperMock struct {
	ShouldSyncCalled       func() bool
	CreateEmptyBlockCalled func(uint32) (*block.TxBlockBody, *block.Header)
}

func (boot *BootstraperMock) ShouldSync() bool {
	return boot.ShouldSyncCalled()
}

func (boot *BootstraperMock) CreateEmptyBlock(shardForCurrentNode uint32) (*block.TxBlockBody, *block.Header) {
	if boot.CreateEmptyBlockCalled != nil {
		return boot.CreateEmptyBlockCalled(shardForCurrentNode)
	}

	return &block.TxBlockBody{}, &block.Header{}
}
