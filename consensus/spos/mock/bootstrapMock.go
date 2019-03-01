package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// BootstraperMock mocks the implementation for a Bootstraper
type BootstraperMock struct {
	CreateAndCommitEmptyBlockCalled func(uint32) (block.BlockBody, *block.Header)
	AddSyncStateListenerCalled      func(func(bool))
	ShouldSyncCalled                func() bool
}

func (boot *BootstraperMock) CreateAndCommitEmptyBlock(shardForCurrentNode uint32) (block.BlockBody, *block.Header) {
	if boot.CreateAndCommitEmptyBlockCalled != nil {
		return boot.CreateAndCommitEmptyBlockCalled(shardForCurrentNode)
	}

	bb := make(block.BlockBody, 0)
	return bb, &block.Header{}
}

func (boot *BootstraperMock) AddSyncStateListener(syncStateNotifier func(bool)) {
	if boot.AddSyncStateListenerCalled != nil {
		boot.AddSyncStateListenerCalled(syncStateNotifier)
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
