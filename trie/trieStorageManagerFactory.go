package trie

import (
	"github.com/ElrondNetwork/elrond-go/common"
)

// CreateTrieStorageManager creates a new trie storage manager based on the given type
func CreateTrieStorageManager(
	args NewTrieStorageManagerArgs,
	pruningEnabled bool,
	snapshotsEnabled bool,
	checkpointsEnabled bool,
) (common.StorageManager, error) {
	log.Debug("trie pruning status", "enabled", pruningEnabled)
	if !pruningEnabled {
		return NewTrieStorageManagerWithoutPruning(args.MainStorer)
	}

	log.Debug("trie snapshot status", "enabled", snapshotsEnabled)
	if !snapshotsEnabled {
		return NewTrieStorageManagerWithoutSnapshot(args)
	}

	log.Debug("trie checkpoints status", "enabled", checkpointsEnabled)
	if !checkpointsEnabled {
		return NewTrieStorageManagerWithoutCheckpoints(args)
	}

	return NewTrieStorageManager(args)
}
