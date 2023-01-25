package scheduledDataSyncer

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ScheduledSyncerStub -
type ScheduledSyncerStub struct {
	UpdateSyncDataIfNeededCalled func(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error)
	GetRootHashToSyncCalled      func(notarizedShardHeader data.ShardHeaderHandler) []byte
}

// UpdateSyncDataIfNeeded -
func (sdss *ScheduledSyncerStub) UpdateSyncDataIfNeeded(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error) {
	if sdss.UpdateSyncDataIfNeededCalled != nil {
		return sdss.UpdateSyncDataIfNeededCalled(notarizedShardHeader)
	}
	return nil, nil, nil
}

// GetRootHashToSync -
func (sdss *ScheduledSyncerStub) GetRootHashToSync(notarizedShardHeader data.ShardHeaderHandler) []byte {
	if sdss.GetRootHashToSyncCalled != nil {
		return sdss.GetRootHashToSyncCalled(notarizedShardHeader)
	}
	return []byte("rootHash to sync")
}

// IsInterfaceNil -
func (sdss *ScheduledSyncerStub) IsInterfaceNil() bool {
	return sdss == nil
}
