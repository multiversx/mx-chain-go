package trie

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder/disabled"
)

// trieStorageManagerWithoutCheckpoints manages the storage operations of the trie, but does not create checkpoints
type trieStorageManagerWithoutCheckpoints struct {
	*trieStorageManager
}

// NewTrieStorageManagerWithoutCheckpoints creates a new instance of trieStorageManagerWithoutCheckpoints
func NewTrieStorageManagerWithoutCheckpoints(args NewTrieStorageManagerArgs) (*trieStorageManagerWithoutCheckpoints, error) {
	args.CheckpointHashesHolder = disabled.NewDisabledCheckpointHashesHolder()
	tsm, err := NewTrieStorageManager(args)
	if err != nil {
		return nil, err
	}

	return &trieStorageManagerWithoutCheckpoints{tsm}, nil
}

// SetCheckpoint does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutCheckpoints) SetCheckpoint(_ []byte, chLeaves chan core.KeyValueHolder) {
	if chLeaves != nil {
		close(chLeaves)
	}

	log.Debug("trieStorageManagerWithoutCheckpoints - SetCheckpoint is disabled")
}

// AddDirtyCheckpointHashes returns false
func (tsm *trieStorageManagerWithoutCheckpoints) AddDirtyCheckpointHashes(_ []byte, _ temporary.ModifiedHashes) bool {
	return false
}

// Remove removes the given hash form the storage
func (tsm *trieStorageManagerWithoutCheckpoints) Remove(hash []byte) error {
	return tsm.db.Remove(hash)
}
