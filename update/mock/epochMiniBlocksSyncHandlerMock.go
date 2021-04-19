package mock

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// EpochStartPendingMiniBlocksSyncHandlerMock -
type EpochStartPendingMiniBlocksSyncHandlerMock struct {
	SyncPendingMiniBlocksFromMetaCalled func(epochStart data.MetaHeaderHandler, unFinished map[string]data.MetaHeaderHandler, ctx context.Context) error
	GetMiniBlocksCalled                 func() (map[string]*block.MiniBlock, error)
}

// SyncPendingMiniBlocksFromMeta -
func (ep *EpochStartPendingMiniBlocksSyncHandlerMock) SyncPendingMiniBlocksFromMeta(epochStart data.MetaHeaderHandler, unFinished map[string]data.MetaHeaderHandler, ctx context.Context) error {
	if ep.SyncPendingMiniBlocksFromMetaCalled != nil {
		return ep.SyncPendingMiniBlocksFromMetaCalled(epochStart, unFinished, ctx)
	}
	return nil
}

// GetMiniBlocks -
func (ep *EpochStartPendingMiniBlocksSyncHandlerMock) GetMiniBlocks() (map[string]*block.MiniBlock, error) {
	if ep.GetMiniBlocksCalled != nil {
		return ep.GetMiniBlocksCalled()
	}
	return nil, nil
}

// IsInterfaceNil -
func (ep *EpochStartPendingMiniBlocksSyncHandlerMock) IsInterfaceNil() bool {
	return ep == nil
}
