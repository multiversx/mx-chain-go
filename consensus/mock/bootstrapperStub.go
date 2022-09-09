package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// BootstrapperStub mocks the implementation for a Bootstrapper
type BootstrapperStub struct {
	CreateAndCommitEmptyBlockCalled func(uint32) (data.BodyHandler, data.HeaderHandler, error)
	AddSyncStateListenerCalled      func(func(bool))
	GetNodeStateCalled              func() core.NodeState
	StartSyncingBlocksCalled        func()
}

// CreateAndCommitEmptyBlock -
func (boot *BootstrapperStub) CreateAndCommitEmptyBlock(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
	if boot.CreateAndCommitEmptyBlockCalled != nil {
		return boot.CreateAndCommitEmptyBlockCalled(shardForCurrentNode)
	}

	return &block.Body{}, &block.Header{}, nil
}

// AddSyncStateListener -
func (boot *BootstrapperStub) AddSyncStateListener(syncStateNotifier func(isSyncing bool)) {
	if boot.AddSyncStateListenerCalled != nil {
		boot.AddSyncStateListenerCalled(syncStateNotifier)
	}
}

// GetNodeState -
func (boot *BootstrapperStub) GetNodeState() core.NodeState {
	if boot.GetNodeStateCalled != nil {
		return boot.GetNodeStateCalled()
	}

	return core.NsSynchronized
}

// StartSyncingBlocks -
func (boot *BootstrapperStub) StartSyncingBlocks() {
	boot.StartSyncingBlocksCalled()
}

// Close -
func (boot *BootstrapperStub) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (boot *BootstrapperStub) IsInterfaceNil() bool {
	return boot == nil
}
