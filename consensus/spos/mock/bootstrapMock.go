package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// BootstraperMock mocks the implementation for a Bootstraper
type BootstraperMock struct {
	CreateAndCommitEmptyBlockCalled func(uint32) (*block.TxBlockBody, *block.Header, error)
	AddSyncStateListenerCalled      func(func(bool))
	ShouldSyncCalled                func() bool
}

func (boot *BootstraperMock) CreateAndCommitEmptyBlock(shardForCurrentNode uint32) (*block.TxBlockBody, *block.Header, error) {
	if boot.CreateAndCommitEmptyBlockCalled != nil {
		return boot.CreateAndCommitEmptyBlockCalled(shardForCurrentNode)
	}

	return &block.TxBlockBody{}, &block.Header{}, nil
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
