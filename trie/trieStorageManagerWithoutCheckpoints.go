package trie

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
)

// trieStorageManagerWithoutCheckpoints manages the storage operations of the trie, but does not create checkpoints
type trieStorageManagerWithoutCheckpoints struct {
	common.StorageManager
}

// NewTrieStorageManagerWithoutCheckpoints creates a new instance of trieStorageManagerWithoutCheckpoints
func NewTrieStorageManagerWithoutCheckpoints(tsm common.StorageManager) (*trieStorageManagerWithoutCheckpoints, error) {
	if check.IfNil(tsm) {
		return nil, ErrNilTrieStorage
	}

	return &trieStorageManagerWithoutCheckpoints{
		StorageManager: tsm,
	}, nil
}

// SetCheckpoint does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutCheckpoints) SetCheckpoint(
	_ []byte,
	_ []byte,
	chLeaves chan core.KeyValueHolder,
	_ chan error,
	stats common.SnapshotStatisticsHandler,
) {
	safelyCloseChan(chLeaves)
	stats.SnapshotFinished()

	log.Debug("trieStorageManagerWithoutCheckpoints - SetCheckpoint is disabled")
}

// AddDirtyCheckpointHashes returns false
func (tsm *trieStorageManagerWithoutCheckpoints) AddDirtyCheckpointHashes(_ []byte, _ common.ModifiedHashes) bool {
	return false
}
