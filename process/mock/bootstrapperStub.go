package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
)

// BootstrapperStub mocks the implementation for a Bootstrapper
type BootstrapperStub struct {
	CreateAndCommitEmptyBlockCalled func(uint32) (data.BodyHandler, data.HeaderHandler, error)
	AddSyncStateListenerCalled      func(func(bool))
	GetNodeStateCalled              func() common.NodeState
	StartSyncingBlocksCalled        func() error
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
func (boot *BootstrapperStub) GetNodeState() common.NodeState {
	if boot.GetNodeStateCalled != nil {
		return boot.GetNodeStateCalled()
	}

	return common.NsSynchronized
}

// StartSyncingBlocks -
func (boot *BootstrapperStub) StartSyncingBlocks() error {
	if boot.StartSyncingBlocksCalled != nil {
		return boot.StartSyncingBlocksCalled()
	}

	return nil
}

// Close -
func (boot *BootstrapperStub) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (boot *BootstrapperStub) IsInterfaceNil() bool {
	return boot == nil
}
