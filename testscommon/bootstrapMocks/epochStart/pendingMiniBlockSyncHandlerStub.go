package epochStart

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// PendingMiniBlockSyncHandlerStub -
type PendingMiniBlockSyncHandlerStub struct {
	SyncMiniBlocksCalled func(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error
	GetMiniBlocksCalled  func() (map[string]*block.MiniBlock, error)
}

// SyncMiniBlocks -
func (pm *PendingMiniBlockSyncHandlerStub) SyncMiniBlocks(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error {
	if pm.SyncMiniBlocksCalled != nil {
		return pm.SyncMiniBlocksCalled(miniBlockHeaders, ctx)
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
