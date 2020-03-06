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
	ShouldSyncCalled                func() bool
	StartSyncCalled                 func()
	StopSyncCalled                  func()
	SetStatusHandlerCalled          func(handler core.AppStatusHandler) error
}

// CreateAndCommitEmptyBlock -
func (boot *BootstrapperMock) CreateAndCommitEmptyBlock(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
	if boot.CreateAndCommitEmptyBlockCalled != nil {
		return boot.CreateAndCommitEmptyBlockCalled(shardForCurrentNode)
	}

	bb := make(block.Body, 0)
	return bb, &block.Header{}, nil
}

// AddSyncStateListener -
func (boot *BootstrapperMock) AddSyncStateListener(syncStateNotifier func(isSyncing bool)) {
	if boot.AddSyncStateListenerCalled != nil {
		boot.AddSyncStateListenerCalled(syncStateNotifier)
	}
}

// ShouldSync -
func (boot *BootstrapperMock) ShouldSync() bool {
	if boot.ShouldSyncCalled != nil {
		return boot.ShouldSyncCalled()
	}

	return false
}

// StartSync -
func (boot *BootstrapperMock) StartSync() {
	boot.StartSyncCalled()
}

// StopSync -
func (boot *BootstrapperMock) StopSync() {
	boot.StopSyncCalled()
}

// SetStatusHandler -
func (boot *BootstrapperMock) SetStatusHandler(handler core.AppStatusHandler) error {
	return boot.SetStatusHandlerCalled(handler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (boot *BootstrapperMock) IsInterfaceNil() bool {
	return boot == nil
}
