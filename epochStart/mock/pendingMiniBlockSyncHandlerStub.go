package mock

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// PendingMiniBlockSyncHandlerStub -
type PendingMiniBlockSyncHandlerStub struct {
	SyncPendingMiniBlocksCalled func(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error
	GetMiniBlocksCalled         func() (map[string]*block.MiniBlock, error)
}

// SyncPendingMiniBlocks -
func (pm *PendingMiniBlockSyncHandlerStub) SyncPendingMiniBlocks(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error {
	if pm.SyncPendingMiniBlocksCalled != nil {
		return pm.SyncPendingMiniBlocksCalled(miniBlockHeaders, ctx)
	}
	return nil
}

// GetMiniBlocks -
func (pm *PendingMiniBlockSyncHandlerStub) GetMiniBlocks() (map[string]*block.MiniBlock, error) {
	if pm.GetMiniBlocksCalled != nil {
		return pm.GetMiniBlocksCalled()
	}
	return nil, nil
}

// ClearFields --
func (pm *PendingMiniBlockSyncHandlerStub) ClearFields() {

}

// IsInterfaceNil -
func (pm *PendingMiniBlockSyncHandlerStub) IsInterfaceNil() bool {
	return pm == nil
}
