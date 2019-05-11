package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// BootstraperMock mocks the implementation for a Bootstraper
type BootstraperMock struct {
	CreateAndCommitEmptyBlockCalled func(uint32) (data.BodyHandler, data.HeaderHandler, error)
	AddSyncStateListenerCalled      func(func(bool))
	ShouldSyncCalled                func() bool
	StartSyncCalled                 func()
	StopSyncCalled                  func()
}

func (boot *BootstraperMock) CreateAndCommitEmptyBlock(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
	if boot.CreateAndCommitEmptyBlockCalled != nil {
		return boot.CreateAndCommitEmptyBlockCalled(shardForCurrentNode)
	}

	bb := make(block.Body, 0)
	return bb, &block.Header{}, nil
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

func (boot *BootstraperMock) StartSync() {
	boot.StartSyncCalled()
}

func (boot *BootstraperMock) StopSync() {
	boot.StopSyncCalled()
}
