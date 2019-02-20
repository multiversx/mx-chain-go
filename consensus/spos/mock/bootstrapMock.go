package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// BootstraperMock mocks the implementation for a Bootstraper
type BootstraperMock struct {
	CreateAndCommitEmptyBlockCalled func(uint32) (*block.TxBlockBody, *block.Header)
	AddSyncStateListnerCalled       func(func(bool))
	ShouldSyncCalled                func() bool
}

func (boot *BootstraperMock) CreateAndCommitEmptyBlock(shardForCurrentNode uint32) (*block.TxBlockBody, *block.Header) {
	if boot.CreateAndCommitEmptyBlockCalled != nil {
		return boot.CreateAndCommitEmptyBlockCalled(shardForCurrentNode)
	}

	return &block.TxBlockBody{}, &block.Header{}
}

func (boot *BootstraperMock) AddSyncStateListner(syncStateNotifier func(bool)) {
	if boot.AddSyncStateListnerCalled != nil {
		boot.AddSyncStateListnerCalled(syncStateNotifier)
		return
	}

	return
}

func (boot *BootstraperMock) ShouldSync() bool {
	if boot.ShouldSyncCalled != nil {
		return boot.ShouldSyncCalled()
	}

	return false
}
