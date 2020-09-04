package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// BootstrapperMock mocks the implementation for a Bootstrapper
type BootstrapperMock struct {
	CreateAndCommitEmptyBlockCalled func(uint32) (data.BodyHandler, data.HeaderHandler, error)
	AddSyncStateListenerCalled      func(func(bool))
	GetNodeStateCalled              func() core.NodeState
	StartSyncingBlocksCalled        func()
}

// CreateAndCommitEmptyBlock -
func (boot *BootstrapperMock) CreateAndCommitEmptyBlock(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
	if boot.CreateAndCommitEmptyBlockCalled != nil {
		return boot.CreateAndCommitEmptyBlockCalled(shardForCurrentNode)
	}

	return &block.Body{}, &block.Header{}, nil
}

// AddSyncStateListener -
func (boot *BootstrapperMock) AddSyncStateListener(syncStateNotifier func(isSyncing bool)) {
	if boot.AddSyncStateListenerCalled != nil {
		boot.AddSyncStateListenerCalled(syncStateNotifier)
	}
}

// GetNodeState -
func (boot *BootstrapperMock) GetNodeState() core.NodeState {
	if boot.GetNodeStateCalled != nil {
		return boot.GetNodeStateCalled()
	}

	return core.NsSynchronized
}

// StartSyncingBlocks -
func (boot *BootstrapperMock) StartSyncingBlocks() {
	boot.StartSyncingBlocksCalled()
}

// Close -
func (boot *BootstrapperMock) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (boot *BootstrapperMock) IsInterfaceNil() bool {
	return boot == nil
}
