package trie

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

// trieStorageManagerWithoutCheckpoints manages the storage operations of the trie, but does not create checkpoints
type trieStorageManagerWithoutCheckpoints struct {
	common.StorageManager
	storerWithStats common.StorageManagerWithStats
}

// NewTrieStorageManagerWithoutCheckpoints creates a new instance of trieStorageManagerWithoutCheckpoints
func NewTrieStorageManagerWithoutCheckpoints(tsm common.StorageManager) (*trieStorageManagerWithoutCheckpoints, error) {
	if check.IfNil(tsm) {
		return nil, ErrNilTrieStorage
	}

	tsmWithStats, ok := tsm.GetBaseTrieStorageManager().(common.StorageManagerWithStats)
	if !ok {
		return nil, fmt.Errorf("invalid storage manager type %T", tsm.GetBaseTrieStorageManager())
	}

	return &trieStorageManagerWithoutCheckpoints{
		StorageManager:  tsm,
		storerWithStats: tsmWithStats,
	}, nil
}

// SetCheckpoint does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutCheckpoints) SetCheckpoint(
	_ []byte,
	_ []byte,
	iteratorChannels *common.TrieIteratorChannels,
	_ chan []byte,
	stats common.SnapshotStatisticsHandler,
) {
	if iteratorChannels != nil {
		common.CloseKeyValueHolderChan(iteratorChannels.LeavesChan)
	}
	stats.SnapshotFinished()

	log.Debug("trieStorageManagerWithoutCheckpoints - SetCheckpoint is disabled")
}

// AddDirtyCheckpointHashes returns false
func (tsm *trieStorageManagerWithoutCheckpoints) AddDirtyCheckpointHashes(_ []byte, _ common.ModifiedHashes) bool {
	return false
}

// GetStatsCollector -
func (tsm *trieStorageManagerWithoutCheckpoints) GetStatsCollector() common.StateStatisticsHandler {
	return tsm.storerWithStats.GetStatsCollector()
}
