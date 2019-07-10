package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// BootstrapperMock mocks the implementation for a Bootstrapper
type BootstrapperMock struct {
	CreateAndCommitEmptyBlockCalled func(uint32) (data.BodyHandler, data.HeaderHandler, error)
	AddSyncStateListenerCalled      func(func(bool))
	ShouldSyncCalled                func() bool
	StartSyncCalled                 func()
	StopSyncCalled                  func()
}

func (boot *BootstrapperMock) CreateAndCommitEmptyBlock(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
	if boot.CreateAndCommitEmptyBlockCalled != nil {
		return boot.CreateAndCommitEmptyBlockCalled(shardForCurrentNode)
	}

	bb := make(block.Body, 0)
	return bb, &block.Header{}, nil
}

func (boot *BootstrapperMock) AddSyncStateListener(syncStateNotifier func(bool)) {
	if boot.AddSyncStateListenerCalled != nil {
		boot.AddSyncStateListenerCalled(syncStateNotifier)
		return
	}

	return
}

func (boot *BootstrapperMock) ShouldSync() bool {
	if boot.ShouldSyncCalled != nil {
		return boot.ShouldSyncCalled()
	}

	return false
}

func (boot *BootstrapperMock) StartSync() {
	boot.StartSyncCalled()
}

func (boot *BootstrapperMock) StopSync() {
	boot.StopSyncCalled()
}
